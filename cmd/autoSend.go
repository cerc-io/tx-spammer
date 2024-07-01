// Copyright © 2020 Vulcanize, Inc
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cerc-io/tx-spammer/pkg/auto"
)

// autoSendCmd represents the autoSend command
var autoSendCmd = &cobra.Command{
	Use:   "autoSend",
	Short: "Send large volumes of different tx types to different nodes for testing purposes",
	Long: `Loads tx configuration from .toml config file
Generates txs from configuration and provided private keys and sends them to designated node according to set frequency and number
Support standard, optimism L2, optimism L1 to L2, and EIP1559 transactions`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		autoSend()
	},
}

func init() {
	autoSendCmd.PersistentFlags().Bool("stop-on-error", true,
		"stop service when SendTransaction returns an error")
	viper.BindPFlag(auto.SpammerStopOnError, autoSendCmd.PersistentFlags().Lookup("stop-on-error"))
}

func autoSend() {
	config, err := auto.NewConfig()
	if err != nil {
		logWithCommand.Fatal(err)
	}
	logrus.WithFields(logrus.Fields{"config": config}).Debug("Loaded config")
	txSpammer := auto.NewTxSpammer(config)
	quitChan := make(chan bool)
	doneChan, err := txSpammer.Loop(quitChan)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	go func() {
		shutdown := make(chan os.Signal)
		signal.Notify(shutdown, os.Interrupt)
		<-shutdown
		close(quitChan)
	}()
	<-doneChan
}

func init() {
	rootCmd.AddCommand(autoSendCmd)
}
