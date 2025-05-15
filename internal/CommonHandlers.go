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
package internal

import (
	"context"
	"encoding/json"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

// HandleJSONMessage handles messages that are already in JSON format
// This is used by BLETemperature, ESPStatus, and PHYTemperature handlers
func HandleJSONMessage(client MQTT.Client, message MQTT.Message, data SensorData) {
	// Log message receipt
	logEnabled := data.GetLogEnabled()
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if logEnabled {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}

	// Unmarshal JSON directly into the data struct
	err := json.Unmarshal(message.Payload(), data)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
		return
	}
}

func SendJSONMessage(client MQTT.Client, message MQTT.Message, data SensorData) {
	// Log message receipt
	logEnabled := data.GetLogEnabled()

	// Skip empty data
	if data.IsEmpty() {
		return
	}

	// Log the data
	data.LogJSON()

	// Publish to MQTT if enabled
	if SharedSubscriptionConfig.Repost {
		measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
		PublishClientMessage(client,
			SharedSubscriptionConfig.RepostRootTopic+"vessel/"+data.GetTopicPrefix()+"/"+data.GetSource()+"/"+measurement,
			data.ToJSON(), true)
	}

	// Write to InfluxDB if enabled
	if SharedSubscriptionConfig.InfluxEnabled {
		p := data.ToInfluxPoint()
		err := SharedInfluxWriteAPI.WritePoint(context.Background(), p)
		if err != nil {
			log.Warn().Msgf("Error writing to influx: %v", err.Error())
		}
		log.Trace().Msg("Wrote Point")
		if logEnabled {
			log.Debug().Msg("Wrote Point")
		}
	}
}

// MapMACToLocation maps a MAC address to a location name
func MapMACToLocation(data SensorData, mac string) string {
	loc, ok := SharedSubscriptionConfig.MACtoLocation[strings.ToLower(mac)]
	if ok {
		return loc
	}
	log.Warn().Msgf("Location not found for MAC %v", mac)
	return ""
}
