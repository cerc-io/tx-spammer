// VulcanizeDB
// Copyright © 2020 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package tx_spammer

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
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
	case 5, 420:
		return params.GoerliChainConfig, nil
	default:
		return nil, fmt.Errorf("chain config for chainid %d not available", chainID)
	}
}

// ChainConfig returns the appropriate ethereum chain config for the provided chain id
func TxSigner(chainID uint64) (types.Signer, error) {
	switch chainID {
	case 1, 3, 4, 5:
		return types.NewEIP155Signer(new(big.Int).SetUint64(chainID)), nil
	case 420:
		return types.NewOVMSigner(new(big.Int).SetUint64(chainID)), nil
	default:
		return nil, fmt.Errorf("chain config for chainid %d not available", chainID)
	}
}
