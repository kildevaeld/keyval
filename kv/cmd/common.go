package cmd

import (
	"fmt"
	"os"
	"regexp"

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

var reg = regexp.MustCompile("\\$([a-zA-Z_]+)")

func interpolate(str string) string {
	return reg.ReplaceAllStringFunc(str, func(str string) string {
		if str == "$KV_CONFIG_PATH" {
			return viper.GetString("config_path")
		}
		return os.Getenv(str[1:])
	})
}
