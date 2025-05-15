/*
Copyright © 2025 Don P. McGarry

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

	MQTT "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/rs/zerolog/log"
)

// ESPStatus represents ESP32 device status data
type ESPStatus struct {
	BaseSensorData
	MAC                string `json:"MAC,omitempty"`
	Location           string `json:"Location,omitempty"`
	IPAddress          string `json:"IPAddress,omitempty"`
	MSHVersion         string `json:"MSHVersion,omitempty"`
	FreeSRAM           int64  `json:"FreeSRAM,omitempty"`
	FreeHeap           int64  `json:"FreeHeap,omitempty"`
	FreePSRAM          int64  `json:"FreePSRAM,omitempty"`
	WiFiReconnectCount int64  `json:"WiFiReconnectCount,omitempty"`
	MQTTReconnectCount int64  `json:"MQTTReconnectCount,omitempty"`
	BLEEnabled         bool   `json:"BLEEnabled,omitempty"`
	RTDEnabled         bool   `json:"RTDEnabled,omitempty"`
	WiFiRSSI           int64  `json:"WiFiRSSI,omitempty"`
	HasTime            bool   `json:"HasTime,omitempty"`
	HasResetMQTT       bool   `json:"HasResetMQTT,omitempty"`
}

// OnESPStatusMessage is called when an ESP status message is received
func OnESPStatusMessage(client MQTT.Client, message MQTT.Message) {
	go handleESPStatusMessage(client, message)
}

// handleESPStatusMessage processes ESP status messages
func handleESPStatusMessage(client MQTT.Client, message MQTT.Message) {
	espStatus := &ESPStatus{}
	HandleJSONMessage(client, message, espStatus)

	// Map MAC to location after unmarshalling
	if espStatus.MAC != "" {
		loc := MapMACToLocation(espStatus, espStatus.MAC)
		if loc != "" {
			espStatus.Location = loc
		}
	}
	SendJSONMessage(client, message, espStatus)
}

// ToJSON serializes the data to JSON
func (meas *ESPStatus) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *ESPStatus) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("ESP Status: %v", json)
	if SharedSubscriptionConfig.ESPLogEn {
		log.Info().Msgf("ESP Status: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *ESPStatus) IsEmpty() bool {
	// ESP status data is always considered valid if we received it
	return false
}

// GetInfluxTags returns tags for InfluxDB
func (meas *ESPStatus) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	tagTmp["MAC"] = meas.MAC
	if meas.Location != "" {
		tagTmp["Location"] = meas.Location
	}
	if meas.IPAddress != "" {
		tagTmp["IPAddress"] = meas.IPAddress
	}
	if meas.MSHVersion != "" {
		tagTmp["MSHVersion"] = meas.MSHVersion
	}
	return tagTmp
}

// GetInfluxFields returns fields for InfluxDB
func (meas *ESPStatus) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.FreeSRAM != 0 {
		measTmp["FreeSRAM"] = meas.FreeSRAM
	}
	if meas.FreeHeap != 0 {
		measTmp["FreeHeap"] = meas.FreeHeap
	}
	if meas.FreePSRAM != 0 {
		measTmp["FreePSRAM"] = meas.FreePSRAM
	}
	// These being zero is a valid case so just always include them
	measTmp["WiFiReconnectCount"] = meas.WiFiReconnectCount
	measTmp["MQTTReconnectCount"] = meas.MQTTReconnectCount
	if meas.BLEEnabled {
		measTmp["BLEEnabled"] = 1
	} else {
		measTmp["BLEEnabled"] = 0
	}
	if meas.RTDEnabled {
		measTmp["RTDEnabled"] = 1
	} else {
		measTmp["RTDEnabled"] = 0
	}
	if meas.HasTime {
		measTmp["HasTime"] = 1
	} else {
		measTmp["HasTime"] = 0
	}
	if meas.HasResetMQTT {
		measTmp["HasResetMQTT"] = 1
	} else {
		measTmp["HasResetMQTT"] = 0
	}
	if meas.WiFiRSSI != 0 {
		measTmp["WiFiRSSI"] = meas.WiFiRSSI
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *ESPStatus) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("espStatus", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetSource returns the source of the data
func (meas *ESPStatus) GetSource() string {
	return meas.MAC
}

// SetSource sets the source of the data
func (meas *ESPStatus) SetSource(source string) {
	meas.MAC = source
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *ESPStatus) GetLogEnabled() bool {
	return SharedSubscriptionConfig.ESPLogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *ESPStatus) GetMeasurementName() string {
	return "espStatus"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *ESPStatus) GetTopicPrefix() string {
	return "esp/status"
}
