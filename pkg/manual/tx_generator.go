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
	"sync/atomic"

	"github.com/cerc-io/tx-spammer/pkg/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// TxGenerator generates and signs txs
type TxGenerator struct {
	// keep track of account nonces locally so we aren't spamming to determine the nonce
	// this assumes these accounts are not sending txs outside this process
	nonces map[common.Address]*uint64
}

// NewTxGenerator creates a new tx generator
func NewTxGenerator(params []TxParams) *TxGenerator {
	nonces := make(map[common.Address]*uint64)
	for _, p := range params {
		nonces[p.Sender] = &p.StartingNonce
	}
	return &TxGenerator{
		nonces: nonces,
	}
}

// GenerateTx generates tx from the provided params
func (gen TxGenerator) GenerateTx(params TxParams) ([]byte, error) {
	nonce := atomic.AddUint64(gen.nonces[params.Sender], 1)
	tx := new(types.Transaction)
	if params.To == nil {
		tx = types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   params.ChainID,
				Nonce:     nonce,
				Gas:       params.Gas,
				GasTipCap: params.GasTipCap,
				GasFeeCap: params.GasFeeCap,
				To:        nil,
				Value:     params.Value,
				Data:      params.Data,
			})
		if _, err := shared.WriteContractAddr(params.ContractAddrWritePath, params.Sender, nonce); err != nil {
			return nil, err
		}

	} else {
		tx = types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   params.ChainID,
				Nonce:     nonce,
				Gas:       params.Gas,
				GasTipCap: params.GasTipCap,
				GasFeeCap: params.GasFeeCap,
				To:        params.To,
				Value:     params.Value,
				Data:      params.Data,
			})
	}
	signer := shared.TxSigner(params.ChainID)
	signedTx, err := types.SignTx(tx, signer, params.SenderKey)
	if err != nil {
		return nil, err
	}
	rawTx, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, err
	}
	return rawTx, nil
}
