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
	"net/url"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

var SharedSubscriptionConfig *SubscriptionConfig
var ISOTimeLayout string = "2006-01-02T15:04:05.000Z"

func HandleSubscriptions(subscribeconf SubscriptionConfig) {
	SharedSubscriptionConfig = &subscribeconf
	log.Info().Msgf("Will subscribe on server %v", subscribeconf.Host)
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
	mqttOpts.SetAutoReconnect(true)
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectionAttemptHandler(onConnectionAttempt)
	mqttOpts.SetConnectionLostHandler(onConnectionLost)
	mqttOpts.SetOnConnectHandler(onConnect)
	mqttOpts.SetReconnectingHandler(onReconnect)
	mqttClient := MQTT.NewClient(mqttOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Warn().Msgf("Error Connecting to host: %v", token.Error())
		return
	}
	if SharedSubscriptionConfig.InfluxEnabled {
		log.Info().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
			SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
	}
	if subscribeconf.BLESubEn {
		for _, topic := range subscribeconf.BLETopics {
			addSubscription(topic, OnBLETemperatureMessage, mqttClient)
		}
	}
	if subscribeconf.PHYSubEn {
		for _, topic := range subscribeconf.PHYTopics {
			addSubscription(topic, OnPHYTemperatureMessage, mqttClient)
		}
	}
	if subscribeconf.ESPSubEn {
		for _, topic := range subscribeconf.ESPTopics {
			addSubscription(topic, OnESPStatusMessage, mqttClient)
		}
	}
	if subscribeconf.NavSubEn {
		for _, topic := range subscribeconf.NavTopics {
			addSubscription(topic, OnNavigationMessage, mqttClient)
		}
	}
	if subscribeconf.GNSSSubEn {
		for _, topic := range subscribeconf.GNSSTopics {
			addSubscription(topic, OnGNSSMessage, mqttClient)
		}
	}
	if subscribeconf.SteerSubEn {
		for _, topic := range subscribeconf.SteeringTopics {
			addSubscription(topic, OnSteeringMessage, mqttClient)
		}
	}
	if subscribeconf.WindSubEn {
		for _, topic := range subscribeconf.WindTopics {
			addSubscription(topic, OnWindMessage, mqttClient)
		}
	}
	if subscribeconf.WaterSubEn {
		for _, topic := range subscribeconf.WaterTopics {
			addSubscription(topic, OnWaterMessage, mqttClient)
		}
	}
	if subscribeconf.OutsideSubEn {
		for _, topic := range subscribeconf.OutsideTopics {
			addSubscription(topic, OnOutsideMessage, mqttClient)
		}
	}
	if subscribeconf.PropSubEn {
		for _, topic := range subscribeconf.PropulsionTopics {
			addSubscription(topic, OnPropulsionMessage, mqttClient)
		}
	}
}

func addSubscription(topic string, target MQTT.MessageHandler, mqttClient MQTT.Client) {
	log.Info().Msgf("Subscribing to topic: %v", topic)
	if token := mqttClient.Subscribe(topic, byte(0), target); token.Wait() && token.Error() != nil {
		log.Warn().Msgf("Error subscribing to topic %v with error %v", topic, token.Error())
	}
}

func onConnectionAttempt(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
	log.Info().Msgf("Attempting connection to: %v", broker.String())
	return tlsCfg
}

func onConnectionLost(client MQTT.Client, err error) {
	log.Warn().Msgf("Connection Lost! %v", err.Error())
}

func onConnect(client MQTT.Client) {
	log.Info().Msg("Connected!")
}

func onReconnect(client MQTT.Client, opts *MQTT.ClientOptions) {
	log.Warn().Msg("Attempting to reconnect")
}
