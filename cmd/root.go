/*
Copyright © 2025 Don P. McGarry

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
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getVersion bool
var version string = "%VER%"
var buildTime string = "%BUILDTIME%"

var rootCmd = &cobra.Command{
	Use:   "marine-sensorhub-mqtt",
	Short: "MQTT Middleware for Marine SensorHub",
	Long: `Performs some middleware operations for Marine Sensorhub.

This CLI can publish the keepalive MQTT messages needed by Victron
and others to enable MQTT publishing in their integrated servers.

This CLI also handles receiving these messages and processing them
into a normalized format for storage and display.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		if getVersion {
			fmt.Println("marine-sensorhub-mqtt")
			fmt.Printf("Copyright © %v Don P. McGarry\n", time.Now().Year())
			fmt.Printf("Version: %v\n", version)
			fmt.Printf("Build Date: %v\n", buildTime)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().BoolVar(&getVersion, "version", false, "Print Version Information")
}

func initConfig() {
	log.Info().Msgf("Using config file: %v", viper.ConfigFileUsed())
}
