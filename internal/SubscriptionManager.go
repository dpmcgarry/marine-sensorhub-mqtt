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

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

func HandleSubscriptions(globalconf GlobalConfig, subscribeconf SubscriptionConfig) {
	log.Debug().Msgf("Will subscribe on server %v", subscribeconf.Host)
	mqttOpts := MQTT.NewClientOptions()
	mqttOpts.AddBroker(subscribeconf.Host)
	if subscribeconf.Username != "" {
		mqttOpts.SetUsername(subscribeconf.Username)
		log.Debug().Msgf("Using Username: %v", subscribeconf.Username)
	}
	if subscribeconf.Password != "" {
		mqttOpts.SetPassword(subscribeconf.Password)
		log.Trace().Msgf("Using Password: %v", subscribeconf.Password)
	}
	if len(subscribeconf.CACert) > 0 {
		log.Debug().Msg("Constructing x509 Cert Pool")
		rootCAs, err := x509.SystemCertPool()
		if err != nil || rootCAs == nil {
			log.Warn().Msg("Unable to get system cert pool")
			rootCAs = x509.NewCertPool()
		}
		if ok := rootCAs.AppendCertsFromPEM(subscribeconf.CACert); !ok {
			log.Warn().Msg("No certs appended, using system certs only")
		}

		tlsConfig := &tls.Config{
			RootCAs: rootCAs,
		}
		mqttOpts.SetTLSConfig(tlsConfig)
		log.Debug().Msg("Configured TLS")
	}
	mqttClient := MQTT.NewClient(mqttOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Warn().Msgf("Error Connecting to host: %v", token.Error())
		return
	}

	if token := mqttClient.Subscribe(subscribeconf.ESPMSHRootTopic+"#", byte(0), OnMSHMessage); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}
