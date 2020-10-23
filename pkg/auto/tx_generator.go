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
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vulcanize/tx_spammer/pkg/shared"
)

// TxGenerator generates and signs txs
type TxGenerator struct {
	signer         types.Signer
	optimismConfig *OptimismConfig
	eip1559Config  *EIP1559Config
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
		signer:         config.Signer,
		nonces:         nonces,
		optimismConfig: config.OptimismConfig,
		eip1559Config:  config.EIP1559Config,
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

// GenerateTxs loops and generates txs according the configuration passed in during construction
func (tg TxGenerator) GenerateTxs(quitChan <-chan bool) (<-chan bool, <-chan []byte, <-chan error) {
	// TODO: this
	return nil, nil, nil
}

// GenerateTx generates tx from the provided params
func (tg TxGenerator) GenerateTx(ty shared.TxType, params *GenParams) ([]byte, error) {
	tx := make([]byte, 0)
	switch ty {
	case shared.OptimismL2:
		return tg.genL2(params, tg.optimismConfig)
	case shared.Standard:
		return tg.gen(params)
	case shared.EIP1559:
		return tg.gen1559(params, tg.eip1559Config)
	default:
		return nil, fmt.Errorf("unsupported tx type: %s", ty.String())
	}
	return tx, nil
}

func (gen TxGenerator) genL2(params *GenParams, op *OptimismConfig) ([]byte, error) {
	nonce := atomic.AddUint64(gen.nonces[params.Sender], 1)
	tx := new(types.Transaction)
	if params.To == nil {
		tx = types.NewContractCreation(nonce, params.Amount, params.GasLimit, params.GasPrice, params.Data, op.L1SenderAddr, op.L1RollupTxId, op.QueueOrigin)
		if err := shared.WriteContractAddr(shared.DefaultDeploymentAddrLogPathPrefix, params.Sender, nonce); err != nil {
			return nil, err
		}

	} else {
		tx = types.NewTransaction(nonce, *params.To, params.Amount, params.GasLimit, params.GasPrice, params.Data, op.L1SenderAddr, op.L1RollupTxId, op.QueueOrigin, op.SigHashType)
	}
	signedTx, err := types.SignTx(tx, gen.signer, params.SenderKey)
	if err != nil {
		return nil, err
	}
	txRlp, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, err
	}
	return txRlp, nil
}

func (gen TxGenerator) gen(params *GenParams) ([]byte, error) {
	// TODO: support standard geth
	return nil, errors.New("L1 support not yet available")
}

func (gen TxGenerator) gen1559(params *GenParams, eip1559Config *EIP1559Config) ([]byte, error) {
	// TODO: support EIP1559
	return nil, errors.New("1559 support not yet available")
}
