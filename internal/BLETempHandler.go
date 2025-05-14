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

// BLETemperature represents BLE temperature sensor data
type BLETemperature struct {
	BaseSensorData
	MAC            string  `json:"MAC,omitempty"`
	Location       string  `json:"Location,omitempty"`
	TempF          float64 `json:"TempF,omitempty"`
	BatteryPercent float64 `json:"BatteryPct,omitempty"`
	Humidity       float64 `json:"Humidity,omitempty"`
	RSSI           int64   `json:"RSSI,omitempty"`
}

// OnBLETemperatureMessage is called when a BLE temperature message is received
func OnBLETemperatureMessage(client MQTT.Client, message MQTT.Message) {
	go handleBLETemperatureMessage(client, message)
}

// handleBLETemperatureMessage processes BLE temperature messages
func handleBLETemperatureMessage(client MQTT.Client, message MQTT.Message) {
	bleTemp := &BLETemperature{}
	HandleJSONMessage(client, message, bleTemp)

	// Map MAC to location after unmarshalling
	if bleTemp.MAC != "" {
		loc := MapMACToLocation(bleTemp, bleTemp.MAC)
		if loc != "" {
			bleTemp.Location = loc
		}
	}

	SendJSONMessage(client, message, bleTemp)
}

// ToJSON serializes the data to JSON
func (meas *BLETemperature) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *BLETemperature) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("BLETemp: %v", json)
	if SharedSubscriptionConfig.BLELogEn {
		log.Info().Msgf("BLETemp: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *BLETemperature) IsEmpty() bool {
	// BLE temperature data is always considered valid if we received it
	return false
}

// GetInfluxTags returns tags for InfluxDB
func (meas *BLETemperature) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	tagTmp["MAC"] = meas.MAC
	if meas.Location != "" {
		tagTmp["Location"] = meas.Location
	}
	return tagTmp
}

// GetInfluxFields returns fields for InfluxDB
func (meas *BLETemperature) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.TempF != 0.0 {
		measTmp["TempF"] = meas.TempF
	}
	if meas.BatteryPercent != 0.0 {
		measTmp["BatteryPercent"] = meas.BatteryPercent
	}
	if meas.Humidity != 0.0 {
		measTmp["Humidity"] = meas.Humidity
	}
	if meas.RSSI != 0 {
		measTmp["RSSI"] = meas.RSSI
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *BLETemperature) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("bleTemperature", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetSource returns the source of the data
func (meas *BLETemperature) GetSource() string {
	return meas.MAC
}

// SetSource sets the source of the data
func (meas *BLETemperature) SetSource(source string) {
	meas.MAC = source
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *BLETemperature) GetLogEnabled() bool {
	return SharedSubscriptionConfig.BLELogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *BLETemperature) GetMeasurementName() string {
	return "bleTemperature"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *BLETemperature) GetTopicPrefix() string {
	return "ble/temperature"
}
