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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dpmcgarry/marine-sensorhub-mqtt/internal"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Subscribes for Messages",
	Long: `Subscribes to MQTT Topics.
When messages are received, processes them.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Starting Subscribe")
		log.Info().Msg("Loading Global Config")
		globalConf, err := internal.LoadGlobalConfig()
		if err != nil {
			log.Fatal().Msgf("Error reading Global config. Not much I can do except give up. %v", err.Error())
			os.Exit(2)
		}
		log.Info().Msg("Loading Subscription Config")
		subConf, err := internal.LoadSubscribeServerConfig()
		if err != nil {
			log.Fatal().Msgf("Error reading subscription config. Not much I can do except give up. %v", err.Error())
			os.Exit(2)
		}
		log.Debug().Msgf("%v", subConf)
		internal.HandleSubscriptions(globalConf, subConf)
		if isDaemon {
			log.Info().Msg("Running in daemon mode")
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			done := make(chan bool, 1)

			go func() {

				sig := <-sigs
				log.Info().Msgf("Got Signal: %v", sig)
				done <- true
			}()

			log.Info().Msg("Awaiting Signal")
			<-done
			log.Info().Msg("Exiting")
		} else {
			log.Info().Msgf("Will only run for %v iterations", numIter)
			for i := 0; i < int(numIter); i++ {
				log.Debug().Msgf("Sleeping for %v seconds", globalConf.Interval)
				time.Sleep(time.Duration(globalConf.Interval) * time.Second)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(subCmd)
	subCmd.Flags().BoolVarP(&isDaemon, "daemon", "d", false, "Set to run as a continuous daemon")
	subCmd.Flags().UintVarP(&numIter, "iter", "i", 0, "Number of Iterations to run before exiting")
	subCmd.MarkFlagsOneRequired("daemon", "iter")
	subCmd.MarkFlagsMutuallyExclusive("daemon", "iter")
}
