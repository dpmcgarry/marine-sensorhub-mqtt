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

// Water represents water sensor data
type Water struct {
	BaseSensorData
	TempF                  float64 `json:"TempF,omitempty"`
	DepthUnderTransducerFt float64 `json:"DepthUnderTransducerFt,omitempty"`
}

// OnWaterMessage is called when a water message is received
func OnWaterMessage(client MQTT.Client, message MQTT.Message) {
	go handleWaterMessage(client, message)
}

// handleWaterMessage processes water messages
func handleWaterMessage(client MQTT.Client, message MQTT.Message) {
	water := &Water{}
	HandleSensorMessage(client, message, water, processWaterData)
}

// processWaterData processes specific water data fields
func processWaterData(rawData map[string]any, measurement string, data SensorData) {
	water, ok := data.(*Water)
	if !ok {
		log.Error().Msg("Failed to cast data to Water type")
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
			// My sensor reports in F but SK assumes it is C
			// So Converting from K to C actually gives F
			water.TempF = KelvinToCelsius(floatTmp)
		}
	case "belowTransducer":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			water.DepthUnderTransducerFt = MetersToFeet(floatTmp)
		}
	default:
		log.Warn().Msgf("Unknown measurement %v", measurement)
	}
}

// ToJSON serializes the data to JSON
func (meas *Water) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *Water) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Water: %v", json)
	if SharedSubscriptionConfig.WaterLogEn {
		log.Info().Msgf("Water: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *Water) IsEmpty() bool {
	if meas.DepthUnderTransducerFt == 0.0 && meas.TempF == 0.0 {
		return true
	}
	return false
}

// GetInfluxFields returns fields for InfluxDB
func (meas *Water) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.TempF != 0.0 {
		measTmp["TempF"] = meas.TempF
	}
	if meas.DepthUnderTransducerFt != 0.0 {
		measTmp["DepthUnderTransducerFt"] = meas.DepthUnderTransducerFt
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *Water) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("water", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *Water) GetLogEnabled() bool {
	return SharedSubscriptionConfig.WaterLogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *Water) GetMeasurementName() string {
	return "water"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *Water) GetTopicPrefix() string {
	return "environment/water"
}
