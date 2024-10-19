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
package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Need a separate function to configure viper before logging is configured
func configureViper() {
	viper.SetConfigName("marine-sensorhub-mqtt.conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath(".")
	viper.SetDefault("logdir", "./")
	viper.SetDefault("interval", 15)
	viper.SetDefault("publishtimeout", 250)
	viper.SetDefault("disconnecttimeout", 250)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Fatal: Config File not found in /etc or current dir: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("Fatal: Error reading config file but file was found: %v\n", err)
			os.Exit(1)
		}
	}

	viper.AutomaticEnv()
}
