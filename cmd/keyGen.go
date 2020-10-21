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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// keyGenCmd represents the keyGen command
var keyGenCmd = &cobra.Command{
	Use:   "keyGen",
	Short: "Generates ethereum key pairs",
	Long: `Generates a new ethereum key pair for each file path provided
These keys should only be used for testing purposes`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		keyGen()
	},
}

func keyGen() {
	// and their .toml config bindings
	keyPaths := viper.GetStringSlice("keyGen.paths")
	for _, path := range keyPaths {
		key, err := crypto.GenerateKey()
		if err != nil {
			logWithCommand.Fatal(err)
		}
		if err := crypto.SaveECDSA(path, key); err != nil {
			logWithCommand.Fatal(err)
		}
	}
}

func init() {
	rootCmd.AddCommand(keyGenCmd)

	keyGenCmd.PersistentFlags().StringSlice("write-paths", nil, "file paths to write keys to; generate a key for each path provided")
	viper.BindPFlag("keyGen.paths", keyGenCmd.PersistentFlags().Lookup("write-paths"))
}
