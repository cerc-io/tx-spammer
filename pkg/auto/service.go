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
	"fmt"

	"github.com/cerc-io/tx-spammer/pkg/shared"
	"github.com/sirupsen/logrus"
)

// Spammer underlying struct type for spamming service
type Spammer struct {
	Deployer    *ContractDeployer
	Sender      *EthSender
	TxGenerator *TxGenerator
	config      *Config
}

// NewTxSpammer creates a new tx spamming service
func NewTxSpammer(config *Config) shared.Service {
	gen := NewTxGenerator(config)
	return &Spammer{
		Deployer:    NewContractDeployer(config, gen),
		Sender:      NewEthSender(config),
		TxGenerator: gen,
		config:      config,
	}
}

func (s *Spammer) Loop(quitChan <-chan bool) (<-chan bool, error) {
	contractAddrs, err := s.Deployer.Deploy()
	if err != nil {
		return nil, fmt.Errorf("contract deployment error: %v", err)
	}
	genQuit := make(chan bool)
	senderQuit := make(chan bool)
	doneChan := make(chan bool)
	watcher := NewTxWatcher(s.config.EthClient)
	watcher.Start()

	s.config.CallConfig.ContractAddrs = contractAddrs
	genDoneChan, txChan, genErrChan := s.TxGenerator.GenerateTxs(genQuit)
	sendDoneChan, sendErrChan := s.Sender.Send(senderQuit, txChan, watcher.PendingTxCh)

	go func() {
		<-genDoneChan
		recoverClose(senderQuit)
		<-sendDoneChan
		recoverClose(watcher.quitCh)
		close(doneChan)
	}()

	go func() {
		for err := range genErrChan {
			logrus.Errorf("tx generation error: %v", err)
			recoverClose(genQuit)
		}
	}()

	go func() {
		for err := range sendErrChan {
			logrus.Errorf("tx sending error: %v", err)
			if s.config.StopOnSendError {
				recoverClose(genQuit)
			}
		}
	}()

	go func() {
		<-quitChan
		logrus.Info("shutting down tx spammer")
		recoverClose(genQuit)
	}()

	return doneChan, nil
}

func recoverClose(ch chan bool) (justClosed bool) {
	defer func() {
		if recover() != nil {
			justClosed = false
		}
	}()

	close(ch)
	return true
}
