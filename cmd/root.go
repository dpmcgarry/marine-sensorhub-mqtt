/*
Copyright © 2024 Don P. McGarry

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "mqtt-keepalive",
	Short: "Sends MQTT Messages as a keepalive",
	Long: `Certain MQTT servers such as the Victron Cerbo and SignalK
require an empty MQTT message sent to a specific topic to keep the
server alive. This program is intended as a simple daemon to send those
messages.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mqtt-keepalive.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("mqtt-keepalive.conf")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/")
		viper.AddConfigPath(".")
		viper.SetDefault("logdir", "./")
		viper.SetDefault("interval", 15)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Printf("Config File not found in /etc or current dir: %v\n", err)
				os.Exit(1)
			} else {
				fmt.Printf("Error reading config file but file was found: %v\n", err)
				os.Exit(1)
			}
		}
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
