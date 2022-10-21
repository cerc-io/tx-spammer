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
	"github.com/sirupsen/logrus"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// ChainConfig returns the appropriate ethereum chain config for the provided chain id
func TxSigner(chainID *big.Int) types.Signer {
	return types.NewLondonSigner(chainID)
}

// SendTransaction sends a signed tx using the provided client
func SendTransaction(rpcClient *rpc.Client, tx *types.Transaction) error {
	msg, _ := tx.AsMessage(TxSigner(tx.ChainId()), big.NewInt(1))
	if nil == tx.To() {
		logrus.Debugf("TX %s to create contract %s (sender %s)",
			tx.Hash().Hex(), crypto.CreateAddress(msg.From(), tx.Nonce()), msg.From().Hex())
	} else if nil == tx.Data() || len(tx.Data()) == 0 {
		logrus.Debugf("TX %s sending %s Wei to %s (sender %s)",
			tx.Hash().Hex(), tx.Value().String(), msg.To().Hex(), msg.From().Hex())
	} else {
		logrus.Debugf("TX %s calling contract %s (sender %s)",
			tx.Hash().Hex(), msg.To().Hex(), msg.From().Hex())
	}

	data, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	return SendRawTransaction(rpcClient, data)
}

// SendRawTransaction sends a raw, signed tx using the provided client
func SendRawTransaction(rpcClient *rpc.Client, raw []byte) error {
	logrus.Debug("eth_sendRawTransaction: ", hexutil.Encode(raw))
	return rpcClient.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(raw))
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
