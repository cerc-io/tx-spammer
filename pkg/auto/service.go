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
	"github.com/sirupsen/logrus"
	"github.com/vulcanize/tx_spammer/pkg/shared"
)

// Spammer underlying struct type for spamming service
type Spammer struct {
	Deployer *ContractDeployer
	Sender   *EthSender
	TxGenerator *TxGenerator
}

// NewTxSpammer creates a new tx spamming service
func NewTxSpammer(config *Config) shared.Service {
	gen := NewTxGenerator(config)
	return &Spammer{
		Deployer: NewContractDeployer(config, gen),
		Sender:   NewEthSender(config),
		TxGenerator: gen,
	}
}

func (s *Spammer) Loop(quitChan <-chan bool) (<-chan bool, error) {
	if err := s.Deployer.Deploy(); err != nil {
		return nil, err
	}
	senderQuit := make(chan bool)
	generatorQuit := make(chan bool)
	genDoneChan, txRlpChan, genErrChan := s.TxGenerator.GenerateTxs(generatorQuit)

	doneChan, errChan := s.Sender.Send(senderQuit, txRlpChan)
	go func() {
		for {
			select {
			case <-genDoneChan:
				logrus.Info("all txs have been generated, beginning shut down sequence")
				senderQuit <- true
			case err := <-genErrChan:
				logrus.Error(err)
				senderQuit <- true
			case err := <-errChan:
				logrus.Error(err)
				senderQuit <- true // NOTE: sender will close doneChan when it receives a quit signal
			case <-quitChan:
				senderQuit <- true
			case <-doneChan: // NOTE CONT: which will be received here so that this context can close down only once the sender and generator have
				generatorQuit <- true
				return
			}
		}
	}()
	return doneChan, nil
}
