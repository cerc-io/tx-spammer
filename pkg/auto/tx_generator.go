// VulcanizeDB
// Copyright © 2020 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more detailgen.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package auto

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vulcanize/tx_spammer/pkg/shared"
)

// TxGenerator generates and signs txs
type TxGenerator struct {
	config *Config
	// keep track of account nonces locally so we aren't spamming to determine the nonce
	// this assumes these accounts are not sending txs outside this process
	nonces map[common.Address]*uint64
}

// NewTxGenerator creates a new tx generator
func NewTxGenerator(config *Config) *TxGenerator {
	nonces := make(map[common.Address]*uint64)
	for _, addr := range config.SenderAddrs {
		startingNonce := uint64(0)
		nonces[addr] = &startingNonce
	}
	return &TxGenerator{
		nonces: nonces,
		config: config,
	}
}

// GenParams params for GenerateTx method calls
type GenParams struct {
	Sender    common.Address
	SenderKey *ecdsa.PrivateKey
	To        *common.Address
	Amount    *big.Int
	GasLimit  uint64
	GasPrice  *big.Int
	Data      []byte
}

func (gen *TxGenerator) GenerateTxs(quitChan <-chan bool, contractAddrs []common.Address) (<-chan bool, <-chan []byte, <-chan error) {
	txRlpChan := make(chan []byte)
	errChan := make(chan error)
	wg := new(sync.WaitGroup)
	for i, sender := range gen.config.SenderKeys {
		if len(gen.config.SendConfig.DestinationAddresses) > 0 {
			wg.Add(1)
			go gen.genSends(wg, gen.config.Type, txRlpChan, errChan, quitChan, sender, gen.config.SenderAddrs[i], gen.config.SendConfig)
		}
		if len(gen.config.CallConfig.StorageAddrs) > 0 {
			wg.Add(1)
			go gen.genCalls(wg, gen.config.Type, txRlpChan, errChan, quitChan, sender, gen.config.SenderAddrs[i], gen.config.CallConfig)
		}
	}
	doneChan := make(chan bool)
	go func() {
		wg.Wait()
		close(doneChan)
	}()
	return doneChan, txRlpChan, errChan
}

func (gen *TxGenerator) genSends(wg *sync.WaitGroup, ty shared.TxType, txRlpChan chan<- []byte, errChan chan<- error, quitChan <-chan bool, senderKey *ecdsa.PrivateKey, senderAddr common.Address, sendConfig *SendConfig) {
	defer wg.Done()
	ticker := time.NewTicker(sendConfig.Frequency)
	for _, dst := range sendConfig.DestinationAddresses {
		select {
		case <-ticker.C:
			txRlp, _, err := gen.GenerateTx(ty, &GenParams{
				Sender:    senderAddr,
				SenderKey: senderKey,
				GasLimit:  sendConfig.GasLimit,
				GasPrice:  sendConfig.GasPrice,
				Amount:    sendConfig.Amount,
				To:        &dst,
			})
			if err != nil {
				errChan <- err
				continue
			}
			txRlpChan <- txRlp
		case <-quitChan:
			return
		}
	}
}

func (gen *TxGenerator) genCalls(wg *sync.WaitGroup, ty shared.TxType, txRlpChan chan<- []byte, errChan chan<- error, quitChan <-chan bool, senderKey *ecdsa.PrivateKey, senderAddr common.Address, callConfig *CallConfig) {
	defer wg.Done()
	ticker := time.NewTicker(callConfig.Frequency)
	for _, addr := range callConfig.StorageAddrs {
		select {
		case <-ticker.C:
			data, err := callConfig.ABI.Pack(callConfig.MethodName, addr, callConfig.StorageValue)
			if err != nil {
				errChan <- err
				continue
			}
			txRlp, _, err := gen.GenerateTx(ty, &GenParams{
				Sender:    senderAddr,
				SenderKey: senderKey,
				GasLimit:  callConfig.GasLimit,
				GasPrice:  callConfig.GasPrice,
				Data:      data,
			})
			if err != nil {
				errChan <- err
				continue
			}
			txRlpChan <- txRlp
		case <-quitChan:
			return
		}
	}
}

// GenerateTx generates tx from the provided params
func (gen TxGenerator) GenerateTx(ty shared.TxType, params *GenParams) ([]byte, common.Address, error) {
	switch ty {
	case shared.OptimismL2:
		return gen.genL2(params, gen.config.OptimismConfig)
	case shared.Standard:
		return gen.gen(params)
	case shared.EIP1559:
		return gen.gen1559(params, gen.config.EIP1559Config)
	default:
		return nil, common.Address{}, fmt.Errorf("unsupported tx type: %s", ty.String())
	}
}

func (gen TxGenerator) genL2(params *GenParams, op *OptimismConfig) ([]byte, common.Address, error) {
	nonce := atomic.AddUint64(gen.nonces[params.Sender], 1)
	tx := new(types.Transaction)
	var contractAddr common.Address
	var err error
	if params.To == nil {
		tx = types.NewContractCreation(nonce, params.Amount, params.GasLimit, params.GasPrice, params.Data, op.L1SenderAddr, op.L1RollupTxId, op.QueueOrigin)
		contractAddr, err = shared.WriteContractAddr(shared.DefaultDeploymentAddrLogPathPrefix, params.Sender, nonce)
		if err != nil {
			return nil, common.Address{}, err
		}
	} else {
		tx = types.NewTransaction(nonce, *params.To, params.Amount, params.GasLimit, params.GasPrice, params.Data, op.L1SenderAddr, op.L1RollupTxId, op.QueueOrigin, op.SigHashType)
	}
	signedTx, err := types.SignTx(tx, gen.config.Signer, params.SenderKey)
	if err != nil {
		return nil, common.Address{}, err
	}
	txRlp, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, common.Address{}, err
	}
	return txRlp, contractAddr, err
}

func (gen TxGenerator) gen(params *GenParams) ([]byte, common.Address, error) {
	// TODO: support standard geth
	return nil, common.Address{}, errors.New("L1 support not yet available")
}

func (gen TxGenerator) gen1559(params *GenParams, eip1559Config *EIP1559Config) ([]byte, common.Address, error) {
	// TODO: support EIP1559
	return nil, common.Address{}, errors.New("1559 support not yet available")
}
