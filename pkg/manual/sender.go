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
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vulcanize/tx_spammer/pkg/shared"
)

// TxSender type for
type TxSender struct {
	TxGen    *TxGenerator
	TxParams []TxParams
}

// NewTxSender returns a new tx sender
func NewTxSender(params []TxParams) *TxSender {
	return &TxSender{
		TxGen:    NewTxGenerator(params),
		TxParams: params,
	}
}

func (s *TxSender) Send(quitChan <-chan bool) (<-chan bool, <-chan error) {
	// done channel to signal completion of all jobs
	doneChan := make(chan bool)
	// err channel returned to calling context
	errChan := make(chan error)
	// for each tx param set, spin up a goroutine to generate and send the tx at the specified delay and frequency
	wg := new(sync.WaitGroup)
	for _, txParams := range s.TxParams {
		wg.Add(1)
		go func(p TxParams) {
			defer wg.Done()
			// send the first tx after the delay
			timer := time.NewTimer(p.Delay)
			<-timer.C
			if err := s.genAndSend(p); err != nil {
				errChan <- fmt.Errorf("tx %s initial genAndSend error: %v", p.Name, err)
				return
			}
			// send any remaining ones at the provided frequency, also check for quit signal
			ticker := time.NewTicker(p.Frequency)
			for i := uint64(1); i < p.TotalNumber; i++ {
				select {
				case <-ticker.C:
					if err := s.genAndSend(p); err != nil {
						errChan <- fmt.Errorf("tx %s number %d genAndSend error: %v", p.Name, i, err)
						return
					}
				case <-quitChan:
					return
				}
			}
		}(txParams)
	}
	go func() {
		wg.Wait()
		close(doneChan)
	}()
	return doneChan, errChan
}

func (s *TxSender) genAndSend(p TxParams) error {
	tx, err := s.TxGen.GenerateTx(p)
	if err != nil {
		return err
	}
	logrus.Infof("sending tx %s", p.Name)
	return shared.SendRawTransaction(p.Client, tx)
}
