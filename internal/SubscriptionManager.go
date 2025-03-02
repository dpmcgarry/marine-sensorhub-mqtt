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
	"net"
	"net/http"
	"net/url"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

var SharedSubscriptionConfig *SubscriptionConfig
var ISOTimeLayout string = "2006-01-02T15:04:05.000Z"
var SharedInfluxHttpClient *http.Client

func HandleSubscriptions(subscribeconf SubscriptionConfig) {
	SharedSubscriptionConfig = &subscribeconf

	// Create HTTP client
	SharedInfluxHttpClient := &http.Client{
		Timeout: time.Second * time.Duration(60),
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	_ = SharedInfluxHttpClient
	log.Info().Msgf("Will subscribe on server %v", SharedSubscriptionConfig.Host)
	mqttOpts := MQTT.NewClientOptions()
	mqttOpts.AddBroker(SharedSubscriptionConfig.Host)
	if SharedSubscriptionConfig.Username != "" {
		mqttOpts.SetUsername(SharedSubscriptionConfig.Username)
		log.Debug().Msgf("Using Username: %v", SharedSubscriptionConfig.Username)
	}
	if SharedSubscriptionConfig.Password != "" {
		mqttOpts.SetPassword(SharedSubscriptionConfig.Password)
		log.Trace().Msgf("Using Password: %v", SharedSubscriptionConfig.Password)
	}
	if len(SharedSubscriptionConfig.CACert) > 0 {
		log.Debug().Msg("Constructing x509 Cert Pool")
		rootCAs, err := x509.SystemCertPool()
		if err != nil || rootCAs == nil {
			log.Warn().Msg("Unable to get system cert pool")
			rootCAs = x509.NewCertPool()
		}
		if ok := rootCAs.AppendCertsFromPEM(SharedSubscriptionConfig.CACert); !ok {
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

// This function gets called when a connection is successful
// In the event of a reconnect scenario we can resubscribe to topics here
func onConnect(client MQTT.Client) {
	log.Info().Msg("Connected!")
	subscribeToTopics(client)
}

func onReconnect(client MQTT.Client, opts *MQTT.ClientOptions) {
	log.Warn().Msg("Attempting to reconnect")
}

func subscribeToTopics(mqttClient MQTT.Client) {
	if SharedSubscriptionConfig.InfluxEnabled {
		log.Info().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
			SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
	}
	if SharedSubscriptionConfig.BLESubEn {
		for _, topic := range SharedSubscriptionConfig.BLETopics {
			addSubscription(topic, OnBLETemperatureMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.PHYSubEn {
		for _, topic := range SharedSubscriptionConfig.PHYTopics {
			addSubscription(topic, OnPHYTemperatureMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.ESPSubEn {
		for _, topic := range SharedSubscriptionConfig.ESPTopics {
			addSubscription(topic, OnESPStatusMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.NavSubEn {
		for _, topic := range SharedSubscriptionConfig.NavTopics {
			addSubscription(topic, OnNavigationMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.GNSSSubEn {
		for _, topic := range SharedSubscriptionConfig.GNSSTopics {
			addSubscription(topic, OnGNSSMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.SteerSubEn {
		for _, topic := range SharedSubscriptionConfig.SteeringTopics {
			addSubscription(topic, OnSteeringMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.WindSubEn {
		for _, topic := range SharedSubscriptionConfig.WindTopics {
			addSubscription(topic, OnWindMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.WaterSubEn {
		for _, topic := range SharedSubscriptionConfig.WaterTopics {
			addSubscription(topic, OnWaterMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.OutsideSubEn {
		for _, topic := range SharedSubscriptionConfig.OutsideTopics {
			addSubscription(topic, OnOutsideMessage, mqttClient)
		}
	}
	if SharedSubscriptionConfig.PropSubEn {
		for _, topic := range SharedSubscriptionConfig.PropulsionTopics {
			addSubscription(topic, OnPropulsionMessage, mqttClient)
		}
	}
}
