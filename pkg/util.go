package tx_spammer

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

// ChainConfig returns the appropriate ethereum chain config for the provided chain id
func ChainConfig(chainID uint64) (*params.ChainConfig, error) {
	switch chainID {
	case 1:
		return params.MainnetChainConfig, nil
	case 3:
		return params.TestnetChainConfig, nil // Ropsten
	case 4:
		return params.RinkebyChainConfig, nil
	case 5:
		return params.GoerliChainConfig, nil
	case 420:
	default:
		return nil, fmt.Errorf("chain config for chainid %d not available", chainID)
	}
}

// ChainConfig returns the appropriate ethereum chain config for the provided chain id
func TxSigner(chainID uint64) (types.Signer, error) {
	switch chainID {
	case 1:
		return params.MainnetChainConfig, nil
	case 3:
		return params.TestnetChainConfig, nil // Ropsten
	case 4:
		return params.RinkebyChainConfig, nil
	case 5:
		return params.GoerliChainConfig, nil
	case 420:
		return types.NewOVMSigner(big.NewInt()), nil
	default:
		return nil, fmt.Errorf("chain config for chainid %d not available", chainID)
	}
}