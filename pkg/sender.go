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
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

type TxSender struct {
	TxGen *TxGenerator
}

func NewTxSender(params []TxParams) *TxSender {
	return &TxSender{
		TxGen: NewTxGenerator(params),
	}
}
func (s *TxSender) Send(quitChan <-chan bool) <-chan error {
	errChan := make(chan error)
	go func() {
		for s.TxGen.Next() {
			select {
			case <-quitChan:
				return
			default:
			}
			if err := sendRawTransaction(s.TxGen.Current()); err != nil {
				errChan <- err
			}
		}
		if s.TxGen.Error() != nil {
			errChan <- s.TxGen.Error()
		}
	}()
	return errChan
}

func sendRawTransaction(rpcClient *rpc.Client, txRlp []byte) error {
	return rpcClient.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(txRlp))
}
