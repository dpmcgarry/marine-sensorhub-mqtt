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

type Wind struct {
	Source        string
	SpeedApp      float64
	AngleApp      float64
	SOG           float64
	DirectionTrue float64
	Timestamp     time.Time
}

func OnWindMessage(client MQTT.Client, message MQTT.Message) {
	log.Debug().Msgf("Got a message from: %v", message.Topic())
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Debug().Msgf("Got Measurement: %v", measurement)
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	wind := Wind{}
	wind.Source = rawData["$source"].(string)
	wind.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "speedOverGround":
		wind.SOG = rawData["value"].(float64)
	case "directionTrue":
		wind.DirectionTrue = rawData["value"].(float64)
	case "speedApparent":
		wind.SpeedApp = rawData["value"].(float64)
	case "angleApparent":
		wind.AngleApp = rawData["value"].(float64)
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[wind.Source]
	if ok {
		wind.Source = name
	}
	if wind.Timestamp.IsZero() {
		wind.Timestamp = time.Now()
	}
	wind.LogJSON()
}

func (meas Wind) LogJSON() {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	log.Debug().Msgf("BLETemp: %v", string(jsonData))
}
