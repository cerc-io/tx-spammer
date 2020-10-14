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
	"sync"

	"github.com/spf13/cobra"
	"github.com/vulcanize/tx_spammer/pkg"
)

// sendTxsCmd represents the sendTxs command
var sendTxsCmd = &cobra.Command{
	Use:   "sendTxs",
	Short: "send large volumes of different tx types to different nodes",
	Long: `Loads tx configuration from .toml config file
Generates txs from configuration and sends them to designated node according to set frequency and number`,
	Run: func(cmd *cobra.Command, args []string) {
		sendTxs()
	},
}

func sendTxs() {
	params, err := tx_spammer.NewTxParams()
	if err != nil {
		logWithCommand.Fatal(err)
	}
	txSpammer := tx_spammer.NewTxSpammer(params)
	wg := new(sync.WaitGroup)
	quitChan := make(chan bool)
	txSpammer.Loop(wg, quitChan)

	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	close(quitChan)
	wg.Wait()

}
func init() {
	rootCmd.AddCommand(sendTxsCmd)
}
