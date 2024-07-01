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
	"math/big"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

// TxSigner returns the Cancun signer at the provided block height
func TxSigner(chainID *big.Int) types.Signer {
	return types.NewCancunSigner(chainID)
}

// SendTransaction sends a signed tx using the provided client
func SendTransaction(rpcClient *rpc.Client, tx *types.Transaction) error {
	msg, err := core.TransactionToMessage(tx, TxSigner(tx.ChainId()), big.NewInt(1))
	if err != nil {
		return err
	}
	if nil == tx.To() {
		logrus.WithFields(logrus.Fields{
			"hash":     tx.Hash(),
			"from":     msg.From,
			"contract": crypto.CreateAddress(msg.From, tx.Nonce()),
		}).Debug("Sending TX to create contract")
	} else {
		fields := logrus.Fields{
			"hash":  tx.Hash(),
			"from":  msg.From,
			"to":    tx.To(),
			"nonce": tx.Nonce(),
		}
		if numblobs := len(tx.BlobHashes()); numblobs > 0 {
			fields["blobs"] = numblobs
		}
		if len(tx.Data()) == 0 {
			fields["value"] = tx.Value()
			logrus.WithFields(fields).Debug("Sending TX to transfer ETH")
		} else {
			logrus.WithFields(fields).Debug("Sending TX to call contract")
		}
	}
	data, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	return SendRawTransaction(rpcClient, data)
}

// SendRawTransaction sends a raw, signed tx using the provided client
func SendRawTransaction(rpcClient *rpc.Client, raw []byte) error {
	logrus.Debugf("eth_sendRawTransaction: %x... (%d bytes)", raw[:min(10, len(raw))], len(raw))
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
