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

type PHYTemperature struct {
	MAC       string    `json:"MAC,omitempty"`
	Location  string    `json:"Location,omitempty"`
	Device    string    `json:"Device,omitempty"`
	Component string    `json:"Component,omitempty"`
	TempF     float64   `json:"TempF,omitempty"`
	Timestamp time.Time `json:"Timestamp,omitempty"`
}

func OnPHYTemperatureMessage(client MQTT.Client, message MQTT.Message) {
	go handlePHYTemperatureMessage(client, message)
}

func handlePHYTemperatureMessage(client MQTT.Client, message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.PHYLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	phyTemp := PHYTemperature{}
	err := json.Unmarshal(message.Payload(), &phyTemp)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	loc, ok := SharedSubscriptionConfig.MACtoLocation[strings.ToLower(phyTemp.MAC)]
	if ok {
		phyTemp.Location = loc
	} else {
		log.Warn().Msgf("Location not found for MAC %v", phyTemp.MAC)
	}
	if phyTemp.Timestamp.IsZero() {
		phyTemp.Timestamp = time.Now()
	}
	phyTemp.LogJSON()
	if SharedSubscriptionConfig.Repost {
		PublishClientMessage(client, SharedSubscriptionConfig.RepostRootTopic+"rtd/temperature/"+phyTemp.Location, phyTemp.ToJSON())
	}
}

func (meas PHYTemperature) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas PHYTemperature) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Physical Temp: %v", json)
	if SharedSubscriptionConfig.PHYLogEn {
		log.Info().Msgf("Physical Temp: %v", json)
	}
}
