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
	"crypto/tls"
	"crypto/x509"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

// It isn't super efficient to be creating / tearing down connection each time
// But that said the complexity of reusing connections isn't worth it at this point
func PublishMessage(publishConf PublishConfig, serverConf MQTTDestination) {
	log.Debug().Msgf("Will publish to %v", serverConf.Host)
	mqttOpts := MQTT.NewClientOptions()
	mqttOpts.AddBroker(serverConf.Host)
	if serverConf.Username != "" {
		mqttOpts.SetUsername(serverConf.Username)
		log.Debug().Msgf("Using Username: %v", serverConf.Username)
	}
	if serverConf.Password != "" {
		mqttOpts.SetPassword(serverConf.Password)
		log.Trace().Msgf("Using Password: %v", serverConf.Password)
	}
	if len(serverConf.CACert) > 0 {
		log.Debug().Msg("Constructing x509 Cert Pool")
		rootCAs, err := x509.SystemCertPool()
		if err != nil || rootCAs == nil {
			log.Warn().Msg("Unable to get system cert pool")
			rootCAs = x509.NewCertPool()
		}
		if ok := rootCAs.AppendCertsFromPEM(serverConf.CACert); !ok {
			log.Warn().Msg("No certs appended, using system certs only")
		}

		tlsConfig := &tls.Config{
			RootCAs: rootCAs,
		}
		mqttOpts.SetTLSConfig(tlsConfig)
		log.Debug().Msg("Configured TLS")
	}
	client := MQTT.NewClient(mqttOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Warn().Msgf("Error Connecting to host: %v", token.Error())
		return
	}
	for _, topic := range serverConf.Topics {
		log.Info().Msgf("Publish message to host %v on topic %v", serverConf.Host, topic)
		token := client.Publish(topic, byte(0), false, "")
		token.WaitTimeout(time.Duration(publishConf.DisconnectTimeout) * time.Millisecond)
	}
	client.Disconnect(uint(publishConf.DisconnectTimeout))
}

func PublishClientMessage(client MQTT.Client, topic string, messagedata string) {
	log.Trace().Msgf("Will publish to topic: %v", topic)
	log.Trace().Msgf("Will publish message: %v", messagedata)
	token := client.Publish(topic, byte(0), false, messagedata)
	token.WaitTimeout(time.Duration(SharedSubscriptionConfig.PublishTimeout) * time.Millisecond)
}
