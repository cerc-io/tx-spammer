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
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deriveAddressCmd represents the deriveAddress command
var deriveAddressCmd = &cobra.Command{
	Use:   "deriveAddress",
	Short: "Derive address from key pair",
	Long:  `Derive the account address from an pubkey/address`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		deriveAddress()
	},
}

func deriveAddress() {
	var addr common.Address
	keyPath := viper.GetString("keyGen.path")
	key, err := crypto.LoadECDSA(keyPath)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	addr = crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println(addr.Hex())
}

func init() {
	rootCmd.AddCommand(deriveAddressCmd)
}
