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
	"github.com/sirupsen/logrus"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/cerc-io/tx-spammer/pkg/shared"
)

const (
	contractDeploymentDelay = time.Duration(15) * time.Second
)

// ContractDeployer is responsible for deploying contracts
type ContractDeployer struct {
	client      *rpc.Client
	txGenerator *TxGenerator
	senderKeys  []*ecdsa.PrivateKey
	senderAddrs []common.Address
	config      *DeploymentConfig
}

// NewContractDeployer returns a new ContractDeployer
func NewContractDeployer(config *Config, gen *TxGenerator) *ContractDeployer {
	return &ContractDeployer{
		client:      config.RpcClient,
		txGenerator: gen,
		config:      config.DeploymentConfig,
		senderKeys:  config.SenderKeys,
		senderAddrs: config.SenderAddrs,
	}
}

// Deploy deploys the contracts according to the config provided at construction
func (cp *ContractDeployer) Deploy() ([]common.Address, error) {
	contractAddrs := make([]common.Address, 0, cp.config.Number*uint64(len(cp.senderKeys)))
	ticker := time.NewTicker(contractDeploymentDelay)
	defer ticker.Stop()
	for i := uint64(0); i < cp.config.Number; i++ {
		<-ticker.C
		for i, key := range cp.senderKeys {
			logrus.Debugf("Generating contract deployment for %s.", cp.senderAddrs[i].Hex())
			signedTx, contractAddr, err := cp.txGenerator.GenerateTx(&GenParams{
				ChainID:   cp.config.ChainID,
				Sender:    cp.senderAddrs[i],
				SenderKey: key,
				GasLimit:  cp.config.GasLimit,
				GasFeeCap: cp.config.GasFeeCap,
				GasTipCap: cp.config.GasTipCap,
				Data:      cp.config.Data,
			})
			if err != nil {
				return nil, err
			}
			if err := shared.SendTransaction(cp.client, signedTx); err != nil {
				return nil, err
			}
			contractAddrs = append(contractAddrs, contractAddr)
		}
	}
	return contractAddrs, nil
}
