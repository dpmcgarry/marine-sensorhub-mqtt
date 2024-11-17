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

type BLETemperature struct {
	MAC            string
	Location       string `json:"Location,omitempty"`
	TempF          float64
	BatteryPercent float64 `json:"BatteryPct,omitempty"`
	Humidity       float64 `json:"Humidity,omitempty"`
	RSSI           int64
	Timestamp      time.Time
}

func OnBLETemperatureMessage(client MQTT.Client, message MQTT.Message) {
	log.Debug().Msgf("Got a message from: %v", message.Topic())
	bleTemp := BLETemperature{}
	err := json.Unmarshal(message.Payload(), &bleTemp)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}

	loc, ok := SharedSubscriptionConfig.MACtoLocation[bleTemp.MAC]
	if ok {
		bleTemp.Location = loc
	}
	if bleTemp.Timestamp.IsZero() {
		bleTemp.Timestamp = time.Now()
	}
	bleTemp.LogJSON()
}

func (meas BLETemperature) LogJSON() {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	log.Debug().Msgf("BLETemp: %v", string(jsonData))
}
