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
	"github.com/Sirupsen/logrus"
	"github.com/kildevaeld/keyval"
	"github.com/kildevaeld/keyval/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var httpAddressFlag string

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if err := httpImpl(cmd, args); err != nil {
			printError(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(httpCmd)
	httpCmd.Flags().StringVarP(&httpAddressFlag, "host", "H", ":3000", "")

}

func httpImpl(cmd *cobra.Command, args []string) error {

	var (
		kv     keyval.KeyValStore
		err    error
		server *http.HttpServer
	)

	if kv, err = getKeyValueStore(); err != nil {
		return err
	}

	options := http.ServerOptions{
		WorkQueue:  viper.GetInt("http.work_queue"),
		ScriptPath: interpolate(viper.GetString("http.script_path")),
	}

	if server, err = http.NewServer(kv, options); err != nil {
		return err
	}
	logrus.Infof("kv:http started on %s", httpAddressFlag)
	return server.Listen(httpAddressFlag)
}
