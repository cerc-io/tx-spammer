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

import "github.com/spf13/viper"

const (
	// env variables
	ETH_KEY_DIR_PATH  = "ETH_KEY_DIR_PATH"
	ETH_ADDR_DIR_PATH = "ETH_ADDR_DIR_PATH"
	ETH_HTTP_PATH     = "ETH_HTTP_PATH"
	ETH_CHAIN_ID      = "ETH_CHAIN_ID"
	ETH_TX_TYPE       = "ETH_TX_TYPE"

	ETH_DEPLOYMENT_NUMBER    = "ETH_DEPLOYMENT_NUMBER"
	ETH_DEPLOYMENT_HEX_DATA  = "ETH_DEPLOYMENT_HEX_DATA"
	ETH_DEPLOYMENT_GAS_LIMIT = "ETH_DEPLOYMENT_GAS_LIMIT"
	ETH_DEPLOYMENT_GAS_PRICE = "ETH_DEPLOYMENT_GAS_PRICE"

	ETH_OPTIMISM_L1_SENDER     = "ETH_OPTIMISM_L1_SENDER"
	ETH_OPTIMISM_ROLLUP_TX_ID  = "ETH_OPTIMISM_ROLLUP_TX_ID"
	ETH_OPTIMISM_SIG_HASH_TYPE = "ETH_OPTIMISM_SIG_HASH_TYPE"
	ETH_OPTIMISM_QUEUE_ORIGIN  = "ETH_OPTIMISM_QUEUE_ORIGIN"

	ETH_CALL_FREQ           = "ETH_CALL_FREQ"
	ETH_CALL_TOTAL_NUMBER   = "ETH_CALL_TOTAL_NUMBER"
	ETH_CALL_ABI_PATH       = "ETH_CALL_ABI_PATH"
	ETH_CALL_METHOD_NAME    = "ETH_CALL_METHOD_NAME"
	ETH_CALL_POSITION_START = "ETH_CALL_POSITION_START"
	ETH_CALL_POSITION_END   = "ETH_CALL_POSITION_END"
	ETH_CALL_STORAGE_VALUE  = "ETH_CALL_STORAGE_VALUE"
	ETH_CALL_GAS_LIMIT      = "ETH_CALL_GAS_LIMIT"
	ETH_CALL_GAS_PRICE      = "ETH_CALL_GAS_PRICE"

	ETH_SEND_FREQ         = "ETH_SEND_FREQ"
	ETH_SEND_TOTAL_NUMBER = "ETH_SEND_TOTAL_NUMBER"
	ETH_SEND_AMOUNT       = "ETH_SEND_AMOUNT"
	ETH_SEND_GAS_LIMIT    = "ETH_SEND_GAS_LIMIT"
	ETH_SEND_GAS_PRICE    = "ETH_SEND_GAS_PRICE"

	// toml bindings
	ethKeyDirPath   = "eth.keyDirPath"
	ethAddrFilePath = "eth.addrFilePath"
	ethHttpPath     = "eth.httpPath"
	ethChainID      = "eth.chainID"
	ethType         = "eth.type"

	ethDeploymentNumber   = "deployment.number"
	ethDeploymentData     = "deployment.hexData"
	ethDeploymentGasPrice = "deployment.gasPrice"
	ethDeploymentGasLimit = "deployment.gasLimit"

	ethOptimismL1Sender    = "optimism.l1Sender"
	ethOptimismRollupTxID  = "optimism.l1RollupTxId"
	ethOptimismSigHashType = "optimism.sigHashType"
	ethOptimismQueueOrigin = "optimism.queueOrigin"

	ethCallFrequency     = "contractSpammer.frequency"
	ethCallTotalNumber   = "contractSpammer.totalNumber"
	ethCallABIPath       = "contractSpammer.abiPath"
	ethCallMethodName    = "contractSpammer.methodName"
	ethCallPositionStart = "contractSpammer.positionStart"
	ethCallPositionEnd   = "contractSpammer.positionEnd"
	ethCallStorageValue  = "contractSpammer.storageValue"
	ethCallGasLimit      = "contractSpammer.gasLimit"
	ethCallGasPrice      = "contractSpammer.gasPrice"

	ethSendFrequency   = "sendSpammer.frequency"
	ethSendTotalNumber = "sendSpammer.totalNumber"
	ethSendAmount      = "sendSpammer.amount"
	ethSendGasLimit    = "sendSpammer.gasLimit"
	ethSendGasPrice    = "sendSpammer.gasPrice"
)

func bindEnv() {
	viper.BindEnv(ethKeyDirPath, ETH_KEY_DIR_PATH)
	viper.BindEnv(ethAddrFilePath, ETH_ADDR_DIR_PATH)
	viper.BindEnv(ethHttpPath, ETH_HTTP_PATH)
	viper.BindEnv(ethType, ETH_TX_TYPE)
	viper.BindEnv(ethChainID, ETH_CHAIN_ID)

	viper.BindEnv(ethOptimismL1Sender, ETH_OPTIMISM_L1_SENDER)
	viper.BindEnv(ethOptimismQueueOrigin, ETH_OPTIMISM_QUEUE_ORIGIN)
	viper.BindEnv(ethOptimismRollupTxID, ETH_OPTIMISM_ROLLUP_TX_ID)
	viper.BindEnv(ethOptimismSigHashType, ETH_OPTIMISM_SIG_HASH_TYPE)

	viper.BindEnv(ethDeploymentNumber, ETH_DEPLOYMENT_NUMBER)
	viper.BindEnv(ethDeploymentData, ETH_DEPLOYMENT_HEX_DATA)
	viper.BindEnv(ethDeploymentGasLimit, ETH_DEPLOYMENT_GAS_LIMIT)
	viper.BindEnv(ethDeploymentGasPrice, ETH_DEPLOYMENT_GAS_PRICE)

	viper.BindEnv(ethCallABIPath, ETH_CALL_ABI_PATH)
	viper.BindEnv(ethCallFrequency, ETH_CALL_FREQ)
	viper.BindEnv(ethCallGasLimit, ETH_CALL_GAS_LIMIT)
	viper.BindEnv(ethCallGasPrice, ETH_CALL_GAS_PRICE)
	viper.BindEnv(ethCallMethodName, ETH_CALL_METHOD_NAME)
	viper.BindEnv(ethCallPositionEnd, ETH_CALL_POSITION_END)
	viper.BindEnv(ethCallPositionStart, ETH_CALL_POSITION_START)
	viper.BindEnv(ethCallStorageValue, ETH_CALL_STORAGE_VALUE)
	viper.BindEnv(ethCallTotalNumber, ETH_CALL_TOTAL_NUMBER)

	viper.BindEnv(ethSendFrequency, ETH_SEND_FREQ)
	viper.BindEnv(ethSendTotalNumber, ETH_SEND_TOTAL_NUMBER)
	viper.BindEnv(ethSendAmount, ETH_SEND_AMOUNT)
	viper.BindEnv(ethSendGasLimit, ETH_SEND_GAS_LIMIT)
	viper.BindEnv(ethSendGasPrice, ETH_SEND_GAS_PRICE)
}
