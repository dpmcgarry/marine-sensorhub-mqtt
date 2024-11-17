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

type Outside struct {
	Source    string    `json:"Source,omitempty"`
	TempF     float64   `json:"TempF,omitempty"`
	Pressure  float64   `json:"Pressure,omitempty"`
	Timestamp time.Time `json:"Timestamp,omitempty"`
}

func OnOutsideMessage(client MQTT.Client, message MQTT.Message) {
	log.Debug().Msgf("Got a message from: %v", message.Topic())
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Debug().Msgf("Got Measurement: %v", measurement)
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	out := Outside{}
	out.Source = rawData["$source"].(string)
	out.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "temperature":
		out.TempF = rawData["value"].(float64)
	case "pressure":
		out.Pressure = rawData["value"].(float64)
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[out.Source]
	if ok {
		out.Source = name
	}
	if out.Timestamp.IsZero() {
		out.Timestamp = time.Now()
	}
	out.LogJSON()
}

func (meas Outside) LogJSON() {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	log.Debug().Msgf("Outside: %v", string(jsonData))
}
