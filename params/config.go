// Copyright 2016 The go-ethereum Authors
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

package params

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/sha3"
)

const (
	BLSSignatureLength = 96 // BLSSignatureLength defines the byte length of a BLSSignature.
	BLSSecretKeyLength = 32
	BLSPubkeyLength    = 48 // BLSPubkeyLength defines the byte length of a BLSSignature.
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash      = common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
	RopstenGenesisHash      = common.HexToHash("0x41941023680923e0fe4d74a34bdac8141f2540e3ae90623718e47d66d1ca4a2d")
	SepoliaGenesisHash      = common.HexToHash("0x25a5cc106eea7138acab33231d7160d69cb777ee0c2c553fcddf5138993e6dd9")
	RinkebyGenesisHash      = common.HexToHash("0x6341fd3daf94b748c72ced5a5b26028f2474f5f00d824504e4fa37a75767e177")
	GoerliGenesisHash       = common.HexToHash("0xbf7e331f7f7c1dd2e05159666b3bf8bc7a8a3a9eb1d518969eab529dd9b88c1a")
	RoninMainnetGenesisHash = common.HexToHash("0x6e675ee97607f4e695188786c3c1853fb1562f1c075629eb5dbcff269422a1a4")
	RoninTestnetGenesisHash = common.HexToHash("0x13e47595099383189b8b0d5f3b67aa161495e478bb3fea64f4cf85cdf69cac4d")
)

// TrustedCheckpoints associates each known checkpoint with the genesis hash of
// the chain it belongs to.
var TrustedCheckpoints = map[common.Hash]*TrustedCheckpoint{
	MainnetGenesisHash: MainnetTrustedCheckpoint,
	RopstenGenesisHash: RopstenTrustedCheckpoint,
	SepoliaGenesisHash: SepoliaTrustedCheckpoint,
	RinkebyGenesisHash: RinkebyTrustedCheckpoint,
	GoerliGenesisHash:  GoerliTrustedCheckpoint,
}

// CheckpointOracles associates each known checkpoint oracles with the genesis hash of
// the chain it belongs to.
var CheckpointOracles = map[common.Hash]*CheckpointOracleConfig{
	MainnetGenesisHash: MainnetCheckpointOracle,
	RopstenGenesisHash: RopstenCheckpointOracle,
	RinkebyGenesisHash: RinkebyCheckpointOracle,
	GoerliGenesisHash:  GoerliCheckpointOracle,
}

var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(1_150_000),
		DAOForkBlock:        big.NewInt(1_920_000),
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(2_463_000),
		EIP150Hash:          common.HexToHash("0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0"),
		EIP155Block:         big.NewInt(2_675_000),
		EIP158Block:         big.NewInt(2_675_000),
		ByzantiumBlock:      big.NewInt(4_370_000),
		ConstantinopleBlock: big.NewInt(7_280_000),
		PetersburgBlock:     big.NewInt(7_280_000),
		IstanbulBlock:       big.NewInt(9_069_000),
		MuirGlacierBlock:    big.NewInt(9_200_000),
		BerlinBlock:         big.NewInt(12_244_000),
		LondonBlock:         big.NewInt(12_965_000),
		ArrowGlacierBlock:   big.NewInt(13_773_000),
		Ethash:              new(EthashConfig),
	}

	// MainnetTrustedCheckpoint contains the light client trusted checkpoint for the main network.
	MainnetTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 413,
		SectionHead:  common.HexToHash("0x8aa8e64ceadcdc5f23bc41d2acb7295a261a5cf680bb00a34f0e01af08200083"),
		CHTRoot:      common.HexToHash("0x008af584d385a2610706c5a439d39f15ddd4b691c5d42603f65ae576f703f477"),
		BloomRoot:    common.HexToHash("0x5a081af71a588f4d90bced242545b08904ad4fb92f7effff2ceb6e50e6dec157"),
	}

	// MainnetCheckpointOracle contains a set of configs for the main network oracle.
	MainnetCheckpointOracle = &CheckpointOracleConfig{
		Address: common.HexToAddress("0x9a9070028361F7AAbeB3f2F2Dc07F82C4a98A02a"),
		Signers: []common.Address{
			common.HexToAddress("0x1b2C260efc720BE89101890E4Db589b44E950527"), // Peter
			common.HexToAddress("0x78d1aD571A1A09D60D9BBf25894b44e4C8859595"), // Martin
			common.HexToAddress("0x286834935f4A8Cfb4FF4C77D5770C2775aE2b0E7"), // Zsolt
			common.HexToAddress("0xb86e2B0Ab5A4B1373e40c51A7C712c70Ba2f9f8E"), // Gary
			common.HexToAddress("0x0DF8fa387C602AE62559cC4aFa4972A7045d6707"), // Guillaume
		},
		Threshold: 2,
	}

	// RopstenChainConfig contains the chain parameters to run a node on the Ropsten test network.
	RopstenChainConfig = &ChainConfig{
		ChainID:             big.NewInt(3),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.HexToHash("0x41941023680923e0fe4d74a34bdac8141f2540e3ae90623718e47d66d1ca4a2d"),
		EIP155Block:         big.NewInt(10),
		EIP158Block:         big.NewInt(10),
		ByzantiumBlock:      big.NewInt(1_700_000),
		ConstantinopleBlock: big.NewInt(4_230_000),
		PetersburgBlock:     big.NewInt(4_939_394),
		IstanbulBlock:       big.NewInt(6_485_846),
		MuirGlacierBlock:    big.NewInt(7_117_117),
		BerlinBlock:         big.NewInt(9_812_189),
		LondonBlock:         big.NewInt(10_499_401),
		Ethash:              new(EthashConfig),
	}

	// RopstenTrustedCheckpoint contains the light client trusted checkpoint for the Ropsten test network.
	RopstenTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 346,
		SectionHead:  common.HexToHash("0xafa0384ebd13a751fb7475aaa7fc08ac308925c8b2e2195bca2d4ab1878a7a84"),
		CHTRoot:      common.HexToHash("0x522ae1f334bfa36033b2315d0b9954052780700b69448ecea8d5877e0f7ee477"),
		BloomRoot:    common.HexToHash("0x4093fd53b0d2cc50181dca353fe66f03ae113e7cb65f869a4dfb5905de6a0493"),
	}

	// RopstenCheckpointOracle contains a set of configs for the Ropsten test network oracle.
	RopstenCheckpointOracle = &CheckpointOracleConfig{
		Address: common.HexToAddress("0xEF79475013f154E6A65b54cB2742867791bf0B84"),
		Signers: []common.Address{
			common.HexToAddress("0x32162F3581E88a5f62e8A61892B42C46E2c18f7b"), // Peter
			common.HexToAddress("0x78d1aD571A1A09D60D9BBf25894b44e4C8859595"), // Martin
			common.HexToAddress("0x286834935f4A8Cfb4FF4C77D5770C2775aE2b0E7"), // Zsolt
			common.HexToAddress("0xb86e2B0Ab5A4B1373e40c51A7C712c70Ba2f9f8E"), // Gary
			common.HexToAddress("0x0DF8fa387C602AE62559cC4aFa4972A7045d6707"), // Guillaume
		},
		Threshold: 2,
	}

	// SepoliaChainConfig contains the chain parameters to run a node on the Sepolia test network.
	SepoliaChainConfig = &ChainConfig{
		ChainID:             big.NewInt(11155111),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		Ethash:              new(EthashConfig),
	}

	// SepoliaTrustedCheckpoint contains the light client trusted checkpoint for the Sepolia test network.
	SepoliaTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 1,
		SectionHead:  common.HexToHash("0x5dde65e28745b10ff9e9b86499c3a3edc03587b27a06564a4342baf3a37de869"),
		CHTRoot:      common.HexToHash("0x042a0d914f7baa4f28f14d12291e5f346e88c5b9d95127bf5422a8afeacd27e8"),
		BloomRoot:    common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
	}

	// RinkebyChainConfig contains the chain parameters to run a node on the Rinkeby test network.
	RinkebyChainConfig = &ChainConfig{
		ChainID:             big.NewInt(4),
		HomesteadBlock:      big.NewInt(1),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(2),
		EIP150Hash:          common.HexToHash("0x9b095b36c15eaf13044373aef8ee0bd3a382a5abb92e402afa44b8249c3a90e9"),
		EIP155Block:         big.NewInt(3),
		EIP158Block:         big.NewInt(3),
		ByzantiumBlock:      big.NewInt(1_035_301),
		ConstantinopleBlock: big.NewInt(3_660_663),
		PetersburgBlock:     big.NewInt(4_321_234),
		IstanbulBlock:       big.NewInt(5_435_345),
		MuirGlacierBlock:    nil,
		BerlinBlock:         big.NewInt(8_290_928),
		LondonBlock:         big.NewInt(8_897_988),
		ArrowGlacierBlock:   nil,
		Clique: &CliqueConfig{
			Period: 15,
			Epoch:  30000,
		},
	}

	// RinkebyTrustedCheckpoint contains the light client trusted checkpoint for the Rinkeby test network.
	RinkebyTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 292,
		SectionHead:  common.HexToHash("0x4185c2f1bb85ecaa04409d1008ff0761092ea2e94e8a71d64b1a5abc37b81414"),
		CHTRoot:      common.HexToHash("0x03b0191e6140effe0b88bb7c97bfb794a275d3543cb3190662fb72d9beea423c"),
		BloomRoot:    common.HexToHash("0x3d5f6edccc87536dcbc0dd3aae97a318205c617dd3957b4261470c71481629e2"),
	}

	// RinkebyCheckpointOracle contains a set of configs for the Rinkeby test network oracle.
	RinkebyCheckpointOracle = &CheckpointOracleConfig{
		Address: common.HexToAddress("0xebe8eFA441B9302A0d7eaECc277c09d20D684540"),
		Signers: []common.Address{
			common.HexToAddress("0xd9c9cd5f6779558b6e0ed4e6acf6b1947e7fa1f3"), // Peter
			common.HexToAddress("0x78d1aD571A1A09D60D9BBf25894b44e4C8859595"), // Martin
			common.HexToAddress("0x286834935f4A8Cfb4FF4C77D5770C2775aE2b0E7"), // Zsolt
			common.HexToAddress("0xb86e2B0Ab5A4B1373e40c51A7C712c70Ba2f9f8E"), // Gary
		},
		Threshold: 2,
	}

	// GoerliChainConfig contains the chain parameters to run a node on the Görli test network.
	GoerliChainConfig = &ChainConfig{
		ChainID:             big.NewInt(5),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(1_561_651),
		MuirGlacierBlock:    nil,
		BerlinBlock:         big.NewInt(4_460_644),
		LondonBlock:         big.NewInt(5_062_605),
		ArrowGlacierBlock:   nil,
		Clique: &CliqueConfig{
			Period: 15,
			Epoch:  30000,
		},
	}

	RoninMainnetBlacklistContract                  = common.HexToAddress("0x313b24994c93FA0471CB4D7aB796b07467041806")
	RoninMainnetFenixValidatorContractAddress      = common.HexToAddress("0x7f13232Bdc3a010c3f749a1c25bF99f1C053CE70")
	RoninMainnetRoninValidatorSetAddress           = common.HexToAddress("0x617c5d73662282EA7FfD231E020eCa6D2B0D552f")
	RoninMainnetSlashIndicatorAddress              = common.HexToAddress("0xEBFFF2b32fA0dF9C5C8C5d5AAa7e8b51d5207bA3")
	RoninMainnetStakingContractAddress             = common.HexToAddress("0x545edb750eB8769C868429BE9586F5857A768758")
	RoninMainnetProfileContractAddress             = common.HexToAddress("0x840EBf1CA767CB690029E91856A357a43B85d035")
	RoninMainnetFinalityTrackingAddress            = common.HexToAddress("0xA30B2932CD8b8A89E34551Cdfa13810af38dA576")
	RoninMainnetWhiteListDeployerContractV2Address = common.HexToAddress("0xc1876d5C4BFAF0eE325E4226B2bdf216D9896AE1")
	RoninMainnetTreasuryAddress                    = common.HexToAddress("0xb903E3936d3ca90b69b29F1df2810083a2DC0d71")

	RoninMainnetChainConfig = &ChainConfig{
		ChainID:                            big.NewInt(2020),
		HomesteadBlock:                     big.NewInt(0),
		EIP150Block:                        big.NewInt(0),
		EIP155Block:                        big.NewInt(0),
		EIP158Block:                        big.NewInt(0),
		ByzantiumBlock:                     big.NewInt(0),
		ConstantinopleBlock:                big.NewInt(0),
		PetersburgBlock:                    big.NewInt(0),
		IstanbulBlock:                      big.NewInt(4977778),
		OdysseusBlock:                      big.NewInt(10301597),
		FenixBlock:                         big.NewInt(14938103),
		BlacklistContractAddress:           &RoninMainnetBlacklistContract,
		FenixValidatorContractAddress:      &RoninMainnetFenixValidatorContractAddress,
		WhiteListDeployerContractV2Address: &RoninMainnetWhiteListDeployerContractV2Address,
		Consortium: &ConsortiumConfig{
			Period:  3,
			Epoch:   600,
			EpochV2: 200,
		},
		ConsortiumV2Contracts: &ConsortiumV2Contracts{
			RoninValidatorSet: RoninMainnetRoninValidatorSetAddress,
			SlashIndicator:    RoninMainnetSlashIndicatorAddress,
			StakingContract:   RoninMainnetStakingContractAddress,
			ProfileContract:   RoninMainnetProfileContractAddress,
			FinalityTracking:  RoninMainnetFinalityTrackingAddress,
		},
		ConsortiumV2Block: big.NewInt(23155200),
		PuffyBlock:        big.NewInt(0),
		BubaBlock:         big.NewInt(0),
		OlekBlock:         big.NewInt(24935500),
		ShillinBlock:      big.NewInt(28825400),
		AntennaBlock:      big.NewInt(28825400),
		MikoBlock:         big.NewInt(32367400),
		RoninTrustedOrgUpgrade: &ContractUpgrade{
			ProxyAddress:          common.HexToAddress("0x98D0230884448B3E2f09a177433D60fb1E19C090"),
			ImplementationAddress: common.HexToAddress("0x59646258Ec25CC329f5ce93223e0A50ccfA3e885"),
		},
		LondonBlock:          big.NewInt(36052600),
		BerlinBlock:          big.NewInt(36052600),
		TrippBlock:           big.NewInt(36052600),
		TrippPeriod:          big.NewInt(19907),
		AaronBlock:           big.NewInt(36052600),
		ShanghaiBlock:        big.NewInt(43447600),
		CancunBlock:          big.NewInt(43447600),
		VenokiBlock:          big.NewInt(43447600),
		RoninTreasuryAddress: &RoninMainnetTreasuryAddress,
	}

	RoninTestnetBlacklistContract                  = common.HexToAddress("0xF53EED5210c9cF308abFe66bA7CF14884c95A8aC")
	RoninTestnetFenixValidatorContractAddress      = common.HexToAddress("0x1454cAAd1637b662432Bb795cD5773d21281eDAb")
	RoninTestnetRoninValidatorSetAddress           = common.HexToAddress("0x54B3AC74a90E64E8dDE60671b6fE8F8DDf18eC9d")
	RoninTestnetSlashIndicatorAddress              = common.HexToAddress("0xF7837778b6E180Df6696C8Fa986d62f8b6186752")
	RoninTestnetStakingContractAddress             = common.HexToAddress("0x9C245671791834daf3885533D24dce516B763B28")
	RoninTestnetProfileContractAddress             = common.HexToAddress("0x3b67c8D22a91572a6AB18acC9F70787Af04A4043")
	RoninTestnetFinalityTrackingAddress            = common.HexToAddress("0x41aCDFe786171824a037f2Cd6224c5916A58969a")
	RoninTestnetWhiteListDeployerContractV2Address = common.HexToAddress("0x50a7e07Aa75eB9C04281713224f50403cA79851F")
	RoninTestnetTreasuryAddress                    = common.HexToAddress("0x5cfca565c09cc32bb7ba7222a648f1b014d6c30b")
	RoninTestnetChainConfig                        = &ChainConfig{
		ChainID:                            big.NewInt(2021),
		HomesteadBlock:                     big.NewInt(0),
		EIP150Block:                        big.NewInt(0),
		EIP155Block:                        big.NewInt(0),
		EIP158Block:                        big.NewInt(0),
		ByzantiumBlock:                     big.NewInt(0),
		ConstantinopleBlock:                big.NewInt(0),
		PetersburgBlock:                    big.NewInt(0),
		IstanbulBlock:                      big.NewInt(0),
		OdysseusBlock:                      big.NewInt(3315095),
		FenixBlock:                         big.NewInt(6770400),
		BlacklistContractAddress:           &RoninTestnetBlacklistContract,
		FenixValidatorContractAddress:      &RoninTestnetFenixValidatorContractAddress,
		WhiteListDeployerContractV2Address: &RoninTestnetWhiteListDeployerContractV2Address,
		Consortium: &ConsortiumConfig{
			Period:  3,
			Epoch:   30,
			EpochV2: 200,
		},
		ConsortiumV2Contracts: &ConsortiumV2Contracts{
			RoninValidatorSet: RoninTestnetRoninValidatorSetAddress,
			SlashIndicator:    RoninTestnetSlashIndicatorAddress,
			StakingContract:   RoninTestnetStakingContractAddress,
			ProfileContract:   RoninTestnetProfileContractAddress,
			FinalityTracking:  RoninTestnetFinalityTrackingAddress,
		},
		ConsortiumV2Block: big.NewInt(11706000),
		PuffyBlock:        big.NewInt(12254000),
		BubaBlock:         big.NewInt(14260600),
		OlekBlock:         big.NewInt(16849000),
		ShillinBlock:      big.NewInt(20268000),
		AntennaBlock:      big.NewInt(20737258),
		MikoBlock:         big.NewInt(23694400),
		RoninTrustedOrgUpgrade: &ContractUpgrade{
			ProxyAddress:          common.HexToAddress("0x7507dc433a98E1fE105d69f19f3B40E4315A4F32"),
			ImplementationAddress: common.HexToAddress("0x6A51C2B073a6daDBeCAC1A420AFcA7788C81612f"),
		},
		LondonBlock:          big.NewInt(27580600),
		BerlinBlock:          big.NewInt(27580600),
		TrippBlock:           big.NewInt(27580600),
		TrippPeriod:          big.NewInt(19866),
		AaronBlock:           big.NewInt(28231200),
		ShanghaiBlock:        big.NewInt(35554400),
		CancunBlock:          big.NewInt(35554400),
		VenokiBlock:          big.NewInt(35554400),
		RoninTreasuryAddress: &RoninTestnetTreasuryAddress,
	}

	// GoerliTrustedCheckpoint contains the light client trusted checkpoint for the Görli test network.
	GoerliTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 176,
		SectionHead:  common.HexToHash("0x2de018858528434f93adb40b1f03f2304a86d31b4ef2b1f930da0134f5c32427"),
		CHTRoot:      common.HexToHash("0x8c17e497d38088321c147abe4acbdfb3c0cab7d7a2b97e07404540f04d12747e"),
		BloomRoot:    common.HexToHash("0x02a41b6606bd3f741bd6ae88792d75b1ad8cf0ea5e28fbaa03bc8b95cbd20034"),
	}

	// GoerliCheckpointOracle contains a set of configs for the Goerli test network oracle.
	GoerliCheckpointOracle = &CheckpointOracleConfig{
		Address: common.HexToAddress("0x18CA0E045F0D772a851BC7e48357Bcaab0a0795D"),
		Signers: []common.Address{
			common.HexToAddress("0x4769bcaD07e3b938B7f43EB7D278Bc7Cb9efFb38"), // Peter
			common.HexToAddress("0x78d1aD571A1A09D60D9BBf25894b44e4C8859595"), // Martin
			common.HexToAddress("0x286834935f4A8Cfb4FF4C77D5770C2775aE2b0E7"), // Zsolt
			common.HexToAddress("0xb86e2B0Ab5A4B1373e40c51A7C712c70Ba2f9f8E"), // Gary
			common.HexToAddress("0x0DF8fa387C602AE62559cC4aFa4972A7045d6707"), // Guillaume
		},
		Threshold: 2,
	}

	// AllEthashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Ethash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllEthashProtocolChanges = &ChainConfig{
		ChainID:                       big.NewInt(1337),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP150Hash:                    common.Hash{},
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		MikoBlock:                     big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             nil,
		OdysseusBlock:                 nil,
		FenixBlock:                    nil,
		ConsortiumV2Block:             nil,
		PuffyBlock:                    nil,
		BlacklistContractAddress:      nil,
		FenixValidatorContractAddress: nil,
		TerminalTotalDifficulty:       nil,
		Ethash:                        new(EthashConfig),
		Clique:                        nil,
		Consortium:                    nil,
		ConsortiumV2Contracts:         nil,
	}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{
		ChainID:                       big.NewInt(1337),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP150Hash:                    common.Hash{},
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		MikoBlock:                     big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             nil,
		OdysseusBlock:                 nil,
		FenixBlock:                    nil,
		ConsortiumV2Block:             nil,
		PuffyBlock:                    nil,
		BlacklistContractAddress:      nil,
		FenixValidatorContractAddress: nil,
		TerminalTotalDifficulty:       nil,
		Ethash:                        nil,
		Clique:                        &CliqueConfig{Period: 0, Epoch: 30000},
		Consortium:                    nil,
		ConsortiumV2Contracts:         nil,
	}

	TestChainConfig = &ChainConfig{
		ChainID:                       big.NewInt(1),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP150Hash:                    common.Hash{},
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		MikoBlock:                     big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		CancunBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             nil,
		OdysseusBlock:                 nil,
		FenixBlock:                    nil,
		ConsortiumV2Block:             nil,
		PuffyBlock:                    nil,
		BlacklistContractAddress:      nil,
		FenixValidatorContractAddress: nil,
		TerminalTotalDifficulty:       nil,
		Ethash:                        new(EthashConfig),
		Clique:                        nil,
		Consortium:                    nil,
		ConsortiumV2Contracts:         nil,
		RoninTreasuryAddress:          &common.Address{},
	}
	NonActivatedConfig = &ChainConfig{
		ChainID:                       nil,
		HomesteadBlock:                nil,
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   nil,
		EIP150Hash:                    common.Hash{},
		EIP155Block:                   nil,
		EIP158Block:                   nil,
		ByzantiumBlock:                nil,
		ConstantinopleBlock:           nil,
		PetersburgBlock:               nil,
		IstanbulBlock:                 nil,
		MuirGlacierBlock:              nil,
		MikoBlock:                     nil,
		BerlinBlock:                   nil,
		LondonBlock:                   nil,
		ArrowGlacierBlock:             nil,
		OdysseusBlock:                 nil,
		FenixBlock:                    nil,
		ConsortiumV2Block:             nil,
		PuffyBlock:                    nil,
		BlacklistContractAddress:      nil,
		FenixValidatorContractAddress: nil,
		TerminalTotalDifficulty:       nil,
		Ethash:                        new(EthashConfig),
		Clique:                        nil,
		Consortium:                    nil,
		ConsortiumV2Contracts:         nil,
	}
	TestRules = TestChainConfig.Rules(new(big.Int))
)

// TrustedCheckpoint represents a set of post-processed trie roots (CHT and
// BloomTrie) associated with the appropriate section index and head hash. It is
// used to start light syncing from this checkpoint and avoid downloading the
// entire header chain while still being able to securely access old headers/logs.
type TrustedCheckpoint struct {
	SectionIndex uint64      `json:"sectionIndex"`
	SectionHead  common.Hash `json:"sectionHead"`
	CHTRoot      common.Hash `json:"chtRoot"`
	BloomRoot    common.Hash `json:"bloomRoot"`
}

// HashEqual returns an indicator comparing the itself hash with given one.
func (c *TrustedCheckpoint) HashEqual(hash common.Hash) bool {
	if c.Empty() {
		return hash == common.Hash{}
	}
	return c.Hash() == hash
}

// Hash returns the hash of checkpoint's four key fields(index, sectionHead, chtRoot and bloomTrieRoot).
func (c *TrustedCheckpoint) Hash() common.Hash {
	var sectionIndex [8]byte
	binary.BigEndian.PutUint64(sectionIndex[:], c.SectionIndex)

	w := sha3.NewLegacyKeccak256()
	w.Write(sectionIndex[:])
	w.Write(c.SectionHead[:])
	w.Write(c.CHTRoot[:])
	w.Write(c.BloomRoot[:])

	var h common.Hash
	w.Sum(h[:0])
	return h
}

// Empty returns an indicator whether the checkpoint is regarded as empty.
func (c *TrustedCheckpoint) Empty() bool {
	return c.SectionHead == (common.Hash{}) || c.CHTRoot == (common.Hash{}) || c.BloomRoot == (common.Hash{})
}

// CheckpointOracleConfig represents a set of checkpoint contract(which acts as an oracle)
// config which used for light client checkpoint syncing.
type CheckpointOracleConfig struct {
	Address   common.Address   `json:"address"`
	Signers   []common.Address `json:"signers"`
	Threshold uint64           `json:"threshold"`
}

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block *big.Int    `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP150Hash  common.Hash `json:"eip150Hash,omitempty"`  // EIP150 HF hash (needed for header only clients as only gas pricing changed)

	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	PetersburgBlock     *big.Int `json:"petersburgBlock,omitempty"`     // Petersburg switch block (nil = same as Constantinople)
	IstanbulBlock       *big.Int `json:"istanbulBlock,omitempty"`       // Istanbul switch block (nil = no fork, 0 = already on istanbul)
	MuirGlacierBlock    *big.Int `json:"muirGlacierBlock,omitempty"`    // Eip-2384 (bomb delay) switch block (nil = no fork, 0 = already activated)
	BerlinBlock         *big.Int `json:"berlinBlock,omitempty"`         // Berlin switch block (nil = no fork, 0 = already on berlin)
	LondonBlock         *big.Int `json:"londonBlock,omitempty"`         // London switch block (nil = no fork, 0 = already on london)
	ArrowGlacierBlock   *big.Int `json:"arrowGlacierBlock,omitempty"`   // Eip-4345 (bomb delay) switch block (nil = no fork, 0 = already activated)
	OdysseusBlock       *big.Int `json:"odysseusBlock,omitempty"`       // Odysseus switch block (nil = no fork, 0 = already on activated)
	FenixBlock          *big.Int `json:"fenixBlock,omitempty"`          // Fenix switch block (nil = no fork, 0 = already on activated)
	ConsortiumV2Block   *big.Int `json:"consortiumV2Block,omitempty"`   // Consortium v2 switch block (nil = no fork, 0 = already on activated)

	// Puffy hardfork fix the wrong order in system transactions included in a block
	PuffyBlock *big.Int `json:"puffyBlock,omitempty"` // Puffy switch block (nil = no fork, 0 = already on activated)
	BubaBlock  *big.Int `json:"bubaBlock,omitempty"`  // Buba switch block (nil = no fork, 0 = already on activated)
	// Olek hardfork reduces the delay in block time of out of turn miner
	OlekBlock *big.Int `json:"olekBlock,omitempty"` // Olek switch block (nil = no fork, 0 = already on activated)
	// Shillin hardfork introduces fast finality
	ShillinBlock *big.Int `json:"shillinBlock,omitempty"` // Shillin switch block (nil = no fork, 0 = already on activated)

	AntennaBlock *big.Int `json:"antennaBlock,omitempty"` // AntennaBlock switch block (nil = no fork, 0 = already on activated)
	// Miko hardfork introduces sponsored transactions
	MikoBlock     *big.Int `json:"mikoBlock,omitempty"`     // Miko switch block (nil = no fork, 0 = already on activated)
	TrippBlock    *big.Int `json:"trippBlock,omitempty"`    // Tripp switch block (nil = no fork, 0 = already on activated)
	TrippPeriod   *big.Int `json:"trippPeriod,omitempty"`   // The period number at Tripp fork block.
	AaronBlock    *big.Int `json:"aaronBlock,omitempty"`    // Aaron switch block (nil = no fork, 0 = already on activated)
	ShanghaiBlock *big.Int `json:"shanghaiBlock,omitempty"` // Shanghai switch block (nil = no fork, 0 = already on activated)
	CancunBlock   *big.Int `json:"cancunBlock,omitempty"`   // Cancun switch block (nil = no fork, 0 = already on activated)
	VenokiBlock   *big.Int `json:"venokiBlock,omitempty"`   // Venoki switch block (nil = no fork, 0 = already on activated)
	PragueBlock   *big.Int `json:"pragueBlock,omitempty"`   // Prague switch block (nil = no fork, 0 = already on activated)

	BlacklistContractAddress           *common.Address `json:"blacklistContractAddress,omitempty"`           // Address of Blacklist Contract (nil = no blacklist)
	FenixValidatorContractAddress      *common.Address `json:"fenixValidatorContractAddress,omitempty"`      // Address of Ronin Contract in the Fenix hardfork (nil = no blacklist)
	WhiteListDeployerContractV2Address *common.Address `json:"whiteListDeployerContractV2Address,omitempty"` // Address of Whitelist Ronin Contract V2 (nil = no blacklist)
	RoninTreasuryAddress               *common.Address `json:"roninTreasuryAddress,omitempty"`

	// TerminalTotalDifficulty is the amount of total difficulty reached by
	// the network that triggers the consensus upgrade.
	TerminalTotalDifficulty *big.Int `json:"terminalTotalDifficulty,omitempty"`

	// Various consensus engines
	Ethash                      *EthashConfig          `json:"ethash,omitempty"`
	Clique                      *CliqueConfig          `json:"clique,omitempty"`
	Consortium                  *ConsortiumConfig      `json:"consortium,omitempty"`
	ConsortiumV2Contracts       *ConsortiumV2Contracts `json:"consortiumV2Contracts"`
	RoninTrustedOrgUpgrade      *ContractUpgrade       `json:"roninTrustedOrgUpgrade"`
	TransparentProxyCodeUpgrade *ContractCodeUpgrade   `json:"transparentProxyCodeUpgrade"`
}

type ContractUpgrade struct {
	ProxyAddress          common.Address `json:"proxyAddress"`
	ImplementationAddress common.Address `json:"implementationAddress"`
}

type ContractCodeUpgrade struct {
	AxieAddress common.Address `json:"axieAddress"`
	LandAddress common.Address `json:"landAddress"`
	Code        hexutil.Bytes  `json:"code"`
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *EthashConfig) String() string {
	return "ethash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// ConsortiumConfig is the consensus engine configs for proof-of-authority based sealing.
type ConsortiumConfig struct {
	Period  uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch   uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
	EpochV2 uint64 `json:"epochV2"`
}

// String implements the stringer interface, returning the consensus engine details.
func (c *ConsortiumConfig) String() string {
	return "consortium"
}

type ConsortiumV2Contracts struct {
	StakingContract   common.Address `json:"stakingContract"`
	RoninValidatorSet common.Address `json:"roninValidatorSet"`
	SlashIndicator    common.Address `json:"slashIndicator"`
	ProfileContract   common.Address `json:"profileContract"`
	FinalityTracking  common.Address `json:"finalityTracking"`
}

func (c *ConsortiumV2Contracts) IsSystemContract(address common.Address) bool {
	e := reflect.ValueOf(c).Elem()
	for i := 0; i < e.NumField(); i++ {
		if e.Field(i).Interface().(common.Address) == address {
			return true
		}
	}

	return false
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.Ethash != nil:
		engine = c.Ethash
	case c.Clique != nil:
		engine = c.Clique
	case c.Consortium != nil:
		engine = c.Consortium
	default:
		engine = "unknown"
	}
	roninValidatorSetSC := common.HexToAddress("")
	if c.ConsortiumV2Contracts != nil {
		roninValidatorSetSC = c.ConsortiumV2Contracts.RoninValidatorSet
	}

	slashIndicatorSC := common.HexToAddress("")
	if c.ConsortiumV2Contracts != nil {
		slashIndicatorSC = c.ConsortiumV2Contracts.SlashIndicator
	}

	stakingContract := common.HexToAddress("")
	if c.ConsortiumV2Contracts != nil {
		stakingContract = c.ConsortiumV2Contracts.StakingContract
	}

	profileContract := common.HexToAddress("")
	if c.ConsortiumV2Contracts != nil {
		profileContract = c.ConsortiumV2Contracts.ProfileContract
	}

	finalityTrackingContract := common.HexToAddress("")
	if c.ConsortiumV2Contracts != nil {
		finalityTrackingContract = c.ConsortiumV2Contracts.FinalityTracking
	}

	whiteListDeployerContractV2Address := common.HexToAddress("")
	if c.WhiteListDeployerContractV2Address != nil {
		whiteListDeployerContractV2Address = *c.WhiteListDeployerContractV2Address
	}

	roninTreasuryAddress := common.HexToAddress("")
	if c.RoninTreasuryAddress != nil {
		roninTreasuryAddress = *c.RoninTreasuryAddress
	}

	chainConfigFmt := "{ChainID: %v, Homestead: %v, DAO: %v, DAOSupport: %v, EIP150: %v, EIP155: %v, EIP158: %v, Byzantium: %v, Constantinople: %v, "
	chainConfigFmt += "Petersburg: %v, Istanbul: %v, Odysseus: %v, Fenix: %v, Muir Glacier: %v, Berlin: %v, London: %v, Arrow Glacier: %v, "
	chainConfigFmt += "Engine: %v, Blacklist Contract: %v, Fenix Validator Contract: %v, ConsortiumV2: %v, ConsortiumV2.RoninValidatorSet: %v, "
	chainConfigFmt += "ConsortiumV2.SlashIndicator: %v, ConsortiumV2.StakingContract: %v, Puffy: %v, Buba: %v, Olek: %v, Shillin: %v, Antenna: %v, "
	chainConfigFmt += "ConsortiumV2.ProfileContract: %v, ConsortiumV2.FinalityTracking: %v, whiteListDeployerContractV2Address: %v, roninTreasuryAddress: %v, "
	chainConfigFmt += "Miko: %v, Tripp: %v, TrippPeriod: %v, Aaron: %v, Shanghai: %v, Cancun: %v, Venoki: %v, Prague: %v}"

	return fmt.Sprintf(chainConfigFmt,
		c.ChainID,
		c.HomesteadBlock,
		c.DAOForkBlock,
		c.DAOForkSupport,
		c.EIP150Block,
		c.EIP155Block,
		c.EIP158Block,
		c.ByzantiumBlock,
		c.ConstantinopleBlock,
		c.PetersburgBlock,
		c.IstanbulBlock,
		c.OdysseusBlock,
		c.FenixBlock,
		c.MuirGlacierBlock,
		c.BerlinBlock,
		c.LondonBlock,
		c.ArrowGlacierBlock,
		engine,
		c.BlacklistContractAddress,
		c.FenixValidatorContractAddress,
		c.ConsortiumV2Block,
		roninValidatorSetSC.Hex(),
		slashIndicatorSC.Hex(),
		stakingContract.Hex(),
		c.PuffyBlock,
		c.BubaBlock,
		c.OlekBlock,
		c.ShillinBlock,
		c.AntennaBlock,
		profileContract.Hex(),
		finalityTrackingContract.Hex(),
		whiteListDeployerContractV2Address.Hex(),
		roninTreasuryAddress.Hex(),
		c.MikoBlock,
		c.TrippBlock,
		c.TrippPeriod,
		c.AaronBlock,
		c.ShanghaiBlock,
		c.CancunBlock,
		c.VenokiBlock,
		c.PragueBlock,
	)
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsHomestead(num *big.Int) bool {
	return isForked(c.HomesteadBlock, num)
}

// IsDAOFork returns whether num is either equal to the DAO fork block or greater.
func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
	return isForked(c.DAOForkBlock, num)
}

// IsEIP150 returns whether num is either equal to the EIP150 fork block or greater.
func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return isForked(c.EIP150Block, num)
}

// IsEIP155 returns whether num is either equal to the EIP155 fork block or greater.
func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return isForked(c.EIP155Block, num)
}

// IsEIP158 returns whether num is either equal to the EIP158 fork block or greater.
func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return isForked(c.EIP158Block, num)
}

// IsByzantium returns whether num is either equal to the Byzantium fork block or greater.
func (c *ChainConfig) IsByzantium(num *big.Int) bool {
	return isForked(c.ByzantiumBlock, num)
}

// IsConstantinople returns whether num is either equal to the Constantinople fork block or greater.
func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
	return isForked(c.ConstantinopleBlock, num)
}

// IsMuirGlacier returns whether num is either equal to the Muir Glacier (EIP-2384) fork block or greater.
func (c *ChainConfig) IsMuirGlacier(num *big.Int) bool {
	return isForked(c.MuirGlacierBlock, num)
}

// IsPetersburg returns whether num is either
// - equal to or greater than the PetersburgBlock fork block,
// - OR is nil, and Constantinople is active
func (c *ChainConfig) IsPetersburg(num *big.Int) bool {
	return isForked(c.PetersburgBlock, num) || c.PetersburgBlock == nil && isForked(c.ConstantinopleBlock, num)
}

// IsIstanbul returns whether num is either equal to the Istanbul fork block or greater.
func (c *ChainConfig) IsIstanbul(num *big.Int) bool {
	return isForked(c.IstanbulBlock, num)
}

// IsBerlin returns whether num is either equal to the Berlin fork block or greater.
func (c *ChainConfig) IsBerlin(num *big.Int) bool {
	return isForked(c.BerlinBlock, num)
}

// IsLondon returns whether num is either equal to the London fork block or greater.
func (c *ChainConfig) IsLondon(num *big.Int) bool {
	return isForked(c.LondonBlock, num)
}

// IsArrowGlacier returns whether num is either equal to the Arrow Glacier (EIP-4345) fork block or greater.
func (c *ChainConfig) IsArrowGlacier(num *big.Int) bool {
	return isForked(c.ArrowGlacierBlock, num)
}

// IsTerminalPoWBlock returns whether the given block is the last block of PoW stage.
func (c *ChainConfig) IsTerminalPoWBlock(parentTotalDiff *big.Int, totalDiff *big.Int) bool {
	if c.TerminalTotalDifficulty == nil {
		return false
	}
	return parentTotalDiff.Cmp(c.TerminalTotalDifficulty) < 0 && totalDiff.Cmp(c.TerminalTotalDifficulty) >= 0
}

// IsOdysseus returns whether the num is equals to or larger than the Odysseus fork block.
func (c *ChainConfig) IsOdysseus(num *big.Int) bool {
	return isForked(c.OdysseusBlock, num)
}

// IsFenix returns whether the num is equals to or larger than the Fenix fork block.
func (c *ChainConfig) IsFenix(num *big.Int) bool {
	return isForked(c.FenixBlock, num)
}

// IsLastConsortiumV1Block return if num is the last block in Consortium v1
func (c *ChainConfig) IsLastConsortiumV1Block(num *big.Int) bool {
	if c.ConsortiumV2Block != nil && num != nil {
		// ConsortiumV2Block must be >= 1 so no overflow check here
		return new(big.Int).Sub(c.ConsortiumV2Block, common.Big1).Cmp(num) == 0
	}
	return false
}

// IsConsortiumV2 returns whether the num is equals to or larger than the consortiumV2 fork block.
func (c *ChainConfig) IsConsortiumV2(num *big.Int) bool {
	return isForked(c.ConsortiumV2Block, num)
}

// IsOnConsortiumV2 returns whether the num is equals to the consortiumV2 fork block.
func (c *ChainConfig) IsOnConsortiumV2(num *big.Int) bool {
	return configNumEqual(c.ConsortiumV2Block, num)
}

// IsPuffy returns whether the num is equals to or larger than the puffy fork block.
func (c *ChainConfig) IsPuffy(num *big.Int) bool {
	return isForked(c.PuffyBlock, num)
}

// IsBuba returns whether the num is equals to or larger than the buba fork block.
func (c *ChainConfig) IsBuba(num *big.Int) bool {
	return isForked(c.BubaBlock, num)
}

// IsOlek returns whether the num is equals to or larger than the olek fork block.
func (c *ChainConfig) IsOlek(num *big.Int) bool {
	return isForked(c.OlekBlock, num)
}

// IsAntenna returns whether the num is equals to or larger than the Antenna fork block.
func (c *ChainConfig) IsAntenna(num *big.Int) bool {
	return isForked(c.AntennaBlock, num)
}

// IsShillin returns whether the num is equals to or larger than the shillin fork block.
func (c *ChainConfig) IsShillin(num *big.Int) bool {
	return isForked(c.ShillinBlock, num)
}

// IsMiko returns whether the num is equals to or larger than the miko fork block.
func (c *ChainConfig) IsMiko(num *big.Int) bool {
	return isForked(c.MikoBlock, num)
}

// IsTripp returns whether the num is equals to or larger than the tripp fork block.
func (c *ChainConfig) IsTripp(num *big.Int) bool {
	return isForked(c.TrippBlock, num)
}

// IsAaron returns whether the num is equals to or larger than the aaron fork block.
func (c *ChainConfig) IsAaron(num *big.Int) bool {
	return isForked(c.AaronBlock, num)
}

// IsShanghai returns whether the num is equals to or larger than the shanghai fork block.
func (c *ChainConfig) IsShanghai(num *big.Int) bool {
	return isForked(c.ShanghaiBlock, num)
}

// IsCancun returns whether the num is equals to or larger than the cancun fork block.
func (c *ChainConfig) IsCancun(num *big.Int) bool {
	return isForked(c.CancunBlock, num)
}

// IsVenoki returns whether the num is equals to or larger than the venoki fork block.
func (c *ChainConfig) IsVenoki(num *big.Int) bool {
	return isForked(c.VenokiBlock, num)
}

// IsPrague returns whether the num is equals to or larger than the prague fork block.
func (c *ChainConfig) IsPrague(num *big.Int) bool {
	return isForked(c.PragueBlock, num)
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

// CheckConfigForkOrder checks that we don't "skip" any forks, geth isn't pluggable enough
// to guarantee that forks can be implemented in a different order than on official networks
func (c *ChainConfig) CheckConfigForkOrder() error {
	type fork struct {
		name     string
		block    *big.Int
		optional bool // if true, the fork may be nil and next fork is still allowed
	}
	var lastFork fork
	for _, cur := range []fork{
		{name: "homesteadBlock", block: c.HomesteadBlock},
		{name: "daoForkBlock", block: c.DAOForkBlock, optional: true},
		{name: "eip150Block", block: c.EIP150Block},
		{name: "eip155Block", block: c.EIP155Block},
		{name: "eip158Block", block: c.EIP158Block},
		{name: "byzantiumBlock", block: c.ByzantiumBlock},
		{name: "constantinopleBlock", block: c.ConstantinopleBlock},
		{name: "petersburgBlock", block: c.PetersburgBlock},
		{name: "istanbulBlock", block: c.IstanbulBlock},
		{name: "muirGlacierBlock", block: c.MuirGlacierBlock, optional: true},
		{name: "berlinBlock", block: c.BerlinBlock},
		{name: "londonBlock", block: c.LondonBlock},
		{name: "arrowGlacierBlock", block: c.ArrowGlacierBlock, optional: true},
	} {
		if lastFork.name != "" {
			// Next one must be higher number
			if lastFork.block == nil && cur.block != nil {
				return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at %v",
					lastFork.name, cur.name, cur.block)
			}
			if lastFork.block != nil && cur.block != nil {
				if lastFork.block.Cmp(cur.block) > 0 {
					return fmt.Errorf("unsupported fork ordering: %v enabled at %v, but %v enabled at %v",
						lastFork.name, lastFork.block, cur.name, cur.block)
				}
			}
		}
		// If it was optional and not set, then ignore it
		if !cur.optional || cur.block != nil {
			lastFork = cur
		}
	}
	return nil
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int) *ConfigCompatError {
	if isForkIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, head) {
		return newCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if isForkIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, head) {
		return newCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if c.IsDAOFork(head) && c.DAOForkSupport != newcfg.DAOForkSupport {
		return newCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if isForkIncompatible(c.EIP150Block, newcfg.EIP150Block, head) {
		return newCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	}
	if isForkIncompatible(c.EIP155Block, newcfg.EIP155Block, head) {
		return newCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	}
	if isForkIncompatible(c.EIP158Block, newcfg.EIP158Block, head) {
		return newCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	}
	if c.IsEIP158(head) && !configNumEqual(c.ChainID, newcfg.ChainID) {
		return newCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	}
	if isForkIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, head) {
		return newCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	if isForkIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, head) {
		return newCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	}
	if isForkIncompatible(c.PetersburgBlock, newcfg.PetersburgBlock, head) {
		// the only case where we allow Petersburg to be set in the past is if it is equal to Constantinople
		// mainly to satisfy fork ordering requirements which state that Petersburg fork be set if Constantinople fork is set
		if isForkIncompatible(c.ConstantinopleBlock, newcfg.PetersburgBlock, head) {
			return newCompatError("Petersburg fork block", c.PetersburgBlock, newcfg.PetersburgBlock)
		}
	}
	if isForkIncompatible(c.IstanbulBlock, newcfg.IstanbulBlock, head) {
		return newCompatError("Istanbul fork block", c.IstanbulBlock, newcfg.IstanbulBlock)
	}
	if isForkIncompatible(c.MuirGlacierBlock, newcfg.MuirGlacierBlock, head) {
		return newCompatError("Muir Glacier fork block", c.MuirGlacierBlock, newcfg.MuirGlacierBlock)
	}
	if isForkIncompatible(c.BerlinBlock, newcfg.BerlinBlock, head) {
		return newCompatError("Berlin fork block", c.BerlinBlock, newcfg.BerlinBlock)
	}
	if isForkIncompatible(c.LondonBlock, newcfg.LondonBlock, head) {
		return newCompatError("London fork block", c.LondonBlock, newcfg.LondonBlock)
	}
	if isForkIncompatible(c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock, head) {
		return newCompatError("Arrow Glacier fork block", c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock)
	}
	if isForkIncompatible(c.OdysseusBlock, newcfg.OdysseusBlock, head) {
		return newCompatError("Odysseus fork block", c.OdysseusBlock, newcfg.OdysseusBlock)
	}
	if isForkIncompatible(c.FenixBlock, newcfg.FenixBlock, head) {
		return newCompatError("Fenix fork block", c.FenixBlock, newcfg.FenixBlock)
	}
	if isForkIncompatible(c.ConsortiumV2Block, newcfg.ConsortiumV2Block, head) {
		return newCompatError("Consortium v2 fork block", c.ConsortiumV2Block, newcfg.ConsortiumV2Block)
	}
	if isForkIncompatible(c.PuffyBlock, newcfg.PuffyBlock, head) {
		return newCompatError("Puffy fork block", c.PuffyBlock, newcfg.PuffyBlock)
	}
	if isForkIncompatible(c.BubaBlock, newcfg.BubaBlock, head) {
		return newCompatError("Buba fork block", c.BubaBlock, newcfg.BubaBlock)
	}
	if isForkIncompatible(c.OlekBlock, newcfg.OlekBlock, head) {
		return newCompatError("Olek fork block", c.OlekBlock, newcfg.OlekBlock)
	}
	if isForkIncompatible(c.ShillinBlock, newcfg.ShillinBlock, head) {
		return newCompatError("Shillin fork block", c.ShillinBlock, newcfg.ShillinBlock)
	}
	if isForkIncompatible(c.AntennaBlock, newcfg.AntennaBlock, head) {
		return newCompatError("Antenna fork block", c.AntennaBlock, newcfg.AntennaBlock)
	}
	if isForkIncompatible(c.MikoBlock, newcfg.MikoBlock, head) {
		return newCompatError("Miko fork block", c.MikoBlock, newcfg.MikoBlock)
	}
	if isForkIncompatible(c.TrippBlock, newcfg.TrippBlock, head) {
		return newCompatError("Tripp fork block", c.TrippBlock, newcfg.TrippBlock)
	}
	if isForkIncompatible(c.AaronBlock, newcfg.AaronBlock, head) {
		return newCompatError("Aaron fork block", c.AaronBlock, newcfg.AaronBlock)
	}
	if isForkIncompatible(c.ShanghaiBlock, newcfg.ShanghaiBlock, head) {
		return newCompatError("Shanghai fork block", c.ShanghaiBlock, newcfg.ShanghaiBlock)
	}
	if isForkIncompatible(c.CancunBlock, newcfg.CancunBlock, head) {
		return newCompatError("Cancun fork block", c.CancunBlock, newcfg.CancunBlock)
	}
	if isForkIncompatible(c.VenokiBlock, newcfg.VenokiBlock, head) {
		return newCompatError("Venoki fork block", c.VenokiBlock, newcfg.VenokiBlock)
	}
	if isForkIncompatible(c.PragueBlock, newcfg.PragueBlock, head) {
		return newCompatError("Prague fork block", c.PragueBlock, newcfg.PragueBlock)
	}
	return nil
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntactic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID                                                 *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158               bool
	IsByzantium, IsConstantinople, IsPetersburg, IsIstanbul bool
	IsBerlin, IsLondon, IsOdysseusFork                      bool
	IsFenix, IsShillin, IsConsortiumV2, IsAntenna           bool
	IsMiko, IsTripp, IsAaron, IsShanghai, IsCancun          bool
	IsVenoki, IsLastConsortiumV1Block, IsPrague             bool
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) Rules(num *big.Int) Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	return Rules{
		ChainID:                 new(big.Int).Set(chainID),
		IsHomestead:             c.IsHomestead(num),
		IsEIP150:                c.IsEIP150(num),
		IsEIP155:                c.IsEIP155(num),
		IsEIP158:                c.IsEIP158(num),
		IsByzantium:             c.IsByzantium(num),
		IsConstantinople:        c.IsConstantinople(num),
		IsPetersburg:            c.IsPetersburg(num),
		IsIstanbul:              c.IsIstanbul(num),
		IsBerlin:                c.IsBerlin(num),
		IsLondon:                c.IsLondon(num),
		IsOdysseusFork:          c.IsOdysseus(num),
		IsFenix:                 c.IsFenix(num),
		IsShillin:               c.IsShillin(num),
		IsLastConsortiumV1Block: c.IsLastConsortiumV1Block(num),
		IsConsortiumV2:          c.IsConsortiumV2(num),
		IsAntenna:               c.IsAntenna(num),
		IsMiko:                  c.IsMiko(num),
		IsTripp:                 c.IsTripp(num),
		IsAaron:                 c.IsAaron(num),
		IsShanghai:              c.IsShanghai(num),
		IsCancun:                c.IsCancun(num),
		IsVenoki:                c.IsVenoki(num),
		IsPrague:                c.IsPrague(num),
	}
}
