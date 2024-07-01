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
	"errors"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/holiman/uint256"
	log "github.com/sirupsen/logrus"

	"github.com/cerc-io/tx-spammer/pkg/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	BlobParams *BlobParams
}

type BlobParams struct {
	BlobFeeCap *uint256.Int
	BlobHashes []common.Hash
	Sidecar    *types.BlobTxSidecar
}

func (gen *TxGenerator) GenerateTxs(quitChan <-chan bool) (<-chan bool, <-chan *types.Transaction, <-chan error) {
	txChan := make(chan *types.Transaction)
	errChan := make(chan error)
	wg := new(sync.WaitGroup)
	for i, sender := range gen.config.SenderKeys {
		if gen.config.SendConfig.TotalNumber > 0 {
			wg.Add(1)
			go gen.genSends(wg, txChan, errChan, quitChan, sender, gen.config.SenderAddrs[i], gen.config.SendConfig)
		}
		if gen.config.CallConfig.TotalNumber > 0 {
			wg.Add(1)
			go gen.genCalls(wg, txChan, errChan, quitChan, sender, gen.config.SenderAddrs[i], gen.config.CallConfig)
		}
		if gen.config.BlobTxConfig.TotalNumber > 0 {
			wg.Add(1)
			go gen.genBlobTx(senderArgs{
				wg:         wg,
				txChan:     txChan,
				errChan:    errChan,
				quitChan:   quitChan,
				senderKey:  sender,
				senderAddr: gen.config.SenderAddrs[i],
			}, gen.config.BlobTxConfig)
		}
	}
	doneChan := make(chan bool)
	go func() {
		wg.Wait()
		close(errChan)
		close(doneChan)
	}()
	return doneChan, txChan, errChan
}

func (gen *TxGenerator) genSends(wg *sync.WaitGroup, txChan chan<- *types.Transaction, errChan chan<- error, quitChan <-chan bool, senderKey *ecdsa.PrivateKey, senderAddr common.Address, sendConfig *SendConfig) {
	defer wg.Done()
	ticker := time.NewTicker(sendConfig.Frequency)
	for i := 0; i < sendConfig.TotalNumber; i++ {
		select {
		case <-ticker.C:
			dst := crypto.CreateAddress(receiverAddressSeed, uint64(i))
			log.Debugf("Generating send from %s to %s.", senderAddr.Hex(), dst.Hex())
			params := &GenParams{
				ChainID:   sendConfig.ChainID,
				To:        &dst,
				Sender:    senderAddr,
				SenderKey: senderKey,
				GasLimit:  sendConfig.GasLimit,
				GasFeeCap: sendConfig.GasFeeCap,
				GasTipCap: sendConfig.GasTipCap,
				Amount:    sendConfig.Amount,
			}
			rawTx, _, err := gen.createTx(params)
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
			data, err := callConfig.ABI.Pack(callConfig.MethodName, contractAddr, big.NewInt(int64(i)))
			if err != nil {
				errChan <- err
				continue
			}
			rawTx, _, err := gen.createTx(&GenParams{
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

type senderArgs struct {
	wg         *sync.WaitGroup
	txChan     chan<- *types.Transaction
	errChan    chan<- error
	quitChan   <-chan bool
	senderKey  *ecdsa.PrivateKey
	senderAddr common.Address
}

func (gen *TxGenerator) genBlobTx(args senderArgs, blobTxConfig *BlobTxConfig) {
	defer args.wg.Done()
	ticker := time.NewTicker(blobTxConfig.Frequency)
	for i := 0; i < blobTxConfig.TotalNumber; i++ {
		select {
		case <-ticker.C:
			dst := crypto.CreateAddress(receiverAddressSeed, uint64(i))
			log.Debugf("Generating send from %s to %s.", args.senderAddr, dst)
			params := &GenParams{
				ChainID:   blobTxConfig.ChainID,
				To:        &dst,
				Sender:    args.senderAddr,
				SenderKey: args.senderKey,
				GasLimit:  blobTxConfig.GasLimit,
				GasFeeCap: blobTxConfig.GasFeeCap,
				GasTipCap: blobTxConfig.GasTipCap,
				Amount:    blobTxConfig.Amount,
			}
			blobdata := make([]byte, blobTxConfig.BlobCount)
			for i := range blobdata {
				blobdata[i] = byte(i + 1)
			}
			sidecar, err := makeSidecar(blobdata)
			if err != nil {
				args.errChan <- err
				continue
			}
			params.BlobParams = &BlobParams{
				BlobFeeCap: blobTxConfig.BlobFeeCap,
				BlobHashes: sidecar.BlobHashes(),
				Sidecar:    sidecar,
			}

			rawTx, _, err := gen.createTx(params)
			if err != nil {
				args.errChan <- err
				continue
			}
			args.txChan <- rawTx
		case <-args.quitChan:
			return
		}
	}
	log.Info("Done generating sends for ", args.senderAddr)
}

// createTx generates tx from the provided params
func (gen *TxGenerator) createTx(params *GenParams) (*types.Transaction, common.Address, error) {
	nonce := gen.claimNonce(params.Sender)
	tx := new(types.Transaction)
	var contractAddr common.Address
	var err error
	if params.To == nil {
		if params.BlobParams != nil {
			return nil, common.Address{}, errors.New("BlobTx cannot be used for contract creation")
		}
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
	} else if params.BlobParams == nil {
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
	} else {
		tx = types.NewTx(
			&types.BlobTx{
				ChainID:    uint256.MustFromBig(params.ChainID),
				Nonce:      nonce,
				Gas:        params.GasLimit,
				GasTipCap:  uint256.MustFromBig(params.GasTipCap),
				GasFeeCap:  uint256.MustFromBig(params.GasFeeCap),
				To:         *params.To,
				Value:      uint256.MustFromBig(params.Amount),
				Data:       params.Data,
				BlobFeeCap: params.BlobParams.BlobFeeCap,
				BlobHashes: params.BlobParams.BlobHashes,
				Sidecar:    params.BlobParams.Sidecar,
			})
	}

	signedTx, err := types.SignTx(tx, gen.config.Signer, params.SenderKey)
	if err != nil {
		return nil, common.Address{}, err
	}
	return signedTx, contractAddr, err
}

// From go-ethereum/cmd/devp2p/internal/ethtest/suite.go
func makeSidecar(data []byte) (*types.BlobTxSidecar, error) {
	var (
		blobs       = make([]kzg4844.Blob, len(data))
		commitments []kzg4844.Commitment
		proofs      []kzg4844.Proof
	)
	for i := range blobs {
		blobs[i][0] = data[i]
		c, err := kzg4844.BlobToCommitment(&blobs[i])
		if err != nil {
			return nil, err
		}
		p, err := kzg4844.ComputeBlobProof(&blobs[i], c)
		if err != nil {
			return nil, err
		}
		commitments = append(commitments, c)
		proofs = append(proofs, p)
	}
	return &types.BlobTxSidecar{
		Blobs:       blobs,
		Commitments: commitments,
		Proofs:      proofs,
	}, nil
}
