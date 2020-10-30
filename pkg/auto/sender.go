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
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/vulcanize/tx_spammer/pkg/shared"
)

// EthSender sends eth value transfer txs
type EthSender struct {
	client *rpc.Client
}

// NewEthSender returns a new EthSender
func NewEthSender(config *Config) *EthSender {
	return &EthSender{
		client: config.Client,
	}
}

// Send awaits txs off the provided work queue and sends them
func (s *EthSender) Send(quitChan <-chan bool, txRlpChan <-chan []byte) (<-chan bool, <-chan error) {
	// err channel returned to calling context
	errChan := make(chan error)
	doneChan := make(chan bool)
	go func() {
		defer close(doneChan)
		for {
			select {
			case tx := <-txRlpChan:
				if err := shared.SendRawTransaction(s.client, tx); err != nil {
					errChan <- err
				}
			case <-quitChan:
				logrus.Info("quitting Send loop")
				return
			}
		}
	}()
	return doneChan, errChan
}
