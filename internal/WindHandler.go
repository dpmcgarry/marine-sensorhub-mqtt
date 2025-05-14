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

// Wind represents wind sensor data
type Wind struct {
	BaseSensorData
	SpeedApp      float64 `json:"SpeedApp,omitempty"`
	AngleApp      float64 `json:"AngleApp,omitempty"`
	SOG           float64 `json:"SOG,omitempty"`
	DirectionTrue float64 `json:"DirectionTrue,omitempty"`
}

// OnWindMessage is called when a wind message is received
func OnWindMessage(client MQTT.Client, message MQTT.Message) {
	go handleWindMessage(client, message)
}

// handleWindMessage processes wind messages
func handleWindMessage(client MQTT.Client, message MQTT.Message) {
	wind := &Wind{}
	HandleSensorMessage(client, message, wind, processWindData)
}

// processWindData processes specific wind data fields
func processWindData(rawData map[string]any, measurement string, data SensorData) {
	wind, ok := data.(*Wind)
	if !ok {
		log.Error().Msg("Failed to cast data to Wind type")
		return
	}

	var err error
	var floatTmp float64

	switch measurement {
	case "speedOverGround":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			wind.SOG = MetersPerSecondToKnots(floatTmp)
		}
	case "directionTrue":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			wind.DirectionTrue = RadiansToDegrees(floatTmp)
		}
	case "speedApparent":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			wind.SpeedApp = MetersPerSecondToKnots(floatTmp)
		}
	case "angleApparent":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			wind.AngleApp = RadiansToDegrees(floatTmp)
		}
	default:
		log.Warn().Msgf("Unknown measurement %v", measurement)
	}
}

// ToJSON serializes the data to JSON
func (meas *Wind) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *Wind) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Wind: %v", json)
	if SharedSubscriptionConfig.WindLogEn {
		log.Info().Msgf("Wind: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *Wind) IsEmpty() bool {
	if meas.SpeedApp == 0.0 && meas.AngleApp == 0.0 && meas.SOG == 0.0 && meas.DirectionTrue == 0.0 {
		return true
	}
	return false
}

// GetInfluxFields returns fields for InfluxDB
func (meas *Wind) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.SpeedApp != 0.0 {
		measTmp["SpeedApp"] = meas.SpeedApp
	}
	if meas.AngleApp != 0.0 {
		measTmp["AngleApp"] = meas.AngleApp
	}
	if meas.SOG != 0.0 {
		measTmp["SOG"] = meas.SOG
	}
	if meas.DirectionTrue != 0.0 {
		measTmp["DirectionTrue"] = meas.DirectionTrue
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *Wind) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("wind", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *Wind) GetLogEnabled() bool {
	return SharedSubscriptionConfig.WindLogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *Wind) GetMeasurementName() string {
	return "wind"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *Wind) GetTopicPrefix() string {
	return "environment/wind"
}
