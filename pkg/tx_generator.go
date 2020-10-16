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
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// TxGenerator generates and signs txs
type TxGenerator struct {
	// keep track of account nonces locally so we aren't spamming to determine the nonce
	// this assumes these accounts are not sending txs outside this process
	nonces map[common.Address]uint64
}

// NewTxGenerator creates a new tx generator
func NewTxGenerator(params []TxParams) *TxGenerator {
	nonces := make(map[common.Address]uint64)
	for _, p := range params {
		nonces[p.Sender] = p.StartingNonce
	}
	return &TxGenerator{
		nonces: nonces,
	}
}

// GenerateTx generates tx from the provided params
func (tg TxGenerator) GenerateTx(params TxParams) ([]byte, error) {
	tx := make([]byte, 0)
	switch params.Type {
	case Standard, OptimismL1ToL2, OptimismL2:
		return tg.gen(params)
	case EIP1559:
		return tg.gen1559(params)
	default:
		return nil, fmt.Errorf("unsupported tx type: %s", params.Type.String())
	}
	return tx, nil
}

func (gen TxGenerator) gen(params TxParams) ([]byte, error) {
	nonce := gen.nonces[params.Sender]
	tx := new(types.Transaction)
	if params.To == nil {
		tx = types.NewContractCreation(nonce, params.Amount, params.GasLimit, params.GasPrice, params.Data, params.L1SenderAddr, params.L1RollupTxId, params.QueueOrigin)
	} else {
		tx = types.NewTransaction(nonce, *params.To, params.Amount, params.GasLimit, params.GasPrice, params.Data, params.L1SenderAddr, params.L1RollupTxId, params.QueueOrigin, params.SigHashType)
	}
	signer, err := TxSigner(params.Type, params.ChainID)
	if err != nil {
		return nil, err
	}
	signedTx, err := types.SignTx(tx, signer, params.SenderKey)
	if err != nil {
		return nil, err
	}
	txRlp, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, err
	}
	gen.nonces[params.Sender]++
	return txRlp, nil
}

func (gen TxGenerator) gen1559(params TxParams) ([]byte, error) {
	// TODO: support EIP1559; new to make a new major version, vendor it, or release with different pkg name so that we can import both optimism and eip1559 geth
	return nil, fmt.Errorf("1559 support not yet available")
}
