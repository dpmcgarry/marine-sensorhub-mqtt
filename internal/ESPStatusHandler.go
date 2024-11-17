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
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type ESPStatus struct {
	MAC                string
	Location           string `json:"Location,omitempty"`
	IPAddress          string
	MSHVersion         string
	FreeSRAM           int64 `json:"FreeSRAM,omitempty"`
	FreeHeap           int64 `json:"FreeHeap,omitempty"`
	FreePSRAM          int64 `json:"FreePSRAM,omitempty"`
	WiFiReconnectCount int64
	MQTTReconnectCount int64
	BLEEnabled         bool
	RTDEnabled         bool
	WiFiRSSI           int64
	HasTime            bool
	MasResetMQTT       bool
	Timestamp          time.Time
}

func OnESPStatusMessage(client MQTT.Client, message MQTT.Message) {
	log.Debug().Msgf("Got a message from: %v", message.Topic())
	espStatus := ESPStatus{}
	err := json.Unmarshal(message.Payload(), &espStatus)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}

	loc, ok := SharedSubscriptionConfig.MACtoLocation[espStatus.MAC]
	if ok {
		espStatus.Location = loc
	}
	if espStatus.Timestamp.IsZero() {
		espStatus.Timestamp = time.Now()
	}

	espStatus.LogJSON()
}

func (meas ESPStatus) LogJSON() {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	log.Debug().Msgf("BLETemp: %v", string(jsonData))
}
