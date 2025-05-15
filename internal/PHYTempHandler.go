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

// PHYTemperature represents physical temperature sensor data
type PHYTemperature struct {
	BaseSensorData
	MAC       string  `json:"MAC,omitempty"`
	Location  string  `json:"Location,omitempty"`
	Device    string  `json:"Device,omitempty"`
	Component string  `json:"Component,omitempty"`
	TempF     float64 `json:"TempF,omitempty"`
}

// OnPHYTemperatureMessage is called when a physical temperature message is received
func OnPHYTemperatureMessage(client MQTT.Client, message MQTT.Message) {
	go handlePHYTemperatureMessage(client, message)
}

// handlePHYTemperatureMessage processes physical temperature messages
func handlePHYTemperatureMessage(client MQTT.Client, message MQTT.Message) {
	phyTemp := &PHYTemperature{}
	HandleJSONMessage(client, message, phyTemp)

	// Map MAC to location after unmarshalling
	if phyTemp.MAC != "" {
		loc := MapMACToLocation(phyTemp, phyTemp.MAC)
		if loc != "" {
			phyTemp.Location = loc
		}
	}

	SendJSONMessage(client, message, phyTemp)
}

// ToJSON serializes the data to JSON
func (meas *PHYTemperature) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *PHYTemperature) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Physical Temp: %v", json)
	if SharedSubscriptionConfig.PHYLogEn {
		log.Info().Msgf("Physical Temp: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *PHYTemperature) IsEmpty() bool {
	return meas.TempF == 0.0
}

// GetInfluxTags returns tags for InfluxDB
func (meas *PHYTemperature) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	tagTmp["MAC"] = meas.MAC
	if meas.Location != "" {
		tagTmp["Location"] = meas.Location
	}
	if meas.Device != "" {
		tagTmp["Device"] = meas.Device
	}
	if meas.Component != "" {
		tagTmp["Component"] = meas.Component
	}
	return tagTmp
}

// GetInfluxFields returns fields for InfluxDB
func (meas *PHYTemperature) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.TempF != 0.0 {
		measTmp["TempF"] = meas.TempF
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *PHYTemperature) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("phyTemperature", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetSource returns the source of the data
func (meas *PHYTemperature) GetSource() string {
	return meas.MAC
}

// SetSource sets the source of the data
func (meas *PHYTemperature) SetSource(source string) {
	meas.MAC = source
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *PHYTemperature) GetLogEnabled() bool {
	return SharedSubscriptionConfig.PHYLogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *PHYTemperature) GetMeasurementName() string {
	return "phyTemperature"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *PHYTemperature) GetTopicPrefix() string {
	return "rtd/temperature"
}
