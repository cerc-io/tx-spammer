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
	"context"
	"crypto/ecdsa"
	log "github.com/sirupsen/logrus"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/vulcanize/tx_spammer/pkg/shared"
)

// TxGenerator generates and signs txs
type TxGenerator struct {
	config *Config
	// keep track of account nonces locally so we aren't spamming to determine the nonce
	// this assumes these accounts are not sending txs outside this process
	nonces map[common.Address]uint64
	lock   sync.Mutex
}

func (gen *TxGenerator) claimNonce(addr common.Address) uint64 {
	gen.lock.Lock()
	ret := gen.nonces[addr]
	gen.nonces[addr] += 1
	gen.lock.Unlock()
	return ret
}

// NewTxGenerator creates a new tx generator
func NewTxGenerator(config *Config) *TxGenerator {
	nonces := make(map[common.Address]uint64)
	for _, addr := range config.SenderAddrs {
		nonce, _ := config.EthClient.PendingNonceAt(context.Background(), addr)
		nonces[addr] = nonce
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

	ChainID   *big.Int
	GasTipCap *big.Int
	GasFeeCap *big.Int
	GasLimit  uint64
	To        *common.Address
	Amount    *big.Int
	Data      []byte
}

func (gen *TxGenerator) GenerateTxs(quitChan <-chan bool) (<-chan bool, <-chan *types.Transaction, <-chan error) {
	txChan := make(chan *types.Transaction)
	errChan := make(chan error)
	wg := new(sync.WaitGroup)
	for i, sender := range gen.config.SenderKeys {
		if len(gen.config.SendConfig.DestinationAddresses) > 0 {
			wg.Add(1)
			go gen.genSends(wg, txChan, errChan, quitChan, sender, gen.config.SenderAddrs[i], gen.config.SendConfig)
		}
		if gen.config.CallConfig.TotalNumber > 0 {
			wg.Add(1)
			go gen.genCalls(wg, txChan, errChan, quitChan, sender, gen.config.SenderAddrs[i], gen.config.CallConfig)
		}
	}
	doneChan := make(chan bool)
	go func() {
		wg.Wait()
		close(doneChan)
	}()
	return doneChan, txChan, errChan
}

func (gen *TxGenerator) genSends(wg *sync.WaitGroup, txChan chan<- *types.Transaction, errChan chan<- error, quitChan <-chan bool, senderKey *ecdsa.PrivateKey, senderAddr common.Address, sendConfig *SendConfig) {
	defer wg.Done()
	ticker := time.NewTicker(sendConfig.Frequency)
	for _, dst := range sendConfig.DestinationAddresses {
		select {
		case <-ticker.C:
			log.Debugf("Generating send from %s to %s.", senderAddr.Hex(), dst.Hex())
			rawTx, _, err := gen.GenerateTx(&GenParams{
				ChainID:   sendConfig.ChainID,
				To:        &dst,
				Sender:    senderAddr,
				SenderKey: senderKey,
				GasLimit:  sendConfig.GasLimit,
				GasFeeCap: sendConfig.GasFeeCap,
				GasTipCap: sendConfig.GasTipCap,
				Amount:    sendConfig.Amount,
			})
			if err != nil {
				errChan <- err
				continue
			}
			txChan <- rawTx
		case <-quitChan:
			return
		}
	}
	log.Info("Done generating sends for ", senderAddr.Hex())
}

func (gen *TxGenerator) genCalls(wg *sync.WaitGroup, txChan chan<- *types.Transaction, errChan chan<- error, quitChan <-chan bool, senderKey *ecdsa.PrivateKey, senderAddr common.Address, callConfig *CallConfig) {
	defer wg.Done()
	ticker := time.NewTicker(callConfig.Frequency)
	for i := 0; i < callConfig.TotalNumber; i++ {
		select {
		case <-ticker.C:
			contractAddr := callConfig.ContractAddrs[rand.Intn(len(callConfig.ContractAddrs))]
			log.Debugf("Generating call from %s to %s.", senderAddr.Hex(), contractAddr.Hex())
			data, err := callConfig.ABI.Pack(callConfig.MethodName, contractAddr, big.NewInt(time.Now().UnixNano()))
			if err != nil {
				errChan <- err
				continue
			}
			rawTx, _, err := gen.GenerateTx(&GenParams{
				Sender:    senderAddr,
				SenderKey: senderKey,
				GasLimit:  callConfig.GasLimit,
				GasFeeCap: callConfig.GasFeeCap,
				GasTipCap: callConfig.GasTipCap,
				Data:      data,
				To:        &contractAddr,
			})
			if err != nil {
				errChan <- err
				continue
			}
			txChan <- rawTx
		case <-quitChan:
			return
		}
	}
	log.Info("Done generating calls for ", senderAddr.Hex())
}

// GenerateTx generates tx from the provided params
func (gen *TxGenerator) GenerateTx(params *GenParams) (*types.Transaction, common.Address, error) {
	nonce := gen.claimNonce(params.Sender)
	tx := new(types.Transaction)
	var contractAddr common.Address
	var err error
	if params.To == nil {
		tx = types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   params.ChainID,
				Nonce:     nonce,
				Gas:       params.GasLimit,
				GasTipCap: params.GasTipCap,
				GasFeeCap: params.GasFeeCap,
				To:        nil,
				Value:     params.Amount,
				Data:      params.Data,
			})
		contractAddr, err = shared.WriteContractAddr("", params.Sender, nonce)
		if err != nil {
			return nil, common.Address{}, err
		}
	} else {
		tx = types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   params.ChainID,
				Nonce:     nonce,
				GasTipCap: params.GasTipCap,
				GasFeeCap: params.GasFeeCap,
				Gas:       params.GasLimit,
				To:        params.To,
				Value:     params.Amount,
				Data:      params.Data,
			})
	}
	signedTx, err := types.SignTx(tx, gen.config.Signer, params.SenderKey)
	if err != nil {
		return nil, common.Address{}, err
	}
	return signedTx, contractAddr, err
}
