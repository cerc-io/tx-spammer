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

package shared

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// ChainConfig returns the appropriate ethereum chain config for the provided chain id
func TxSigner(kind TxType, chainID uint64) (types.Signer, error) {
	switch kind {
	case Standard, EIP1559:
		return types.NewEIP155Signer(new(big.Int).SetUint64(chainID)), nil
	case OptimismL2, OptimismL1ToL2:
		return types.NewOVMSigner(new(big.Int).SetUint64(chainID)), nil
	default:
		return nil, fmt.Errorf("chain config for chainid %d not available", chainID)
	}
}

// SendRawTransaction sends a raw, signed tx using the provided client
func SendRawTransaction(rpcClient *rpc.Client, txRlp []byte) error {
	return rpcClient.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(txRlp))
}

// WriteContractAddr appends a contract addr to an out file
func WriteContractAddr(filePath string, senderAddr common.Address, nonce uint64) (common.Address, error) {
	if filePath == "" {
		filePath = DefaultDeploymentAddrLogPathPrefix + senderAddr.Hex()
	}
	contractAddr := crypto.CreateAddress(senderAddr, nonce)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return common.Address{}, err
	}
	if _, err := f.WriteString(contractAddr.Hex() + "\n"); err != nil {
		return common.Address{}, err
	}
	return contractAddr, f.Close()
}
