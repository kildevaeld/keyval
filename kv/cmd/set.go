// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"os"

	"github.com/kildevaeld/keyval"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:     "set",
	Short:   "",
	Aliases: []string{"s"},
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {

		if err := setImpl(cmd, args); err != nil {
			printError(err)
		}

	},
}

func isPiped() bool {
	fi, _ := os.Stdin.Stat()
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func validateArgs(args []string) error {
	if len(args) < 2 && !isPiped() {
		return errors.New("usage: kv set <key> <value>")
	} else if isPiped() && len(args) == 0 {
		return errors.New("usage: kv set <key>")
	}
	return nil
}

func init() {
	RootCmd.AddCommand(setCmd)
}

func setImpl(cmd *cobra.Command, args []string) error {
	var (
		kv  keyval.KeyValStore
		err error
	)

	if err = validateArgs(args); err != nil {
		return err
	}

	if kv, err = getKeyValueStore(); err != nil {
		return err
	}

	name := []byte(args[0])

	if isPiped() {
		err = kv.Set(name, os.Stdin)
	} else {
		err = kv.SetBytes(name, []byte(args[1]))
	}

	return err
}
