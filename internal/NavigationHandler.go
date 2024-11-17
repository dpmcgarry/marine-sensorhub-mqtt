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
	go handleNavigationMessage(message)
}

func handleNavigationMessage(message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.NavLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.NavLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	nav := Navigation{}
	nav.Source = rawData["$source"].(string)
	if strings.Contains(nav.Source, "venus.com.victronenergy.gps.") {
		return
	}
	nav.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "headingMagnetic":
		nav.HeadingMag = RadiansToDegrees(rawData["value"].(float64))
	case "rateOfTurn":
		nav.ROT = RadiansToDegrees(rawData["value"].(float64))
	case "speedOverGround":
		nav.SOG = MetersPerSecondToKnots(rawData["value"].(float64))
	case "position":
		postmp := rawData["value"].(map[string]any)
		nav.Lat = postmp["latitude"].(float64)
		nav.Lat = postmp["longitude"].(float64)
		alt, ok := postmp["altitude"].(float64)
		if ok {
			nav.Alt = MetersToFeet(alt)
		}
	case "headingTrue":
		nav.HeadingTrue = RadiansToDegrees(rawData["value"].(float64))
	case "magneticVariation":
		nav.MagVariation = RadiansToDegrees(rawData["value"].(float64))
	case "magneticDeviation":
		nav.MagDeviation = RadiansToDegrees(rawData["value"].(float64))
	case "datetime":
		break
	case "courseOverGroundTrue":
		nav.COGTrue = RadiansToDegrees(rawData["value"].(float64))
	case "attitude":
		atttmp := rawData["value"].(map[string]any)
		flttmp, ok := atttmp["yaw"].(float64)
		if ok {
			nav.Yaw = RadiansToDegrees(flttmp)
		}
		flttmp, ok = atttmp["pitch"].(float64)
		if ok {
			nav.Pitch = RadiansToDegrees(flttmp)
		}
		flttmp, ok = atttmp["roll"].(float64)
		if ok {
			nav.Yaw = RadiansToDegrees(flttmp)
		}
	case "speedThroughWater":
		nav.STW = MetersPerSecondToKnots(rawData["value"].(float64))
	case "speedThroughWaterReferenceType":
		break
	case "log":
		break
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(nav.Source)]
	if ok {
		nav.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", nav.Source)
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
	log.Trace().Msgf("Navigation: %v", string(jsonData))
	if SharedSubscriptionConfig.NavLogEn {
		log.Info().Msgf("Navigation: %v", string(jsonData))
	}
}
