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
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

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
		log.Info().Msg("Loading Subscription Config")
		subConf, err := internal.LoadSubscribeServerConfig()
		if err != nil {
			log.Fatal().Msgf("Error reading subscription config. Not much I can do except give up. %v", err.Error())
			os.Exit(2)
		}
		jsonData, err := json.Marshal(subConf)
		if err != nil {
			log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
		}
		log.Debug().Msgf("%v", string(jsonData))
		internal.HandleSubscriptions(subConf)
		log.Info().Msg("Running in daemon mode")
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		done := make(chan bool, 1)

		go func() {

			sig := <-sigs
			log.Warn().Msgf("Got Signal: %v", sig)
			done <- true
		}()

		log.Info().Msg("Awaiting Signal")
		<-done
		log.Info().Msg("Exiting")
	},
}

func init() {
	rootCmd.AddCommand(subCmd)
}
