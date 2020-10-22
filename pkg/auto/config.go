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
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
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

const (
	// toml bindings
	ethKeyDirPath   = "eth.keyDirPath"
	ethAddrFilePath = "eth.addrFilePath"
	ethHttpPath     = "eth.httpPath"
	ethChainID      = "eth.chainID"
	ethType         = "eth.type"

	ethDeploymentNumber   = "deployment.number"
	ethDeploymentData     = "deployment.hexData"
	ethDeploymentGasPrice = "deployment.gasPrice"
	ethDeploymentGasLimit = "deployment.gasLimit"

	ethOptimismL1Sender    = "optimism.l1Sender"
	ethOptimismRollupTxID  = "optimism.l1RollupTxId"
	ethOptimismSigHashType = "optimism.sigHashType"
	ethOptimismQueueOrigin = "optimism.queueOrigin"

	ethCallFrequency     = "contractSpammer.frequency"
	ethCallTotalNumber   = "contractSpammer.totalNumber"
	ethCallABIPath       = "contractSpammer.abiPath"
	ethCallMethodName    = "contractSpammer.methodName"
	ethCallPositionStart = "contractSpammer.positionStart"
	ethCallPositionEnd   = "contractSpammer.positionEnd"
	ethCallStorageValue  = "contractSpammer.storageValue"
	ethCallGasLimit      = "contractSpammer.gasLimit"
	ethCallGasPrice      = "contractSpammer.gasPrice"

	ethSendFrequency   = "sendSpammer.frequency"
	ethSendTotalNumber = "sendSpammer.totalNumber"
	ethSendAmount      = "sendSpammer.amount"
	ethSendGasLimit    = "sendSpammer.gasLimit"
	ethSendGasPrice    = "sendSpammer.gasPrice"

	// env variables
	ETH_KEY_DIR_PATH  = "ETH_KEY_DIR_PATH"
	ETH_ADDR_DIR_PATH = "ETH_ADDR_DIR_PATH"
	ETH_HTTP_PATH     = "ETH_HTTP_PATH"
	ETH_CHAIN_ID      = "ETH_CHAIN_ID"
	ETH_TX_TYPE       = "ETH_TX_TYPE"

	ETH_DEPLOYMENT_NUMBER    = "ETH_DEPLOYMENT_NUMBER"
	ETH_DEPLOYMENT_HEX_DATA  = "ETH_DEPLOYMENT_HEX_DATA"
	ETH_DEPLOYMENT_GAS_LIMIT = "ETH_DEPLOYMENT_GAS_LIMIT"
	ETH_DEPLOYMENT_GAS_PRICE = "ETH_DEPLOYMENT_GAS_PRICE"

	ETH_OPTIMISM_L1_SENDER     = "ETH_OPTIMISM_L1_SENDER"
	ETH_OPTIMISM_ROLLUP_TX_ID  = "ETH_OPTIMISM_ROLLUP_TX_ID"
	ETH_OPTIMISM_SIG_HASH_TYPE = "ETH_OPTIMISM_SIG_HASH_TYPE"
	ETH_OPTIMISM_QUEUE_ORIGIN  = "ETH_OPTIMISM_QUEUE_ORIGIN"

	ETH_CALL_FREQ           = "ETH_CALL_FREQ"
	ETH_CALL_TOTAL_NUMBER   = "ETH_CALL_TOTAL_NUMBER"
	ETH_CALL_ABI_PATH       = "ETH_CALL_ABI_PATH"
	ETH_CALL_METHOD_NAME    = "ETH_CALL_METHOD_NAME"
	ETH_CALL_POSITION_START = "ETH_CALL_POSITION_START"
	ETH_CALL_POSITION_END   = "ETH_CALL_POSITION_END"
	ETH_CALL_STORAGE_VALUE  = "ETH_CALL_STORAGE_VALUE"
	ETH_CALL_GAS_LIMIT      = "ETH_CALL_GAS_LIMIT"
	ETH_CALL_GAS_PRICE      = "ETH_CALL_GAS_PRICE"

	ETH_SEND_FREQ         = "ETH_SEND_FREQ"
	ETH_SEND_TOTAL_NUMBER = "ETH_SEND_TOTAL_NUMBER"
	ETH_SEND_AMOUNT       = "ETH_SEND_AMOUNT"
	ETH_SEND_GAS_LIMIT    = "ETH_SEND_GAS_LIMIT"
	ETH_SEND_GAS_PRICE    = "ETH_SEND_GAS_PRICE"
)

// Config holds all the parameters for the auto tx spammer
type Config struct {
	// HTTP client for sending transactions
	Client *rpc.Client

	// Key pairs for the accounts we will use to deploy contracts and send txs
	SenderKeys []*ecdsa.PrivateKey

	// Addresses to send eth transfer txs to
	DestinationAddrs []common.Address

	// Type of the txs we are working with
	Type shared.TxType

	// Chain ID for the chain we are working with
	ChainID uint64

	// Optimism-specific metadata fields (optional)
	L1SenderAddr *common.Address
	L1RollupTxId *hexutil.Uint64
	SigHashType  types.SignatureHashType
	QueueOrigin  types.QueueOrigin

	// Configuration for the initial contract deployment
	DeploymentConfig *DeploymentConfig

	// Configuration for the contract calling txs
	CallConfig *CallConfig

	// Configuration for the eth transfer txs
	SendConfig *SendConfig
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
	GasLimit      uint64
	GasPrice      *big.Int
	MethodName    string
	ABI           abi.ABI
	PositionStart uint64
	PositionEnd   uint64
	StorageValue  uint64

	Frequency time.Duration
	Number    uint64
}

// SendConfig holds the parameters for the eth transfer txs
type SendConfig struct {
	GasLimit uint64
	GasPrice *big.Int
	Amount   *big.Int

	Frequency time.Duration
	Number    uint64
}

func NewConfig() (*Config, error) {
	viper.BindEnv(ethKeyDirPath, ETH_KEY_DIR_PATH)
	viper.BindEnv(ethAddrFilePath, ETH_ADDR_DIR_PATH)
	viper.BindEnv(ethHttpPath, ETH_HTTP_PATH)
	viper.BindEnv(ethType, ETH_TX_TYPE)
	viper.BindEnv(ethChainID, ETH_CHAIN_ID)

	viper.BindEnv(ethOptimismL1Sender, ETH_OPTIMISM_L1_SENDER)
	viper.BindEnv(ethOptimismQueueOrigin, ETH_OPTIMISM_QUEUE_ORIGIN)
	viper.BindEnv(ethOptimismRollupTxID, ETH_OPTIMISM_ROLLUP_TX_ID)
	viper.BindEnv(ethOptimismSigHashType, ETH_OPTIMISM_SIG_HASH_TYPE)

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
	}

	// Load eth transfer destination addresses
	addrs := make([]common.Address, 0)
	addrFilePath := viper.GetString(ethAddrFilePath)
	if addrFilePath == "" {
		return nil, fmt.Errorf("missing %s", ethAddrFilePath)
	}
	addrFile, err := os.Open(addrFilePath)
	if err != nil {
		return nil, err
	}
	defer addrFile.Close()
	scanner := bufio.NewScanner(addrFile)
	for scanner.Scan() {
		addrBytes := scanner.Bytes()
		addr := common.BytesToAddress(addrBytes)
		addrs = append(addrs, addr)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Load tx type
	txType, err := shared.TxTypeFromString(viper.GetString(ethType))
	if err != nil {
		return nil, err
	}

	// Load optimism params
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
		DestinationAddrs: addrs,
		Type:             txType,
		ChainID:          viper.GetUint64(ethChainID),

		L1SenderAddr: l1Sender,
		L1RollupTxId: &l1rtid,
		SigHashType:  (types.SignatureHashType)(uint8(sigHashType)),
		QueueOrigin:  (types.QueueOrigin)(queueOrigin),

		DeploymentConfig: deploymentConfig,
		CallConfig:       callConfig,
		SendConfig:       sendConfig,
	}, nil
}

// NewDeploymentConfig constructs and returns a new DeploymentConfig
func NewDeploymentConfig() (*DeploymentConfig, error) {
	viper.BindEnv(ethDeploymentNumber, ETH_DEPLOYMENT_NUMBER)
	viper.BindEnv(ethDeploymentData, ETH_DEPLOYMENT_HEX_DATA)
	viper.BindEnv(ethDeploymentGasLimit, ETH_DEPLOYMENT_GAS_LIMIT)
	viper.BindEnv(ethDeploymentGasPrice, ETH_DEPLOYMENT_GAS_PRICE)

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
	viper.BindEnv(ethCallABIPath, ETH_CALL_ABI_PATH)
	viper.BindEnv(ethCallFrequency, ETH_CALL_FREQ)
	viper.BindEnv(ethCallGasLimit, ETH_CALL_GAS_LIMIT)
	viper.BindEnv(ethCallGasPrice, ETH_CALL_GAS_PRICE)
	viper.BindEnv(ethCallMethodName, ETH_CALL_METHOD_NAME)
	viper.BindEnv(ethCallPositionEnd, ETH_CALL_POSITION_END)
	viper.BindEnv(ethCallPositionStart, ETH_CALL_POSITION_START)
	viper.BindEnv(ethCallStorageValue, ETH_CALL_STORAGE_VALUE)
	viper.BindEnv(ethCallTotalNumber, ETH_CALL_TOTAL_NUMBER)

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

	return &CallConfig{
		Number:        viper.GetUint64(ethCallTotalNumber),
		GasPrice:      gasPrice,
		GasLimit:      viper.GetUint64(ethCallGasLimit),
		MethodName:    methodName,
		ABI:           parsedABI,
		PositionEnd:   viper.GetUint64(ethCallPositionEnd),
		PositionStart: viper.GetUint64(ethCallPositionStart),
		StorageValue:  viper.GetUint64(ethCallStorageValue),
		Frequency:     viper.GetDuration(ethCallFrequency),
	}, nil
}

// NewSendConfig constructs and returns a new SendConfig
func NewSendConfig() (*SendConfig, error) {
	viper.BindEnv(ethSendFrequency, ETH_SEND_FREQ)
	viper.BindEnv(ethSendTotalNumber, ETH_SEND_TOTAL_NUMBER)
	viper.BindEnv(ethSendAmount, ETH_SEND_AMOUNT)
	viper.BindEnv(ethSendGasLimit, ETH_SEND_GAS_LIMIT)
	viper.BindEnv(ethSendGasPrice, ETH_SEND_GAS_PRICE)

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
	return &SendConfig{
		Frequency: viper.GetDuration(ethSendFrequency),
		Number:    viper.GetUint64(ethSendTotalNumber),
		Amount:    amount,
		GasPrice:  gasPrice,
		GasLimit:  viper.GetUint64(ethSendGasLimit),
	}, nil
}
