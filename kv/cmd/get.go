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
	"io"
	"os"

	"github.com/kildevaeld/keyval"
	"github.com/spf13/cobra"
)

var outputFlag string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "A brief description of your command",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		if err := getImpl(cmd, args); err != nil {
			printError(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "")

}

func getImpl(cm *cobra.Command, args []string) error {

	if len(args) < 1 {
		return errors.New("usage: kv get <path>")
	}

	var (
		kv   keyval.KeyValStore
		err  error
		file io.ReadCloser
	)

	if kv, err = getKeyValueStore(); err != nil {
		return err
	}

	out := os.Stdout
	if outputFlag != "" {
		if out, err = os.Create(args[0]); err != nil {
			return err
		}
	}

	if file, err = kv.Get([]byte(args[0])); err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(out, file)

	return err
}
