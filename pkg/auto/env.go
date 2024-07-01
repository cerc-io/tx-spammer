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
	SPAMMER_STOP_ON_ERROR = "SPAMMER_STOP_ON_ERROR"

	// env variables
	ETH_KEY_DIR_PATH       = "ETH_KEY_DIR_PATH"
	ETH_HTTP_PATH          = "ETH_HTTP_PATH"
	ETH_CONTRACT_ADDR_PATH = "ETH_CONTRACT_ADDR_PATH"

	ETH_DEPLOYMENT_NUMBER      = "ETH_DEPLOYMENT_NUMBER"
	ETH_DEPLOYMENT_BIN_PATH    = "ETH_DEPLOYMENT_BIN_PATH"
	ETH_DEPLOYMENT_GAS_LIMIT   = "ETH_DEPLOYMENT_GAS_LIMIT"
	ETH_DEPLOYMENT_GAS_FEE_CAP = "ETH_DEPLOYMENT_GAS_FEE_CAP"
	ETH_DEPLOYMENT_GAS_TIP_CAP = "ETH_DEPLOYMENT_GAS_TIP_CAP"

	ETH_CALL_FREQ         = "ETH_CALL_FREQ"
	ETH_CALL_TOTAL_NUMBER = "ETH_CALL_TOTAL_NUMBER"
	ETH_CALL_ABI_PATH     = "ETH_CALL_ABI_PATH"
	ETH_CALL_METHOD_NAME  = "ETH_CALL_METHOD_NAME"
	ETH_CALL_GAS_LIMIT    = "ETH_CALL_GAS_LIMIT"
	ETH_CALL_GAS_FEE_CAP  = "ETH_CALL_GAS_FEE_CAP"
	ETH_CALL_GAS_TIP_CAP  = "ETH_CALL_GAS_TIP_CAP"

	ETH_SEND_FREQ         = "ETH_SEND_FREQ"
	ETH_SEND_TOTAL_NUMBER = "ETH_SEND_TOTAL_NUMBER"
	ETH_SEND_AMOUNT       = "ETH_SEND_AMOUNT"
	ETH_SEND_GAS_LIMIT    = "ETH_SEND_GAS_LIMIT"
	ETH_SEND_GAS_FEE_CAP  = "ETH_SEND_GAS_FEE_CAP"
	ETH_SEND_GAS_TIP_CAP  = "ETH_SEND_GAS_TIP_CAP"

	ETH_BLOBTX_FREQ         = "ETH_BLOBTX_FREQ"
	ETH_BLOBTX_TOTAL_NUMBER = "ETH_BLOBTX_TOTAL_NUMBER"
	ETH_BLOBTX_AMOUNT       = "ETH_BLOBTX_AMOUNT"
	ETH_BLOBTX_GAS_LIMIT    = "ETH_BLOBTX_GAS_LIMIT"
	ETH_BLOBTX_GAS_FEE_CAP  = "ETH_BLOBTX_GAS_FEE_CAP"
	ETH_BLOBTX_GAS_TIP_CAP  = "ETH_BLOBTX_GAS_TIP_CAP"
	ETH_BLOBTX_BLOB_FEE_CAP = "ETH_BLOBTX_BLOB_FEE_CAP"
	ETH_BLOBTX_BLOB_COUNT   = "ETH_BLOBTX_BLOB_COUNT"

	// toml bindings
	SpammerStopOnError = "service.stopOnError"

	ethKeyDirPath       = "eth.keyDirPath"
	ethHttpPath         = "eth.httpPath"
	ethContractAddrPath = "eth.contractAddrPath"

	ethDeploymentBinPath   = "deployment.binPath"
	ethDeploymentNumber    = "deployment.number"
	ethDeploymentGasLimit  = "deployment.gasLimit"
	ethDeploymentGasFeeCap = "deployment.gasFeeCap"
	ethDeploymentGasTipCap = "deployment.gasTipCap"

	ethCallFrequency   = "contractSpammer.frequency"
	ethCallTotalNumber = "contractSpammer.totalNumber"
	ethCallABIPath     = "contractSpammer.abiPath"
	ethCallMethodName  = "contractSpammer.methodName"
	ethCallGasLimit    = "contractSpammer.gasLimit"
	ethCallGasFeeCap   = "contractSpammer.gasFeeCap"
	ethCallGasTipCap   = "contractSpammer.gasTipCap"

	ethSendFrequency   = "sendSpammer.frequency"
	ethSendTotalNumber = "sendSpammer.totalNumber"
	ethSendAmount      = "sendSpammer.amount"
	ethSendGasLimit    = "sendSpammer.gasLimit"
	ethSendGasFeeCap   = "sendSpammer.gasFeeCap"
	ethSendGasTipCap   = "sendSpammer.gasTipCap"

	ethBlobTxFrequency   = "blobSpammer.frequency"
	ethBlobTxTotalNumber = "blobSpammer.totalNumber"
	ethBlobTxAmount      = "blobSpammer.amount"
	ethBlobTxGasLimit    = "blobSpammer.gasLimit"
	ethBlobTxGasFeeCap   = "blobSpammer.gasFeeCap"
	ethBlobTxGasTipCap   = "blobSpammer.gasTipCap"
	ethBlobTxBlobFeeCap  = "blobSpammer.blobFeeCap"
	ethBlobTxBlobCount   = "blobSpammer.blobCount"
)

func bindEnv() {
	viper.BindEnv(SpammerStopOnError, SPAMMER_STOP_ON_ERROR)

	viper.BindEnv(ethKeyDirPath, ETH_KEY_DIR_PATH)
	viper.BindEnv(ethHttpPath, ETH_HTTP_PATH)
	viper.BindEnv(ethContractAddrPath, ETH_CONTRACT_ADDR_PATH)

	viper.BindEnv(ethDeploymentNumber, ETH_DEPLOYMENT_NUMBER)
	viper.BindEnv(ethDeploymentBinPath, ETH_DEPLOYMENT_BIN_PATH)
	viper.BindEnv(ethDeploymentGasLimit, ETH_DEPLOYMENT_GAS_LIMIT)
	viper.BindEnv(ethDeploymentGasFeeCap, ETH_DEPLOYMENT_GAS_FEE_CAP)
	viper.BindEnv(ethDeploymentGasTipCap, ETH_DEPLOYMENT_GAS_TIP_CAP)

	viper.BindEnv(ethCallABIPath, ETH_CALL_ABI_PATH)
	viper.BindEnv(ethCallFrequency, ETH_CALL_FREQ)
	viper.BindEnv(ethCallGasLimit, ETH_CALL_GAS_LIMIT)
	viper.BindEnv(ethCallGasFeeCap, ETH_CALL_GAS_FEE_CAP)
	viper.BindEnv(ethCallGasTipCap, ETH_CALL_GAS_TIP_CAP)
	viper.BindEnv(ethCallMethodName, ETH_CALL_METHOD_NAME)
	viper.BindEnv(ethCallTotalNumber, ETH_CALL_TOTAL_NUMBER)

	viper.BindEnv(ethSendFrequency, ETH_SEND_FREQ)
	viper.BindEnv(ethSendTotalNumber, ETH_SEND_TOTAL_NUMBER)
	viper.BindEnv(ethSendAmount, ETH_SEND_AMOUNT)
	viper.BindEnv(ethSendGasLimit, ETH_SEND_GAS_LIMIT)
	viper.BindEnv(ethSendGasFeeCap, ETH_SEND_GAS_FEE_CAP)
	viper.BindEnv(ethSendGasTipCap, ETH_SEND_GAS_TIP_CAP)
	viper.BindEnv(ethSendGasLimit, ETH_CALL_GAS_LIMIT)

	viper.BindEnv(ethBlobTxFrequency, ETH_BLOBTX_FREQ)
	viper.BindEnv(ethBlobTxTotalNumber, ETH_BLOBTX_TOTAL_NUMBER)
	viper.BindEnv(ethBlobTxAmount, ETH_BLOBTX_AMOUNT)
	viper.BindEnv(ethBlobTxGasLimit, ETH_BLOBTX_GAS_LIMIT)
	viper.BindEnv(ethBlobTxGasFeeCap, ETH_BLOBTX_GAS_FEE_CAP)
	viper.BindEnv(ethBlobTxGasTipCap, ETH_BLOBTX_GAS_TIP_CAP)
	viper.BindEnv(ethBlobTxBlobFeeCap, ETH_BLOBTX_BLOB_FEE_CAP)
	viper.BindEnv(ethBlobTxBlobCount, ETH_BLOBTX_BLOB_COUNT)
}
