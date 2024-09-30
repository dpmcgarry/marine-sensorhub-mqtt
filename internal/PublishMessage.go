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
package internal

import (
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

// It isn't super efficient to be creating / tearing down connection each time
// But that said the complexity of reusing connections isn't worth it at this point
func PublishMessage(globalconf GlobalConfig, serverConf MQTTDestination) {
	log.Debug().Msgf("Will publish to %v", serverConf.Host)
	mqttOpts := MQTT.NewClientOptions()
	mqttOpts.AddBroker(serverConf.Host)
	client := MQTT.NewClient(mqttOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Warn().Msgf("Error Connecting to host: %v", token.Error())
		return
	}
	for _, topic := range serverConf.Topics {
		log.Info().Msgf("Publish message to host %v on topic %v", serverConf.Host, topic)
		token := client.Publish(topic, byte(0), false, "")
		token.WaitTimeout(time.Duration(globalconf.DisconnectTimeout) * time.Millisecond)
	}
	client.Disconnect(uint(globalconf.DisconnectTimeout))
}
