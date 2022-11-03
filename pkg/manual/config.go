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

package manual

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/viper"
)

// TxParams holds the parameters for a given transaction
type TxParams struct {
	// Name of this tx in the .toml file
	Name string

	// HTTP Client for this tx type
	Client *rpc.Client

	// DynamicFeeTx properties - Start
	ChainID   *big.Int
	Nonce     uint64
	GasTipCap *big.Int // a.k.a. maxPriorityFeePerGas
	GasFeeCap *big.Int // a.k.a. maxFeePerGas
	Gas       uint64
	To        *common.Address // nil means contract creation
	Value     *big.Int
	Data      []byte
	// DynamicFeeTx properties - End

	// Sender key, if left the senderKeyPath is empty we generate a new key
	SenderKey             *ecdsa.PrivateKey
	StartingNonce         uint64
	ContractAddrWritePath string

	// Sending params
	Sender common.Address
	// How often we send a tx of this type
	Frequency time.Duration
	// Total number of txs of this type to send
	TotalNumber uint64
	// Delay before beginning to send
	Delay time.Duration
}

// NewTxParams NewConfig returns a new tx spammer config
func NewTxParams() ([]TxParams, error) {
	bindEnv()
	addrLogPath := viper.GetString("eth.addrLogPath")
	txs := viper.GetStringSlice("eth.txs")
	txParams := make([]TxParams, len(txs))
	for i, txName := range txs {
		// Get http client
		httpPathStr := viper.GetString(txName + httpPathSuffix)
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

		// Get basic fields
		toStr := viper.GetString(txName + toSuffix)
		var toAddr *common.Address
		if toStr != "" {
			to := common.HexToAddress(toStr)
			toAddr = &to
		}
		amountStr := viper.GetString(txName + amountSuffix)
		amount := new(big.Int)
		if amountStr != "" {
			if _, ok := amount.SetString(amountStr, 10); !ok {
				return nil, fmt.Errorf("amount (%s) for tx %s is not valid", amountStr, txName)
			}
		}
		gasLimit := viper.GetUint64(txName + gasLimitSuffix)
		hex := viper.GetString(txName + dataSuffix)
		var data []byte = nil
		if hex != "" {
			data = common.Hex2Bytes(hex)
		}

		// Load or generate sender key
		senderKeyPath := viper.GetString(txName + senderKeyPathSuffix)
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
			writePath := viper.GetString(txName + writeSenderPathSuffix)
			if writePath == "" {
				writePath = defaultGenKeyWritePathPrefix + txName
			}
			if err := crypto.SaveECDSA(writePath, key); err != nil {
				return nil, err
			}
		}
		sender := crypto.PubkeyToAddress(key.PublicKey)
		if err := writeSenderAddr(addrLogPath, sender); err != nil {
			return nil, err
		}

		// If gasPrice was empty, attempt to load EIP1559 fields
		var feeCap, tipCap *big.Int
		feeCapStr := viper.GetString(txName + feeCapSuffix)
		tipCapStr := viper.GetString(txName + tipCapSuffix)
		if feeCapStr == "" {
			return nil, fmt.Errorf("tx %s is missing feeCapStr", txName)
		}
		if tipCapStr == "" {
			return nil, fmt.Errorf("tx %s is missing tipCapStr", txName)
		}
		feeCap = new(big.Int)
		tipCap = new(big.Int)
		if _, ok := feeCap.SetString(feeCapStr, 10); !ok {
			return nil, fmt.Errorf("unable to set feeCap to %s for tx %s", feeCapStr, txName)
		}
		if _, ok := tipCap.SetString(tipCapStr, 10); !ok {
			return nil, fmt.Errorf("unable to set tipCap to %s for tx %s", tipCapStr, txName)
		}

		txParams[i] = TxParams{
			Name:                  txName,
			Client:                rpcClient,
			GasTipCap:             tipCap,
			GasFeeCap:             feeCap,
			Gas:                   gasLimit,
			To:                    toAddr,
			Value:                 amount,
			Data:                  data,
			Sender:                sender,
			SenderKey:             key,
			StartingNonce:         viper.GetUint64(txName + startingNonceSuffix),
			ContractAddrWritePath: viper.GetString(txName + contractWriteSuffix),
			Frequency:             viper.GetDuration(txName+frequencySuffix) * time.Millisecond,
			TotalNumber:           viper.GetUint64(txName + totalNumberSuffix),
			Delay:                 viper.GetDuration(txName+delaySuffix) * time.Millisecond,
		}
	}
	return txParams, nil
}

func writeSenderAddr(filePath string, senderAddr common.Address) error {
	if filePath == "" {
		filePath = defaultAddrLogPath
	}
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(senderAddr.Hex() + "\n"); err != nil {
		return err
	}
	return f.Close()
}
