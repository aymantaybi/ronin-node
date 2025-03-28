// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package v1 implements the proof-of-authority consensus engine.
package v1

import (
	"bytes"
	"errors"
	"io"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hashicorp/golang-lru/arc/v2"
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	consortiumCommon "github.com/ethereum/go-ethereum/consensus/consortium/common"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
)

const (
	inmemorySnapshots  = 128  // Number of recent vote snapshots to keep in memory
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory

	wiggleTime = 1000 * time.Millisecond // Random delay (per signer) to allow concurrent signers
)

// Consortium proof-of-authority protocol constants.
var (
	epochLength = uint64(30000) // Default number of blocks after which to checkpoint

	extraVanity = 32 // Fixed number of extra-data prefix bytes reserved for signer vanity

	emptyNonce = hexutil.MustDecode("0x0000000000000000") // Nonce number should be empty

	uncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW

	diffInTurn = big.NewInt(7) // Block difficulty for in-turn signatures
	diffNoTurn = big.NewInt(3) // Block difficulty for out-of-turn signatures
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	// errInvalidNonce is returned if a nonce value is not 0x00..0
	errInvalidNonce = errors.New("nonce not 0x00..0 ")

	// errExtraSigners is returned if non-checkpoint block contain signer data in
	// their extra-data fields.
	errExtraSigners = errors.New("non-checkpoint block contains extra signer list")

	// ErrInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	errInvalidTimestamp = errors.New("invalid timestamp")

	// errUnauthorizedSigner is returned if a header is signed by a non-authorized entity.
	errUnauthorizedSigner = errors.New("unauthorized signer")

	// errWrongCoinbase is returned if the coinbase field in header does not match the signer
	// of that block.
	errWrongCoinbase = errors.New("wrong coinbase address")
)

// Consortium is the proof-of-authority consensus engine proposed to support the
// Ethereum testnet following the Ropsten attacks.
type Consortium struct {
	chainConfig *params.ChainConfig
	config      *params.ConsortiumConfig // Consensus engine configuration parameters
	db          ethdb.Database           // Database to store and retrieve snapshot checkpoints

	recents    *arc.ARCCache[common.Hash, *Snapshot]      // Snapshots for recent block to speed up reorgs
	signatures *arc.ARCCache[common.Hash, common.Address] // Signatures of recent blocks to speed up mining

	proposals map[common.Address]bool // Current list of proposals we are pushing

	val      common.Address // Ethereum address of the signing key
	signer   types.Signer
	signFn   consortiumCommon.SignerFn // Signer function to authorize hashes with
	signTxFn consortiumCommon.SignerTxFn

	lock sync.RWMutex // Protects the signer fields

	contract *consortiumCommon.ContractIntegrator
	ethAPI   *ethapi.PublicBlockChainAPI

	getSCValidators    func() ([]common.Address, error) // Get the list of validator from contract
	getFenixValidators func() ([]common.Address, error) // Get the validator list from Ronin Validator contract of Fenix hardfork

	skipCheckpointHeaderCheck bool
}

// New creates a Consortium proof-of-authority consensus engine with the initial
// signers set to the ones provided by the user.
func New(chainConfig *params.ChainConfig, db ethdb.Database, ethAPI *ethapi.PublicBlockChainAPI, skipCheckpointHeaderCheck bool) *Consortium {
	// Set any missing consensus parameters to their defaults
	consortiumConfig := *chainConfig.Consortium
	if consortiumConfig.Epoch == 0 {
		consortiumConfig.Epoch = epochLength
	}
	// Allocate the snapshot caches and create the engine
	recents, _ := arc.NewARC[common.Hash, *Snapshot](inmemorySnapshots)
	signatures, _ := arc.NewARC[common.Hash, common.Address](inmemorySignatures)

	consortium := Consortium{
		chainConfig:               chainConfig,
		config:                    &consortiumConfig,
		db:                        db,
		recents:                   recents,
		signatures:                signatures,
		ethAPI:                    ethAPI,
		proposals:                 make(map[common.Address]bool),
		signer:                    types.NewEIP155Signer(chainConfig.ChainID),
		skipCheckpointHeaderCheck: skipCheckpointHeaderCheck,
	}

	err := consortium.initContract(common.Address{}, nil)
	if err != nil {
		log.Error("Failed to init system contract caller", "err", err)
	}

	return &consortium
}

// SetGetSCValidatorsFn sets the function to get a list of validators from smart contracts
func (c *Consortium) SetGetSCValidatorsFn(fn func() ([]common.Address, error)) {
	c.getSCValidators = fn
}

// SetGetFenixValidators sets the function to get the validator list from Ronin Validator contract of Fenix hardfork
func (c *Consortium) SetGetFenixValidators(fn func() ([]common.Address, error)) {
	c.getFenixValidators = fn
}

// Author implements consensus.Engine, returning the Ethereum address recovered
// from the signature in the header's extra-data section.
func (c *Consortium) Author(header *types.Header) (common.Address, error) {
	return Ecrecover(header, c.signatures)
}

// VerifyBlobHeader only available in v2
func (c *Consortium) VerifyBlobHeader(block *types.Block, sidecars *[]*types.BlobTxSidecar) error {
	return nil
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *Consortium) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	return c.VerifyHeaderAndParents(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *Consortium) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := c.VerifyHeaderAndParents(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// VerifyHeaderAndParents checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (c *Consortium) VerifyHeaderAndParents(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return consortiumCommon.ErrUnknownBlock
	}
	number := header.Number.Uint64()

	// Don't waste time checking blocks from the future
	if header.Time > uint64(time.Now().Unix()) {
		return consensus.ErrFutureBlock
	}
	// Nonces must be 0x00..0
	if !bytes.Equal(header.Nonce[:], emptyNonce) {
		return errInvalidNonce
	}
	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		return consortiumCommon.ErrMissingVanity
	}
	if len(header.Extra) < extraVanity+consortiumCommon.ExtraSeal {
		return consortiumCommon.ErrMissingSignature
	}
	// Ensure that the extra-data contains a signer list on checkpoint, but none otherwise
	checkpoint := (number % c.config.Epoch) == 0
	signersBytes := len(header.Extra) - extraVanity - consortiumCommon.ExtraSeal
	if !checkpoint && signersBytes != 0 {
		return errExtraSigners
	}
	if checkpoint && signersBytes%common.AddressLength != 0 {
		return consortiumCommon.ErrInvalidCheckpointSigners
	}
	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != (common.Hash{}) {
		return consortiumCommon.ErrInvalidMixDigest
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in PoA
	if header.UncleHash != uncleHash {
		return consortiumCommon.ErrInvalidUncleHash
	}
	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if number > 0 {
		if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
			return consortiumCommon.ErrInvalidDifficulty
		}
	}
	// If all checks passed, validate any special fields for hard forks
	if err := misc.VerifyForkHashes(chain.Config(), header, false); err != nil {
		return err
	}
	// All basic checks passed, verify cascading fields
	return c.verifyCascadingFields(chain, header, parents)
}

// verifyCascadingFields verifies all the header fields that are not standalone,
// rather depend on a batch of previous headers. The caller may optionally pass
// in a batch of parents (ascending order) to avoid looking those up from the
// database. This is useful for concurrently verifying a batch of new headers.
func (c *Consortium) verifyCascadingFields(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}
	// Ensure that the block's timestamp isn't too close to its parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if parent.Time+c.config.Period > header.Time {
		return errInvalidTimestamp
	}

	// If the block is a checkpoint block, verify the signer list
	if number%c.config.Epoch != 0 {
		return c.verifySeal(chain, header, parents)
	}

	if !c.skipCheckpointHeaderCheck {
		signers, err := c.getValidatorsFromContract(chain, number-1)
		if err != nil {
			return err
		}

		extraSuffix := len(header.Extra) - consortiumCommon.ExtraSeal
		checkpointHeaders := consortiumCommon.ExtractAddressFromBytes(header.Extra[extraVanity:extraSuffix])
		validSigners := consortiumCommon.CompareSignersLists(checkpointHeaders, signers)
		if !validSigners {
			log.Error("signers lists are different in checkpoint header and snapshot", "number", number, "signersHeader", checkpointHeaders, "signers", signers)
			return consortiumCommon.ErrInvalidCheckpointSigners
		}
	}

	// All basic checks passed, verify the seal and return
	return c.verifySeal(chain, header, parents)
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (c *Consortium) snapshot(chain consensus.ChainHeaderReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	// Only the parents slice's length is modified so we only need to shallow copy
	// the slice here to make FindAncientHeader find its block ancestor
	cpyParents := parents
	for snap == nil {
		// If an in-memory snapshot was found, use that
		if s, ok := c.recents.Get(hash); ok {
			snap = s
			break
		}
		// If an on-disk checkpoint snapshot can be found, use that
		if number%c.config.Epoch == 0 {
			if s, err := loadSnapshot(c.config, c.signatures, c.db, hash); err == nil {
				log.Trace("Loaded snapshot from disk", "number", number, "hash", hash)
				snap = s
				break
			}
		}
		// If we're at the genesis, snapshot the initial state. Alternatively if we're
		// at a checkpoint block without a parent (light client CHT), or we have piled
		// up more headers than allowed to be reorged (chain reinit from a freezer),
		// consider the checkpoint trusted and snapshot it.
		if number == 0 || (number%c.config.Epoch == 0 && (len(headers) > params.FullImmutabilityThreshold || chain.GetHeaderByNumber(number-1) == nil)) {
			cpHeader := chain.GetHeaderByNumber(number)
			if cpHeader != nil {
				hash := cpHeader.Hash()

				validators, err := c.getValidatorsFromGenesis()
				if err != nil {
					return nil, err
				}
				snap = newSnapshot(c.config, c.signatures, number, hash, validators)
				if err := snap.store(c.db); err != nil {
					return nil, err
				}
				log.Info("Stored checkpoint snapshot to disk", "number", number, "hash", hash)
				break
			}
		}
		// No snapshot for this header, gather the header and move backward
		var header *types.Header
		if len(parents) > 0 {
			// If we have explicit parents, pick from there (enforced)
			header = parents[len(parents)-1]
			if header.Hash() != hash || header.Number.Uint64() != number {
				return nil, consensus.ErrUnknownAncestor
			}
			parents = parents[:len(parents)-1]
		} else {
			// No explicit parents (or no more left), reach out to the database
			header = chain.GetHeader(hash, number)
			if header == nil {
				return nil, consensus.ErrUnknownAncestor
			}
		}
		headers = append(headers, header)
		number, hash = number-1, header.ParentHash
	}
	// Previous snapshot found, apply any pending headers on top of it
	for i := 0; i < len(headers)/2; i++ {
		headers[i], headers[len(headers)-1-i] = headers[len(headers)-1-i], headers[i]
	}
	snap, err := snap.apply(chain, c, headers, cpyParents)
	if err != nil {
		return nil, err
	}
	c.recents.Add(snap.Hash, snap)

	// If we've generated a new checkpoint snapshot, save to disk
	if snap.Number%c.config.Epoch == 0 && len(headers) > 0 {
		if err = snap.store(c.db); err != nil {
			return nil, err
		}
		log.Info("Stored checkpoint snapshot to disk", "number", snap.Number, "hash", snap.Hash)
	}
	return snap, err
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (c *Consortium) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// VerifySeal implements consensus.Engine, checking whether the signature contained
// in the header satisfies the consensus protocol requirements.
func (c *Consortium) VerifySeal(chain consensus.ChainHeaderReader, header *types.Header) error {
	return c.verifySeal(chain, header, nil)
}

// verifySeal checks whether the signature contained in the header satisfies the
// consensus protocol requirements. The method accepts an optional list of parent
// headers that aren't yet part of the local blockchain to generate the snapshots
// from.
func (c *Consortium) verifySeal(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return consortiumCommon.ErrUnknownBlock
	}

	// Verifying the genesis block is not supported
	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := c.snapshot(chain, number-1, header.ParentHash, parents)
	if err != nil {
		return err
	}

	// Resolve the authorization key and check against signers
	signer, err := Ecrecover(header, c.signatures)
	if err != nil {
		return err
	}

	if signer != header.Coinbase {
		return errWrongCoinbase
	}

	//validators, err := c.getValidatorsFromLastCheckpoint(chain, number-1, nil)
	//if err != nil {
	//	return err
	//}

	validators := snap.SignerList
	// If we're amongst the recent signers, wait for the next block
	//for seen, recent := range snap.Recents {
	//	if recent == signer {
	//		// Signer is among recents, only wait if the current block doesn't shift it out
	//		if limit := uint64(len(validators)/2 + 1); seen > number-limit {
	//			return errors.New("signed recently, must wait for others")
	//		}
	//	}
	//}

	if _, ok := snap.SignerSet[signer]; !ok {
		return errUnauthorizedSigner
	}
	// Ensure that the difficulty corresponds to the turn-ness of the signer
	inturn := c.signerInTurn(signer, header.Number.Uint64(), validators)
	if inturn && header.Difficulty.Cmp(diffInTurn) != 0 {
		return consortiumCommon.ErrWrongDifficulty
	}
	if !inturn && header.Difficulty.Cmp(diffNoTurn) != 0 {
		return consortiumCommon.ErrWrongDifficulty
	}
	return nil
}

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (c *Consortium) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	// Set the Coinbase address as the signer
	header.Coinbase = c.val
	header.Nonce = types.BlockNonce{}

	number := header.Number.Uint64()
	validators, err := c.getValidatorsFromLastCheckpoint(chain, number-1, nil)
	if err != nil {
		return err
	}
	// Set the correct difficulty
	header.Difficulty = c.doCalcDifficulty(c.val, number, validators)

	// Ensure the extra data has all its components
	if len(header.Extra) < extraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity]

	if number%c.config.Epoch == 0 {
		validators, err := c.getValidatorsFromContract(chain, number)
		if err != nil {
			return err
		}

		for _, signer := range validators {
			header.Extra = append(header.Extra, signer[:]...)
		}
	}
	header.Extra = append(header.Extra, make([]byte, consortiumCommon.ExtraSeal)...)

	// Mix digest is reserved for now, set to empty
	header.MixDigest = common.Hash{}

	// Ensure the timestamp has the correct delay
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Time = parent.Time + c.config.Period
	if header.Time < uint64(time.Now().Unix()) {
		header.Time = uint64(time.Now().Unix())
	}
	return nil
}

// Finalize implements consensus.Engine, ensuring no uncles are set, nor block
// rewards given.
func (c *Consortium) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs *[]*types.Transaction,
	uncles []*types.Header, receipts *[]*types.Receipt, systemTxs *[]*types.Transaction, internalTxs *[]*types.InternalTransaction, usedGas *uint64) error {
	lastBlockInV1 := c.chainConfig.IsOnConsortiumV2(new(big.Int).Add(header.Number, common.Big1))
	if (len(*systemTxs) > 0 && !lastBlockInV1) || (len(*systemTxs) == 0 && lastBlockInV1) {
		return errors.New("the length of systemTxs does not match")
	}

	if len(*systemTxs) > 0 {
		log.Info("processing system tx from consortium v1", "systemTxs", len(*systemTxs), "coinbase", header.Coinbase.Hex())
		evmContext := core.NewEVMBlockContext(header, consortiumCommon.ChainContext{Chain: chain, Consortium: c}, &header.Coinbase, chain.OpEvents()...)
		transactOpts := &consortiumCommon.ApplyTransactOpts{
			ApplyMessageOpts: &consortiumCommon.ApplyMessageOpts{
				State:       state,
				Header:      header,
				ChainConfig: c.chainConfig,
				EVMContext:  &evmContext,
			},
			Txs:         txs,
			Receipts:    receipts,
			ReceivedTxs: systemTxs,
			UsedGas:     usedGas,
			Mining:      false,
			Signer:      c.signer,
			SignTxFn:    c.signTxFn,
			EthAPI:      c.ethAPI,
		}
		if err := c.contract.WrapUpEpoch(transactOpts); err != nil {
			return err
		}
		if len(*systemTxs) > 0 {
			return errors.New("the length of systemTxs does not match")
		}
	}

	// No block rewards in PoA, so the state remains as is and uncles are dropped
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = types.CalcUncleHash(nil)

	return nil
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
func (c *Consortium) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, []*types.Receipt, error) {
	// No block rewards in PoA, so the state remains as is and uncles are dropped
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = types.CalcUncleHash(nil)

	if c.chainConfig.IsOnConsortiumV2(big.NewInt(header.Number.Int64() + 1)) {
		evmContext := core.NewEVMBlockContext(header, consortiumCommon.ChainContext{Chain: chain, Consortium: c}, &header.Coinbase, chain.OpEvents()...)
		transactOpts := &consortiumCommon.ApplyTransactOpts{
			ApplyMessageOpts: &consortiumCommon.ApplyMessageOpts{
				State:       state,
				Header:      header,
				ChainConfig: c.chainConfig,
				EVMContext:  &evmContext,
			},
			Txs:         &txs,
			Receipts:    &receipts,
			ReceivedTxs: nil,
			UsedGas:     &header.GasUsed,
			Mining:      true,
			Signer:      c.signer,
			SignTxFn:    c.signTxFn,
			EthAPI:      c.ethAPI,
		}
		if err := c.contract.WrapUpEpoch(transactOpts); err != nil {
			log.Error("Failed to update validators", "err", err)
		}
		// should not happen. Once happen, stop the node is better than broadcast the block
		if header.GasLimit < header.GasUsed {
			return nil, nil, errors.New("gas consumption of system txs exceed the gas limit")
		}
		header.UncleHash = types.CalcUncleHash(nil)
		var blk *types.Block
		var rootHash common.Hash
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			rootHash = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
			wg.Done()
		}()
		go func() {
			blk = types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil))
			wg.Done()
		}()
		wg.Wait()
		blk.SetRoot(rootHash)
		// Assemble and return the final block for sealing
		return blk, receipts, nil
	}

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, new(trie.Trie)), receipts, nil
}

// Authorize injects a private key into the consensus engine to mint new blocks
// with.
func (c *Consortium) Authorize(signer common.Address, signFn consortiumCommon.SignerFn, signTxFn consortiumCommon.SignerTxFn) {
	c.lock.Lock()
	c.val = signer
	c.signFn = signFn
	c.signTxFn = signTxFn
	c.lock.Unlock()

	err := c.initContract(signer, signTxFn)
	if err != nil {
		log.Error("Failed to init system contract caller", "err", err)
	}
}

// Seal implements consensus.Engine, attempting to create a sealed block using
// the local signing credentials.
func (c *Consortium) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	header := block.Header()

	// Sealing the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return consortiumCommon.ErrUnknownBlock
	}
	// For 0-period chains, refuse to seal empty blocks (no reward but would spin sealing)
	if c.config.Period == 0 && len(block.Transactions()) == 0 {
		return errors.New("sealing paused while waiting for transactions")
	}
	// Don't hold the signer fields for the entire sealing procedure
	c.lock.RLock()
	signer, signFn := c.val, c.signFn
	c.lock.RUnlock()

	validators, err := c.getValidatorsFromLastCheckpoint(chain, number-1, nil)
	if err != nil {
		return err
	}
	if !consortiumCommon.SignerInList(c.val, validators) {
		return errUnauthorizedSigner
	}
	snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}
	// If we're amongst the recent signers, wait for the next block
	for seen, recent := range snap.Recents {
		if recent == signer {
			// Signer is among recents, only wait if the current block doesn't shift it out
			if limit := uint64(len(validators)/2 + 1); seen > number-limit {
				return consortiumCommon.ErrRecentlySigned
			}
		}
	}

	// Sweet, the protocol permits us to sign the block, wait for our time
	delay := time.Unix(int64(header.Time), 0).Sub(time.Now()) // nolint: gosimple
	if !c.signerInTurn(signer, number, validators) {
		// It's not our turn explicitly to sign, delay it a bit
		wiggle := time.Duration(len(validators)/2+1) * wiggleTime
		delay += time.Duration(rand.Int63n(int64(wiggle))) + wiggleTime // delay for 0.5s more

		log.Trace("Out-of-turn signing requested", "wiggle", common.PrettyDuration(wiggle))
	}
	// Sign all the things!
	sighash, err := signFn(accounts.Account{Address: signer}, accounts.MimetypeTextPlain, consortiumRLP(header))
	if err != nil {
		return err
	}
	copy(header.Extra[len(header.Extra)-consortiumCommon.ExtraSeal:], sighash)
	// Wait until sealing is terminated or delay timeout.
	log.Trace("Waiting for slot to sign and propagate", "delay", common.PrettyDuration(delay))
	go func() {
		select {
		case <-stop:
			return
		case <-time.After(delay):
		}

		select {
		case results <- block.WithSeal(header):
		default:
			log.Warn("Sealing result is not read by miner", "sealhash", SealHash(header))
		}
	}()

	return nil
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *Consortium) SealHash(header *types.Header) common.Hash {
	return SealHash(header)
}

// Close implements consensus.Engine. It's a noop for consortium as there are no background threads.
func (c *Consortium) Close() error {
	return nil
}

// APIs implements consensus.Engine, returning the user facing RPC API.
func (c *Consortium) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "consortium",
		Version:   "1.0",
		Service:   &API{chain: chain, consortium: c},
		Public:    false,
	}}
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have.
func (c *Consortium) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	number := parent.Number.Uint64() + 1
	validators, err := c.getValidatorsFromLastCheckpoint(chain, number-1, []*types.Header{parent})
	if err != nil {
		return nil
	}
	return c.doCalcDifficulty(c.val, number, validators)
}

func (c *Consortium) GetSnapshot(
	chain consensus.ChainHeaderReader,
	number uint64,
	hash common.Hash,
	parents []*types.Header,
) *consortiumCommon.BaseSnapshot {
	snap, err := c.snapshot(chain, number, hash, parents)
	if err != nil {
		return nil
	}
	return &consortiumCommon.BaseSnapshot{
		Number:     snap.Number,
		Hash:       snap.Hash,
		SignerSet:  snap.SignerSet,
		SignerList: snap.SignerList,
		Recents:    snap.Recents,
	}
}

func (c *Consortium) doCalcDifficulty(signer common.Address, number uint64, validators []common.Address) *big.Int {
	if c.signerInTurn(signer, number, validators) {
		return new(big.Int).Set(diffInTurn)
	}
	return new(big.Int).Set(diffNoTurn)
}

// getValidatorsFromGenesis gets the list of validators from the genesis block support backward compatibility in v1, only used with Snap Sync.
func (c *Consortium) getValidatorsFromGenesis() ([]common.Address, error) {
	var validatorSet []string
	switch {
	case c.chainConfig.ChainID.Cmp(big.NewInt(2020)) == 0:
		validatorSet = []string{
			"0x000000000000000000000000f224beff587362a88d859e899d0d80c080e1e812",
			"0x00000000000000000000000011360eacdedd59bc433afad4fc8f0417d1fbebab",
			"0x00000000000000000000000070bb1fb41c8c42f6ddd53a708e2b82209495e455",
		}
	case c.chainConfig.ChainID.Cmp(big.NewInt(2021)) == 0:
		validatorSet = []string{
			"0x0000000000000000000000004a4bc674a97737376cfe990ae2fe0d2b6e738393",
			"0x000000000000000000000000b6bc5bc0410773a3f86b1537ce7495c52e38f88b",
		}
	default:
		return nil, errors.New("no validator set for this chain only support Mainnet & Testnet")
	}
	var addresses []common.Address
	for _, str := range validatorSet {
		addresses = append(addresses, common.HexToAddress(str))
	}
	return addresses, nil
}

// Read the validator list from contract
func (c *Consortium) getValidatorsFromContract(chain consensus.ChainHeaderReader, number uint64) ([]common.Address, error) {
	if chain.Config().IsFenix(big.NewInt(int64(number))) {
		if c.getFenixValidators == nil {
			return nil, errors.New("No getFenixValidators function supplied")
		}
		return c.getFenixValidators()
	}

	if c.getSCValidators == nil {
		return nil, errors.New("No getSCValidators function supplied")
	}

	return c.getSCValidators()
}

// getValidatorsFromLastCheckpoint gets the list of validator in the Extra field in the last checkpoint
// Sometime, when syncing the database have not stored the recent headers yet, so we need to look them up by passing them directly
func (c *Consortium) getValidatorsFromLastCheckpoint(chain consensus.ChainHeaderReader, number uint64, recents []*types.Header) ([]common.Address, error) {
	lastCheckpoint := number / c.config.Epoch * c.config.Epoch

	if lastCheckpoint == 0 {
		// TODO(andy): Review if we should put validators in genesis block's extra data
		return c.getValidatorsFromGenesis()
	}

	var header *types.Header
	if recents != nil {
		for _, parent := range recents {
			if parent.Number.Uint64() == lastCheckpoint {
				header = parent
			}
		}
	}
	if header == nil {
		header = chain.GetHeaderByNumber(lastCheckpoint)
	}
	extraSuffix := len(header.Extra) - consortiumCommon.ExtraSeal
	return consortiumCommon.ExtractAddressFromBytes(header.Extra[extraVanity:extraSuffix]), nil
}

// Check if it is the turn of the signer from the last checkpoint
func (c *Consortium) signerInTurn(signer common.Address, number uint64, validators []common.Address) bool {
	lastCheckpoint := number / c.config.Epoch * c.config.Epoch
	index := (number - lastCheckpoint) % uint64(len(validators))
	return validators[index] == signer
}

func (c *Consortium) initContract(coinbase common.Address, signTxFn consortiumCommon.SignerTxFn) error {
	if c.chainConfig.ConsortiumV2Block != nil && c.chainConfig.ConsortiumV2Contracts != nil {
		contract, err := consortiumCommon.NewContractIntegrator(c.chainConfig, consortiumCommon.NewConsortiumBackend(c.ethAPI), signTxFn, coinbase, c.ethAPI)
		if err != nil {
			return err
		}
		c.contract = contract
	}
	return nil
}

// ecrecover extracts the Ethereum account address from a signed header.
func Ecrecover(header *types.Header, sigcache *arc.ARCCache[common.Hash, common.Address]) (common.Address, error) {
	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address, nil
	}
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < consortiumCommon.ExtraSeal {
		return common.Address{}, consortiumCommon.ErrMissingSignature
	}
	signature := header.Extra[len(header.Extra)-consortiumCommon.ExtraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(SealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	sigcache.Add(hash, signer)
	return signer, nil
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.Sum(hash[:0])
	return hash
}

// consortiumRLP returns the rlp bytes which needs to be signed for the proof-of-authority
// sealing. The RLP to sign consists of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func consortiumRLP(header *types.Header) []byte {
	b := new(bytes.Buffer)
	encodeSigHeader(b, header)
	return b.Bytes()
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	err := rlp.Encode(w, []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-crypto.SignatureLength], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	})
	if err != nil {
		panic("can't encode: " + err.Error())
	}
}
