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
	Caller   *ContractCaller
	Sender   *EthSender
}

// NewTxSpammer creates a new tx spamming service
func NewTxSpammer(config *Config) shared.Service {
	return &Spammer{
		Deployer: NewContractDeployer(config),
		Caller:   NewContractCaller(config),
		Sender:   NewEthSender(config),
	}
}

func (s *Spammer) Loop(quitChan <-chan bool) <-chan bool {
	forwardQuit := make(chan bool)
	doneChan, errChan := s.Sender.Send(forwardQuit)
	go func() {
		for {
			select {
			case err := <-errChan:
				logrus.Error(err)
			case forwardQuit <- <-quitChan:
				return
			case <-doneChan:
				return
			}
		}
	}()
	return doneChan
}
