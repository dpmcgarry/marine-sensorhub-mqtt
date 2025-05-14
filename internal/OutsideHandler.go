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

// Outside represents outside environment sensor data
type Outside struct {
	BaseSensorData
	TempF        float64 `json:"TempF,omitempty"`
	Pressure     float64 `json:"Pressure,omitempty"`
	PressureInHg float64 `json:"PressureInHg,omitempty"`
}

// OnOutsideMessage is called when an outside environment message is received
func OnOutsideMessage(client MQTT.Client, message MQTT.Message) {
	go handleOutsideMessage(client, message)
}

// handleOutsideMessage processes outside environment messages
func handleOutsideMessage(client MQTT.Client, message MQTT.Message) {
	out := &Outside{}
	HandleSensorMessage(client, message, out, processOutsideData)
}

// processOutsideData processes specific outside environment data fields
func processOutsideData(rawData map[string]any, measurement string, data SensorData) {
	out, ok := data.(*Outside)
	if !ok {
		log.Error().Msg("Failed to cast data to Outside type")
		return
	}

	var err error
	var floatTmp float64

	switch measurement {
	case "temperature":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			out.TempF = KelvinToFarenheit(floatTmp)
		}
	case "pressure":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			out.Pressure = floatTmp / 100
			out.PressureInHg = MillibarToInHg(out.Pressure)
		}
	default:
		log.Warn().Msgf("Unknown measurement %v", measurement)
	}
}

// ToJSON serializes the data to JSON
func (meas *Outside) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *Outside) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Outside: %v", json)
	if SharedSubscriptionConfig.OutsideLogEn {
		log.Info().Msgf("Outside: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *Outside) IsEmpty() bool {
	if meas.TempF == 0.0 && meas.Pressure == 0.0 && meas.PressureInHg == 0.0 {
		return true
	}
	return false
}

// GetInfluxFields returns fields for InfluxDB
func (meas *Outside) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.TempF != 0.0 {
		measTmp["TempF"] = meas.TempF
	}
	if meas.Pressure != 0.0 {
		measTmp["Pressure"] = meas.Pressure
	}
	if meas.PressureInHg != 0.0 {
		measTmp["PressureInHg"] = meas.PressureInHg
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *Outside) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("outside", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *Outside) GetLogEnabled() bool {
	return SharedSubscriptionConfig.OutsideLogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *Outside) GetMeasurementName() string {
	return "outside"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *Outside) GetTopicPrefix() string {
	return "environment/outside"
}
