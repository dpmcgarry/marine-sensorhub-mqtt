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
	"os"
	"time"

	"github.com/dpmcgarry/marine-sensorhub-mqtt/internal"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var isDaemon bool
var numIter uint

var pubCmd = &cobra.Command{
	Use:   "pub",
	Short: "Publishes Keepalive Messages",
	Long:  `Publishes Keepalive Messages.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Starting Publish")
		log.Info().Msg("Loading Publish Config")
		publishConf, err := internal.LoadPublishConfig()
		if err != nil {
			log.Fatal().Msgf("Error reading Global config. Not much I can do except give up. %v", err.Error())
			os.Exit(2)
		}
		log.Info().Msg("Loading Server Config")
		serverConfs, err := internal.LoadPublishServerConfig()
		if err != nil {
			log.Fatal().Msgf("Error reading Server config. Not much I can do except give up. %v", err.Error())
			os.Exit(2)
		}
		if isDaemon {
			log.Info().Msg("Running in daemon mode")
			for {
				for _, serverConf := range serverConfs {
					internal.PublishMessage(publishConf, serverConf)
				}
				log.Debug().Msgf("Sleeping for %v seconds", publishConf.Interval)
				time.Sleep(time.Duration(publishConf.Interval) * time.Second)
			}
		} else {
			log.Info().Msgf("Will only run for %v iterations", numIter)
			for i := 0; i < int(numIter); i++ {
				for _, serverConf := range serverConfs {
					internal.PublishMessage(publishConf, serverConf)
				}
				log.Debug().Msgf("Sleeping for %v seconds", publishConf.Interval)
				time.Sleep(time.Duration(publishConf.Interval) * time.Second)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(pubCmd)
	pubCmd.Flags().BoolVarP(&isDaemon, "daemon", "d", false, "Set to run as a continuous daemon")
	pubCmd.Flags().UintVarP(&numIter, "iter", "i", 0, "Number of Iterations to run before exiting")
	pubCmd.MarkFlagsOneRequired("daemon", "iter")
	pubCmd.MarkFlagsMutuallyExclusive("daemon", "iter")
}
