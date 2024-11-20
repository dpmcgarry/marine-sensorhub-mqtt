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

type Steering struct {
	Source           string    `json:"Source,omitempty"`
	RudderAngle      float64   `json:"RudderAngle,omitempty"`
	AutopilotState   string    `json:"AutoPilotState,omitempty"`
	TargetHeadingMag float64   `json:"TargetHeading,omitempty"`
	Timestamp        time.Time `json:"Timestamp,omitempty"`
}

func OnSteeringMessage(client MQTT.Client, message MQTT.Message) {
	go handleSteeringMessage(client, message)
}

func handleSteeringMessage(client MQTT.Client, message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.SteerLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.SteerLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	steer := Steering{}
	steer.Source = rawData["$source"].(string)
	steer.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "rudderAngle":
		steer.RudderAngle = RadiansToDegrees(rawData["value"].(float64))
	case "autopilot":
		break
	case "state":
		steer.AutopilotState = rawData["value"].(string)
	case "target":
		break
	case "headingMagnetic":
		steer.TargetHeadingMag = RadiansToDegrees(rawData["value"].(float64))
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(steer.Source)]
	if ok {
		steer.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", steer.Source)
	}
	if steer.Timestamp.IsZero() {
		steer.Timestamp = time.Now()
	}
	if !steer.IsEmpty() {
		steer.LogJSON()
		if SharedSubscriptionConfig.Repost {
			if measurement == "state" {
				measurement = "autopilotState"
			}
			PublishClientMessage(client,
				SharedSubscriptionConfig.RepostRootTopic+"vessel/steering/"+steer.Source+"/"+measurement, steer.ToJSON())
		}
	}
}

func (meas Steering) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas Steering) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Steering: %v", json)
	if SharedSubscriptionConfig.SteerLogEn {
		log.Info().Msgf("Steering: %v", json)
	}
}

// Since we are dropping fields we can end up with messages that are just a Source and a Timestamp
// This allows us to drop those messages
func (meas Steering) IsEmpty() bool {
	if meas.RudderAngle == 0.0 && meas.AutopilotState == "" && meas.TargetHeadingMag == 0.0 {
		return true
	}
	return false
}
