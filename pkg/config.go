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
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/viper"
)

type TxParams struct {
	// Name of this tx in the .toml file
	Name string

	// HTTP Client for this tx type
	Client *rpc.Client

	// Type of the tx
	Type TxType

	// Universal tx fields
	To *common.Address
	GasLimit uint64
	GasPrice *big.Int // nil if eip1559
	Amount *big.Int
	Data []byte
	Sender common.Address

	// Optimism-specific metadata fields
	L1SenderAddr *common.Address
	L1RollupTxId uint64
	SigHashType uint8

	// EIP1559-specific fields
	GasPremium *big.Int
	FeeCap *big.Int

	// Sender key, if left the senderKeyPath is empty we generate a new key
	SenderKey *ecdsa.PrivateKey

	// Sending params
	// How often we send a tx of this type
	Frequency time.Duration
	// Total number of txs of this type to send
	TotalNumber uint64
	// Txs of different types will be sent according to their order (starting at 0)
	Order uint64
}

const (
	ETH_TX_LIST = "ETH_TX_LIST"

	typeSuffix = ".type"
	httpPathSuffix = ".http"
	toSuffix = ".to"
	amountSuffix = ".amount"
	gasLimitSuffix = ".gasLimit"
	gasPriceSuffix = ".gasPrice"
	gasPremiumSuffix = ".gasPremium"
	feeCapSuffix = ".feeCap"
	dataSuffix = ".data"
	senderKeyPathSuffix = ".senderKeyPath"
	writeSenderPathSuffix = ".writeSenderPath"
	l1SenderSuffix = ".l1Sender"
	l1RollupTxIdSuffix = ".l1RollupTxId"
	sigHashTypeSuffix = ".sigHashType"
	frequencySuffix = ".frequency"
	totalNumberSuffix = ".totalNumber"
)

// NewConfig returns a new tx spammer config
func NewTxParams() ([]TxParams, error) {
	viper.BindEnv("eth.txs", ETH_TX_LIST)

	txs := viper.GetStringSlice("eth.txs")
	txParams := make([]TxParams, len(txs))
	for i, txName := range txs {
		// Get http client
		httpPathStr := viper.GetString(txName+httpPathSuffix)
		if httpPathStr == "" {
			return nil, fmt.Errorf("tx %s is missing an httpPath", txName)
		}
		if !strings.HasPrefix(httpPathStr, "http://") {
			httpPathStr = "http://" + httpPathStr
		}
		rpcClient, err := rpc.Dial(httpPathStr)
		if err != nil {
			return nil, err
		}

		// Get tx type
		txTypeStr := viper.GetString(txName+typeSuffix)
		if txTypeStr == "" {
			return nil, fmt.Errorf("need tx type for tx %s", txName)
		}
		txType, err := TxTypeFromString(txTypeStr)
		if err != nil {
			return nil, err
		}

		// Get basic fields
		toStr := viper.GetString(txName+toSuffix)
		var toAddr *common.Address
		if toStr != "" {
			to := common.HexToAddress(toStr)
			toAddr = &to
		}
		amountStr := viper.GetString(txName+amountSuffix)
		amount := new(big.Int)
		if amountStr != "" {
			if _, ok := amount.SetString(amountStr, 10); !ok {
				return nil, fmt.Errorf("amount (%s) for tx %s is not valid", amountStr, txName)
			}
		}
		gasPriceStr := viper.GetString(txName+gasPriceSuffix)
		var gasPrice *big.Int
		if gasPriceStr != "" {
			gasPrice = new(big.Int)
			if _, ok := gasPrice.SetString(gasPriceStr, 10); !ok {
				return nil, fmt.Errorf("gasPrice (%s) for tx %s is not valid", gasPriceStr, txName)
			}
		}
		gasLimit := viper.GetUint64(txName+gasLimitSuffix)
		hex := viper.GetString(txName+dataSuffix)
		data := make([]byte, 0)
		if hex != "" {
			data = common.Hex2Bytes(hex)
		}

		// Load or generate sender key
		senderKeyPath := viper.GetString(txName+senderKeyPathSuffix)
		var key *ecdsa.PrivateKey
		if senderKeyPath != "" {
			key, err = crypto.LoadECDSA(senderKeyPath)
			if err != nil {
				return nil, fmt.Errorf("unable to load ecdsa at %s key for tx %s", senderKeyPath, txName)
			}
		} else {
			key, err = crypto.GenerateKey()
			if err != nil {
				return nil, fmt.Errorf("unable to generate ecdsa key for tx %s", txName)
			}
			writePath := viper.GetString(txName+writeSenderPathSuffix)
			if writePath != "" {
				if err := crypto.SaveECDSA(writePath, key); err != nil {
					return nil, err
				}
			}
		}

		// Attempt to load Optimism fields
		l1SenderStr := viper.GetString(txName+l1SenderSuffix)
		var l1Sender *common.Address
		if l1SenderStr != "" {
			sender := common.HexToAddress(l1SenderStr)
			l1Sender = &sender
		}

		l1RollupTxId := viper.GetUint64(txName+l1RollupTxIdSuffix)
		sigHashType := viper.GetUint(txName+sigHashTypeSuffix)

		// If gasPrice was empty, attempt to load EIP1559 fields
		var feeCap, gasPremium *big.Int
		if gasPrice == nil {
			feeCapStr := viper.GetString(txName+feeCapSuffix)
			gasPremiumString := viper.GetString(txName+gasPremiumSuffix)
			if feeCapStr == "" {
				return nil, fmt.Errorf("tx %s is missing feeCapStr", txName)
			}
			if  gasPremiumString == "" {
				return nil, fmt.Errorf("tx %s is missing gasPremiumStr", txName)
			}
			feeCap = new(big.Int)
			gasPremium = new(big.Int)
			if _, ok := feeCap.SetString(feeCapStr, 10); !ok {
				return nil, fmt.Errorf("unable to set feeCap to %s for tx %s", feeCapStr, txName)
			}
			if _, ok := gasPremium.SetString(gasPremiumString, 10); !ok {
				return nil, fmt.Errorf("unable to set gasPremium to %s for tx %s", gasPremiumString, txName)
			}
		}

		// Load the sending paramas
		frequency := viper.GetDuration(txName+frequencySuffix)
		totalNumber := viper.GetUint64(txName+totalNumberSuffix)

		txParams[i] = TxParams{
			Client: rpcClient,
			Type: txType,
			Name: txName,
			To: toAddr,
			Amount: amount,
			GasLimit: gasLimit,
			GasPrice: gasPrice,
			GasPremium: gasPremium,
			FeeCap: feeCap,
			Data: data,
			L1SenderAddr: l1Sender,
			L1RollupTxId: l1RollupTxId,
			SigHashType: uint8(sigHashType),
			Frequency: frequency,
			TotalNumber: totalNumber,
		}
	}
	return txParams, nil
}
