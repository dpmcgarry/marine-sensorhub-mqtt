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

type Navigation struct {
	Source       string    `json:"Source,omitempty"`
	Lat          float64   `json:"Latitude,omitempty"`
	Lon          float64   `json:"Longitude,omitempty"`
	Alt          float64   `json:"Altitude,omitempty"`
	SOG          float64   `json:"SpeedOverGround,omitempty"`
	ROT          float64   `json:"RateOfTurn,omitempty"`
	COGTrue      float64   `json:"CourseOverGroundTrue,omitempty"`
	HeadingMag   float64   `json:"HeadingMagnetic,omitempty"`
	MagVariation float64   `json:"MagneticVariation,omitempty"`
	MagDeviation float64   `json:"MagneticDeviation,omitempty"`
	Yaw          float64   `json:"Yaw,omitempty"`
	Pitch        float64   `json:"Pitch,omitempty"`
	Roll         float64   `json:"Roll,omitempty"`
	HeadingTrue  float64   `json:"HeadingTrue,omitempty"`
	STW          float64   `json:"SpeedThroughWater,omitempty"`
	Timestamp    time.Time `json:"Timestamp,omitempty"`
}

func OnNavigationMessage(client MQTT.Client, message MQTT.Message) {
	log.Debug().Msgf("Got a message from: %v", message.Topic())
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Debug().Msgf("Got Measurement: %v", measurement)
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	nav := Navigation{}
	nav.Source = rawData["$source"].(string)
	nav.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "headingMagnetic":
		nav.HeadingMag = rawData["value"].(float64)
	case "rateOfTurn":
		nav.ROT = rawData["value"].(float64)
	case "speedOverGround":
		nav.SOG = rawData["value"].(float64)
	case "position":
		postmp := rawData["value"].(map[string]any)
		nav.Lat = postmp["latitude"].(float64)
		nav.Lat = postmp["longitude"].(float64)
		alt, ok := postmp["altitude"].(float64)
		if ok {
			nav.Alt = alt
		}
	case "headingTrue":
		nav.HeadingTrue = rawData["value"].(float64)
	case "magneticVariation":
		nav.MagVariation = rawData["value"].(float64)
	case "magneticDeviation":
		nav.MagDeviation = rawData["value"].(float64)
	case "datetime":
		break
	case "courseOverGroundTrue":
		nav.COGTrue = rawData["value"].(float64)
	case "attitude":
		atttmp := rawData["value"].(map[string]any)
		flttmp, ok := atttmp["yaw"].(float64)
		if ok {
			nav.Yaw = flttmp
		}
		flttmp, ok = atttmp["pitch"].(float64)
		if ok {
			nav.Pitch = flttmp
		}
		flttmp, ok = atttmp["roll"].(float64)
		if ok {
			nav.Yaw = flttmp
		}
	case "speedThroughWater":
		nav.STW = rawData["value"].(float64)
	case "speedThroughWaterReferenceType":
		break
	case "log":
		break
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[nav.Source]
	if ok {
		nav.Source = name
	}
	if nav.Timestamp.IsZero() {
		nav.Timestamp = time.Now()
	}
	nav.LogJSON()
}

func (meas Navigation) LogJSON() {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	log.Debug().Msgf("Navigation: %v", string(jsonData))
}
