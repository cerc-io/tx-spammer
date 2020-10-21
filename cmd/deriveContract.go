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

// deriveContractCmd represents the deriveContract command
var deriveContractCmd = &cobra.Command{
	Use:   "deriveContract",
	Short: "Derive contract address",
	Long:  `Derive the contract address created from an pubkey/address and nonce`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		deriveContract()
	},
}

func deriveContract() {
	nonce := viper.GetUint64("keyGen.nonce")
	addrStr := viper.GetString("keyGen.address")
	var addr common.Address
	if addrStr == "" {
		keyPath := viper.GetString("keyGen.path")
		key, err := crypto.LoadECDSA(keyPath)
		if err != nil {
			logWithCommand.Fatal(err)
		}
		addr = crypto.PubkeyToAddress(key.PublicKey)
	} else {
		addr = common.HexToAddress(addrStr)
	}
	contractAddr := crypto.CreateAddress(addr, nonce)
	fmt.Println(contractAddr.Hex())
}

func init() {
	rootCmd.AddCommand(deriveContractCmd)

	deriveContractCmd.PersistentFlags().Uint64("nonce", 0, "nonce to derive contract address from")
	deriveContractCmd.PersistentFlags().String("address", "", "address to derive contract address from")

	viper.BindPFlag("keyGen.nonce", deriveContractCmd.PersistentFlags().Lookup("nonce"))
	viper.BindPFlag("keyGen.address", deriveContractCmd.PersistentFlags().Lookup("address"))
}
