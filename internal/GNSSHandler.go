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

type GNSS struct {
	Source        string    `json:"Source,omitempty"`
	AntennaAlt    float64   `json:"AntennaAlt,omitempty"`
	Satellites    int64     `json:"Satellites,omitempty"`
	HozDilution   float64   `json:"HozDilution,omitempty"`
	PosDilution   float64   `json:"PosDilution,omitempty"`
	GeoidalSep    float64   `json:"GeoidalSep,omitempty"`
	Type          string    `json:"Type,omitempty"`
	MethodQuality string    `json:"MethodQuality,omitempty"`
	Timestamp     time.Time `json:"Timestamp,omitempty"`
}

func OnGNSSMessage(client MQTT.Client, message MQTT.Message) {
	go handleGNSSMessage(message)
}

func handleGNSSMessage(message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.GNSSLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.GNSSLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	gnss := GNSS{}
	gnss.Source = rawData["$source"].(string)
	gnss.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "antennaAltitude":
		gnss.AntennaAlt = rawData["value"].(float64)
	case "satellites":
		gnss.Satellites = int64(rawData["value"].(float64))
	case "horizontalDilution":
		gnss.HozDilution = rawData["value"].(float64)
	case "positionDilution":
		gnss.PosDilution = rawData["value"].(float64)
	case "geoidalSeparation":
		gnss.GeoidalSep = rawData["value"].(float64)
	case "type":
		gnss.Type = rawData["value"].(string)
	case "methodQuality":
		gnss.MethodQuality = rawData["value"].(string)
	case "integrity":
		break
	case "satellitesInView":
		break
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(gnss.Source)]
	if ok {
		gnss.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", gnss.Source)
	}
	if gnss.Timestamp.IsZero() {
		gnss.Timestamp = time.Now()
	}
	gnss.LogJSON()
}

func (meas GNSS) LogJSON() {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	log.Trace().Msgf("GNSS: %v", string(jsonData))
	if SharedSubscriptionConfig.GNSSLogEn {
		log.Info().Msgf("GNSS: %v", string(jsonData))
	}
}
