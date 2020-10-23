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

import "github.com/spf13/viper"

const (
	// env variables
	ETH_TX_LIST  = "ETH_TX_LIST"
	ETH_ADDR_LOG = "ETH_ADDR_LOG"

	// write paths
	defaultGenKeyWritePathPrefix = "./accounts/keys/out/"
	defaultAddrLogPath           = "./accounts/addresses/accounts"

	// .toml binding suffixes
	typeSuffix            = ".type"
	httpPathSuffix        = ".http"
	toSuffix              = ".to"
	amountSuffix          = ".amount"
	gasLimitSuffix        = ".gasLimit"
	gasPriceSuffix        = ".gasPrice"
	gasPremiumSuffix      = ".gasPremium"
	feeCapSuffix          = ".feeCap"
	dataSuffix            = ".data"
	senderKeyPathSuffix   = ".senderKeyPath"
	writeSenderPathSuffix = ".writeSenderPath"
	l1SenderSuffix        = ".l1Sender"
	l1RollupTxIdSuffix    = ".l1RollupTxId"
	sigHashTypeSuffix     = ".sigHashType"
	frequencySuffix       = ".frequency"
	totalNumberSuffix     = ".totalNumber"
	delaySuffix           = ".delay"
	startingNonceSuffix   = ".startingNonce"
	queueOriginSuffix     = ".queueOrigin"
	chainIDSuffix         = ".chainID"
	contractWriteSuffix   = ".writeDeploymentAddrPath"
)

func bindEnv() {
	viper.BindEnv("eth.txs", ETH_TX_LIST)
	viper.BindEnv("eth.addrLogPath", ETH_ADDR_LOG)
}
