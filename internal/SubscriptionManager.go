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

var SharedSubscriptionConfig *SubscriptionConfig
var ISOTimeLayout string = "2006-01-02T15:04:05.000Z"

func HandleSubscriptions(subscribeconf SubscriptionConfig) {
	SharedSubscriptionConfig = &subscribeconf
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
	mqttOpts.AutoReconnect = true
	mqttClient := MQTT.NewClient(mqttOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Warn().Msgf("Error Connecting to host: %v", token.Error())
		return
	}

	if subscribeconf.BLESubEn {
		topic := subscribeconf.ESPMSHRootTopic + "ble/temperature/#"
		addSubscription(topic, OnBLETemperatureMessage, mqttClient)
	}
	if subscribeconf.PHYSubEn {
		topic := subscribeconf.ESPMSHRootTopic + "rtd/temperature/#"
		addSubscription(topic, OnPHYTemperatureMessage, mqttClient)
	}
	if subscribeconf.ESPSubEn {
		topic := subscribeconf.ESPMSHRootTopic + "esp/status/#"
		addSubscription(topic, OnESPStatusMessage, mqttClient)
	}
	if subscribeconf.NavSubEn {
		topic := subscribeconf.SignalKRootTopic + "vessels/self/navigation/+"
		addSubscription(topic, OnNavigationMessage, mqttClient)
	}
	if subscribeconf.GNSSSubEn {
		topic := subscribeconf.SignalKRootTopic + "vessels/self/navigation/gnss/#"
		addSubscription(topic, OnGNSSMessage, mqttClient)
	}
	if subscribeconf.SteerSubEn {
		topic := subscribeconf.SignalKRootTopic + "vessels/self/steering/#"
		addSubscription(topic, OnSteeringMessage, mqttClient)
	}
	if subscribeconf.WindSubEn {
		topic := subscribeconf.SignalKRootTopic + "vessels/self/environment/wind/#"
		addSubscription(topic, OnWindMessage, mqttClient)
	}
	if subscribeconf.WaterSubEn {
		topic := subscribeconf.SignalKRootTopic + "vessels/self/environment/water/#"
		addSubscription(topic, OnWaterMessage, mqttClient)
		topic = subscribeconf.SignalKRootTopic + "vessels/self/environment/depth/#"
		addSubscription(topic, OnWaterMessage, mqttClient)
	}
	if subscribeconf.OutsideSubEn {
		topic := subscribeconf.SignalKRootTopic + "vessels/self/environment/outside/#"
		addSubscription(topic, OnOutsideMessage, mqttClient)
	}
	if subscribeconf.PropSubEn {
		topic := subscribeconf.SignalKRootTopic + "vessels/self/propulsion/#"
		addSubscription(topic, OnPropulsionMessage, mqttClient)
	}

}

func addSubscription(topic string, target MQTT.MessageHandler, mqttClient MQTT.Client) {
	log.Info().Msgf("Subscribing to topic: %v", topic)
	if token := mqttClient.Subscribe(topic, byte(0), target); token.Wait() && token.Error() != nil {
		log.Warn().Msgf("Error subscribing to topic %v with error %v", topic, token.Error())
	}
}
