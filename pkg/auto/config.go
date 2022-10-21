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
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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
	RpcClient *rpc.Client
	EthClient *ethclient.Client
	ChainID   *big.Int

	// Key pairs for the accounts we will use to deploy contracts and send txs
	SenderKeys  []*ecdsa.PrivateKey
	SenderAddrs []common.Address

	// Tx signer for the chain we are working with
	Signer types.Signer

	// Configuration for the initial contract deployment
	DeploymentConfig *DeploymentConfig

	// Configuration for the contract calling txs
	CallConfig *CallConfig

	// Configuration for the eth transfer txs
	SendConfig *SendConfig
}

// DeploymentConfig holds the parameters for the contract deployment contracts
type DeploymentConfig struct {
	ChainID   *big.Int
	GasLimit  uint64
	GasFeeCap *big.Int
	GasTipCap *big.Int
	Data      []byte

	Number uint64
}

// CallConfig holds the parameters for the contract calling txs
type CallConfig struct {
	ChainID   *big.Int
	GasLimit  uint64
	GasFeeCap *big.Int
	GasTipCap *big.Int
	Amount    *big.Int

	MethodName    string
	ABI           abi.ABI
	ContractAddrs []common.Address

	Frequency   time.Duration
	TotalNumber int
}

// SendConfig holds the parameters for the eth transfer txs
type SendConfig struct {
	ChainID   *big.Int
	GasLimit  uint64
	GasFeeCap *big.Int
	GasTipCap *big.Int
	Amount    *big.Int

	Frequency   time.Duration
	TotalNumber int
}

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

	ethClient, err := ethclient.Dial(httpPathStr)
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

	// Detect chain ID.
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	// Load signer
	signer := shared.TxSigner(chainID)

	// Load deployment config
	deploymentConfig, err := NewDeploymentConfig(chainID)
	if err != nil {
		return nil, err
	}

	// Load call config
	callConfig, err := NewCallConfig(chainID)
	if err != nil {
		return nil, err
	}

	// Load send config
	sendConfig, err := NewSendConfig(chainID)
	if err != nil {
		return nil, err
	}

	// Assemble and return
	return &Config{
		RpcClient:        rpcClient,
		EthClient:        ethClient,
		SenderKeys:       keys,
		SenderAddrs:      senderAddrs,
		Signer:           signer,
		DeploymentConfig: deploymentConfig,
		CallConfig:       callConfig,
		SendConfig:       sendConfig,
	}, nil
}

// NewDeploymentConfig constructs and returns a new DeploymentConfig
func NewDeploymentConfig(chainID *big.Int) (*DeploymentConfig, error) {
	hexData := viper.GetString(ethDeploymentData)
	data := common.Hex2Bytes(hexData)

	return &DeploymentConfig{
		ChainID:   chainID,
		Number:    viper.GetUint64(ethDeploymentNumber),
		Data:      data,
		GasLimit:  viper.GetUint64(ethDeploymentGasLimit),
		GasFeeCap: big.NewInt(viper.GetInt64(ethDeploymentGasFeeCap)),
		GasTipCap: big.NewInt(viper.GetInt64(ethDeploymentGasTipCap)),
	}, nil
}

// NewCallConfig constructs and returns a new CallConfig
func NewCallConfig(chainID *big.Int) (*CallConfig, error) {
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

	var frequency time.Duration
	tmpFreq := viper.GetInt(ethCallFrequency)
	if tmpFreq <= 0 {
		frequency = time.Microsecond
	} else {
		frequency = viper.GetDuration(ethCallFrequency) * time.Millisecond
	}

	return &CallConfig{
		ChainID:     chainID,
		GasLimit:    viper.GetUint64(ethCallGasLimit),
		GasFeeCap:   big.NewInt(viper.GetInt64(ethCallGasFeeCap)),
		GasTipCap:   big.NewInt(viper.GetInt64(ethCallGasTipCap)),
		MethodName:  methodName,
		ABI:         parsedABI,
		Frequency:   frequency,
		TotalNumber: viper.GetInt(ethCallTotalNumber),
	}, nil
}

// NewSendConfig constructs and returns a new SendConfig
func NewSendConfig(chainID *big.Int) (*SendConfig, error) {
	amountStr := viper.GetString(ethSendAmount)
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return nil, fmt.Errorf("unable to convert amount string (%s) into big.Int", amountStr)
	}

	var frequency time.Duration
	tmpFreq := viper.GetInt(ethCallFrequency)
	if tmpFreq <= 0 {
		frequency = time.Microsecond
	} else {
		frequency = viper.GetDuration(ethCallFrequency) * time.Millisecond
	}

	return &SendConfig{
		ChainID:     chainID,
		Frequency:   frequency,
		Amount:      amount,
		GasLimit:    viper.GetUint64(ethSendGasLimit),
		GasFeeCap:   big.NewInt(viper.GetInt64(ethSendGasFeeCap)),
		GasTipCap:   big.NewInt(viper.GetInt64(ethSendGasTipCap)),
		TotalNumber: viper.GetInt(ethSendTotalNumber),
	}, nil
}
