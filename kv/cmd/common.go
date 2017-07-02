package cmd

import (
	"fmt"
	"os"

	"github.com/kildevaeld/keyval"
	"github.com/spf13/viper"
)

func printError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

func getKeyValueStore() (keyval.KeyValStore, error) {

	t := viper.GetString("store.type")
	i := viper.GetStringMap("store.options")

	return keyval.Store(t, i)

}
