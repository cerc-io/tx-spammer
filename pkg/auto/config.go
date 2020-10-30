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

package auto

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/viper"
	"github.com/vulcanize/tx_spammer/pkg/shared"
)

var (
	receiverAddressSeed = common.HexToAddress("0xe48C9A989438606a79a7560cfba3d34BAfBAC38E")
	storageAddressSeed  = common.HexToAddress("0x029298Ac95662F2b54A7F1116f3F8105eb2b00F5")
)

func init() {
	bindEnv()
}

// Config holds all the parameters for the auto tx spammer
type Config struct {
	// HTTP client for sending transactions
	Client *rpc.Client

	// Key pairs for the accounts we will use to deploy contracts and send txs
	SenderKeys  []*ecdsa.PrivateKey
	SenderAddrs []common.Address

	// Type of the txs we are working with
	Type shared.TxType

	// Tx signer for the chain we are working with
	Signer types.Signer

	// Configuration for Optimism L2
	OptimismConfig *OptimismConfig

	// Configuration for the initial contract deployment
	DeploymentConfig *DeploymentConfig

	// Configuration for the contract calling txs
	CallConfig *CallConfig

	// Configuration for the eth transfer txs
	SendConfig *SendConfig

	// Configuration for EIP1559
	EIP1559Config *EIP1559Config
}

// OptimismConfig holds the tx paramaters specific to Optimism L2
type OptimismConfig struct {
	L1SenderAddr *common.Address
	L1RollupTxId *hexutil.Uint64
	SigHashType  types.SignatureHashType
	QueueOrigin  types.QueueOrigin
}

// DeploymentConfig holds the parameters for the contract deployment contracts
type DeploymentConfig struct {
	GasLimit uint64
	GasPrice *big.Int
	Data     []byte

	Number uint64
}

// CallConfig holds the parameters for the contract calling txs
type CallConfig struct {
	GasLimit uint64
	GasPrice *big.Int

	MethodName   string
	ABI          abi.ABI
	StorageValue uint64
	StorageAddrs []common.Address

	Frequency time.Duration
}

// SendConfig holds the parameters for the eth transfer txs
type SendConfig struct {
	GasLimit uint64
	GasPrice *big.Int
	Amount   *big.Int

	DestinationAddresses []common.Address
	Frequency            time.Duration
}

// todo: EIP1559Config
type EIP1559Config struct{}

func NewConfig() (*Config, error) {
	// Initialize rpc client
	httpPathStr := viper.GetString(ethHttpPath)
	if httpPathStr == "" {
		return nil, fmt.Errorf("missing %s", ethHttpPath)
	}
	if !strings.HasPrefix(httpPathStr, "http://") {
		httpPathStr = "http://" + httpPathStr
	}
	rpcClient, err := rpc.Dial(httpPathStr)
	if err != nil {
		return nil, err
	}

	// Load keys
	keyDirPath := viper.GetString(ethKeyDirPath)
	if keyDirPath == "" {
		return nil, fmt.Errorf("missing %s", ethKeyDirPath)
	}
	keyFiles, err := ioutil.ReadDir(keyDirPath)
	if err != nil {
		return nil, err
	}
	keys := make([]*ecdsa.PrivateKey, 0)
	senderAddrs := make([]common.Address, 0)
	for _, keyFile := range keyFiles {
		if keyFile.IsDir() {
			continue
		}
		filePath := filepath.Join(keyDirPath, keyFile.Name())
		key, err := crypto.LoadECDSA(filePath)
		if err != nil {
			return nil, fmt.Errorf("unable to load ecdsa key file at %s", filePath)
		}
		keys = append(keys, key)
		senderAddrs = append(senderAddrs, crypto.PubkeyToAddress(key.PublicKey))
	}

	// Load tx type
	txType, err := shared.TxTypeFromString(viper.GetString(ethType))
	if err != nil {
		return nil, err
	}

	// Load signer
	chainID := viper.GetUint64(ethChainID)
	signer, err := shared.TxSigner(txType, chainID)
	if err != nil {
		return nil, err
	}

	// Load optimism config
	var optimismConfig *OptimismConfig
	if txType == shared.OptimismL1ToL2 || txType == shared.OptimismL2 {
		optimismConfig = NewOptimismConfig()
	}

	// Load deployment config
	deploymentConfig, err := NewDeploymentConfig()
	if err != nil {
		return nil, err
	}

	// Load call config
	callConfig, err := NewCallConfig()
	if err != nil {
		return nil, err
	}

	// Load send config
	sendConfig, err := NewSendConfig()
	if err != nil {
		return nil, err
	}

	// Assemble and return
	return &Config{
		Client:           rpcClient,
		SenderKeys:       keys,
		SenderAddrs:      senderAddrs,
		Type:             txType,
		Signer:           signer,
		OptimismConfig:   optimismConfig,
		DeploymentConfig: deploymentConfig,
		CallConfig:       callConfig,
		SendConfig:       sendConfig,
	}, nil
}

// NewOptimismConfig constructs and returns a new OptimismConfig
func NewOptimismConfig() *OptimismConfig {
	l1SenderStr := viper.GetString(ethOptimismL1Sender)
	var l1Sender *common.Address
	if l1SenderStr != "" {
		sender := common.HexToAddress(l1SenderStr)
		l1Sender = &sender
	}
	l1RollupTxId := viper.GetUint64(ethOptimismRollupTxID)
	l1rtid := (hexutil.Uint64)(l1RollupTxId)
	sigHashType := viper.GetUint(ethOptimismSigHashType)
	queueOrigin := viper.GetInt64(ethOptimismQueueOrigin)
	return &OptimismConfig{
		L1SenderAddr: l1Sender,
		L1RollupTxId: &l1rtid,
		SigHashType:  (types.SignatureHashType)(uint8(sigHashType)),
		QueueOrigin:  (types.QueueOrigin)(queueOrigin),
	}
}

// NewDeploymentConfig constructs and returns a new DeploymentConfig
func NewDeploymentConfig() (*DeploymentConfig, error) {
	hexData := viper.GetString(ethDeploymentData)
	data := common.Hex2Bytes(hexData)
	gasPriceStr := viper.GetString(ethDeploymentGasPrice)
	gasPrice, ok := new(big.Int).SetString(gasPriceStr, 10)
	if !ok {
		return nil, fmt.Errorf("unable to convert gasPrice string (%s) into big.Int", gasPriceStr)
	}

	return &DeploymentConfig{
		Number:   viper.GetUint64(ethDeploymentNumber),
		Data:     data,
		GasPrice: gasPrice,
		GasLimit: viper.GetUint64(ethDeploymentGasLimit),
	}, nil
}

// NewCallConfig constructs and returns a new CallConfig
func NewCallConfig() (*CallConfig, error) {
	gasPriceStr := viper.GetString(ethCallGasPrice)
	gasPrice, ok := new(big.Int).SetString(gasPriceStr, 10)
	if !ok {
		return nil, fmt.Errorf("unable to convert gasPrice string (%s) into big.Int", gasPriceStr)
	}
	abiPath := viper.GetString(ethCallABIPath)
	if abiPath == "" {
		return nil, fmt.Errorf("missing contractSpammer.abiPath")
	}
	abiBytes, err := ioutil.ReadFile(abiPath)
	if err != nil {
		return nil, err
	}
	parsedABI, err := abi.JSON(bytes.NewReader(abiBytes))
	if err != nil {
		return nil, err
	}
	methodName := viper.GetString(ethCallMethodName)
	_, exist := parsedABI.Methods[methodName]
	if !exist {
		return nil, fmt.Errorf("method '%s' not found in provided abi", methodName)
	}
	number := viper.GetUint64(ethCallTotalNumber)
	addrs := make([]common.Address, number)
	for i := uint64(0); i < number; i++ {
		addrs[i] = crypto.CreateAddress(storageAddressSeed, i)
	}
	return &CallConfig{
		GasPrice:     gasPrice,
		GasLimit:     viper.GetUint64(ethCallGasLimit),
		MethodName:   methodName,
		ABI:          parsedABI,
		StorageValue: viper.GetUint64(ethCallStorageValue),
		Frequency:    viper.GetDuration(ethCallFrequency),
		StorageAddrs: addrs,
	}, nil
}

// NewSendConfig constructs and returns a new SendConfig
func NewSendConfig() (*SendConfig, error) {
	amountStr := viper.GetString(ethSendAmount)
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return nil, fmt.Errorf("unable to convert amount string (%s) into big.Int", amountStr)
	}
	gasPriceStr := viper.GetString(ethSendGasPrice)
	gasPrice, ok := new(big.Int).SetString(gasPriceStr, 10)
	if !ok {
		return nil, fmt.Errorf("unable to convert gasPrice string (%s) into big.Int", gasPriceStr)
	}
	number := viper.GetUint64(ethSendTotalNumber)
	addrs := make([]common.Address, number)
	for i := uint64(0); i < number; i++ {
		addrs[i] = crypto.CreateAddress(receiverAddressSeed, i)
	}
	return &SendConfig{
		DestinationAddresses: addrs,
		Frequency:            viper.GetDuration(ethSendFrequency),
		Amount:               amount,
		GasPrice:             gasPrice,
		GasLimit:             viper.GetUint64(ethSendGasLimit),
	}, nil
}
