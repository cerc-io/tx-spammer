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

import "github.com/ethereum/go-ethereum/rpc"

// TxGenerator generates and signs txs
type TxGenerator struct {
	TxParams []TxParams
	currentTx []byte
	currentClient *rpc.Client
	err error
}

func NewTxGenerator(params []TxParams) *TxGenerator {
	return &TxGenerator{
		TxParams: params,
	}
}

func (gen TxGenerator) Next() bool {
	return false
}

func (gen TxGenerator) Current() (*rpc.Client, []byte) {
	return gen.currentClient, gen.currentTx
}

func (gen TxGenerator) Error() error {
	return gen.err
}

func (gen TxGenerator) gen(params TxParams) ([]byte, error) {
	return nil, nil
}