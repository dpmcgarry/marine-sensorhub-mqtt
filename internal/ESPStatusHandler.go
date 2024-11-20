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

type ESPStatus struct {
	MAC                string    `json:"MAC,omitempty"`
	Location           string    `json:"Location,omitempty"`
	IPAddress          string    `json:"IPAddress,omitempty"`
	MSHVersion         string    `json:"MSHVersion,omitempty"`
	FreeSRAM           int64     `json:"FreeSRAM,omitempty"`
	FreeHeap           int64     `json:"FreeHeap,omitempty"`
	FreePSRAM          int64     `json:"FreePSRAM,omitempty"`
	WiFiReconnectCount int64     `json:"WiFiReconnectCount,omitempty"`
	MQTTReconnectCount int64     `json:"MQTTReconnectCount,omitempty"`
	BLEEnabled         bool      `json:"BLEEnabled,omitempty"`
	RTDEnabled         bool      `json:"RTDEnabled,omitempty"`
	WiFiRSSI           int64     `json:"WiFiRSSI,omitempty"`
	HasTime            bool      `json:"HasTime,omitempty"`
	HasResetMQTT       bool      `json:"HasResetMQTT,omitempty"`
	Timestamp          time.Time `json:"Timestamp,omitempty"`
}

func OnESPStatusMessage(client MQTT.Client, message MQTT.Message) {
	go handleESPStatusMessage(client, message)
}

func handleESPStatusMessage(client MQTT.Client, message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.ESPLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	espStatus := ESPStatus{}
	err := json.Unmarshal(message.Payload(), &espStatus)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}

	loc, ok := SharedSubscriptionConfig.MACtoLocation[strings.ToLower(espStatus.MAC)]
	if ok {
		espStatus.Location = loc
	} else {
		log.Warn().Msgf("Location not found for MAC %v", espStatus.MAC)
	}
	if espStatus.Timestamp.IsZero() {
		espStatus.Timestamp = time.Now()
	}

	espStatus.LogJSON()
	if SharedSubscriptionConfig.Repost {
		PublishClientMessage(client, SharedSubscriptionConfig.RepostRootTopic+"esp/status/"+espStatus.Location, espStatus.ToJSON())
	}
}

func (meas ESPStatus) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas ESPStatus) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("ESP Status: %v", json)
	if SharedSubscriptionConfig.ESPLogEn {
		log.Info().Msgf("ESP Status: %v", json)
	}
}
