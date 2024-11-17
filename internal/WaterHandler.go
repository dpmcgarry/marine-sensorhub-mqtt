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
	"encoding/json"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type Water struct {
	Source                 string    `json:"Source,omitempty"`
	TempF                  float64   `json:"TempF,omitempty"`
	DepthUnderTransducerFt float64   `json:"DepthUnderTransducerFt,omitempty"`
	Timestamp              time.Time `json:"Timestamp,omitempty"`
}

func OnWaterMessage(client MQTT.Client, message MQTT.Message) {
	go handleWaterMessage(message)
}

func handleWaterMessage(message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.WaterLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.WaterLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	water := Water{}
	water.Source = rawData["$source"].(string)
	water.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "temperature":
		water.TempF = KelvinToFarenheit(rawData["value"].(float64))
	case "belowTransducer":
		water.DepthUnderTransducerFt = MetersToFeet(rawData["value"].(float64))
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(water.Source)]
	if ok {
		water.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", water.Source)
	}
	if water.Timestamp.IsZero() {
		water.Timestamp = time.Now()
	}
	water.LogJSON()
}

func (meas Water) LogJSON() {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	log.Trace().Msgf("Water: %v", string(jsonData))
	if SharedSubscriptionConfig.WaterLogEn {
		log.Info().Msgf("Water: %v", string(jsonData))
	}
}
