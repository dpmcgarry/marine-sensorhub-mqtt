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

// GNSS represents GNSS sensor data
type GNSS struct {
	BaseSensorData
	AntennaAlt    float64 `json:"AntennaAlt,omitempty"`
	Satellites    int64   `json:"Satellites,omitempty"`
	HozDilution   float64 `json:"HozDilution,omitempty"`
	PosDilution   float64 `json:"PosDilution,omitempty"`
	GeoidalSep    float64 `json:"GeoidalSep,omitempty"`
	Type          string  `json:"Type,omitempty"`
	MethodQuality string  `json:"MethodQuality,omitempty"`
}

// OnGNSSMessage is called when a GNSS message is received
func OnGNSSMessage(client MQTT.Client, message MQTT.Message) {
	go handleGNSSMessage(client, message)
}

// handleGNSSMessage processes GNSS messages
func handleGNSSMessage(client MQTT.Client, message MQTT.Message) {
	gnss := &GNSS{}
	HandleSensorMessage(client, message, gnss, processGNSSData)
}

// processGNSSData processes specific GNSS data fields
func processGNSSData(rawData map[string]any, measurement string, data SensorData) {
	gnss, ok := data.(*GNSS)
	if !ok {
		log.Error().Msg("Failed to cast data to GNSS type")
		return
	}

	var err error
	var floatTmp float64
	var strtmp string

	switch measurement {
	case "antennaAltitude":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			gnss.AntennaAlt = floatTmp
		}
	case "satellites":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			gnss.Satellites = int64(floatTmp)
		}
	case "horizontalDilution":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			gnss.HozDilution = floatTmp
		}
	case "positionDilution":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			gnss.PosDilution = floatTmp
		}
	case "geoidalSeparation":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			gnss.GeoidalSep = floatTmp
		}
	case "type":
		strtmp, err = ParseString(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing string: %v", err.Error())
		} else {
			gnss.Type = strtmp
		}
	case "methodQuality":
		strtmp, err = ParseString(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing string: %v", err.Error())
		} else {
			gnss.MethodQuality = strtmp
		}
	case "integrity":
		break
	case "satellitesInView":
		break
	default:
		log.Warn().Msgf("Unknown measurement %v", measurement)
	}
}

// ToJSON serializes the data to JSON
func (meas *GNSS) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *GNSS) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("GNSS: %v", json)
	if SharedSubscriptionConfig.GNSSLogEn {
		log.Info().Msgf("GNSS: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *GNSS) IsEmpty() bool {
	if meas.AntennaAlt == 0.0 && meas.Satellites == 0 && meas.HozDilution == 0.0 && meas.PosDilution == 0.0 &&
		meas.GeoidalSep == 0.0 && meas.Type == "" && meas.MethodQuality == "" {
		return true
	}
	return false
}

// GetInfluxFields returns fields for InfluxDB
func (meas *GNSS) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.AntennaAlt != 0.0 {
		measTmp["AntennaAlt"] = meas.AntennaAlt
	}
	if meas.Satellites != 0 {
		measTmp["Satellites"] = meas.Satellites
	}
	if meas.HozDilution != 0.0 {
		measTmp["HozDilution"] = meas.HozDilution
	}
	if meas.PosDilution != 0.0 {
		measTmp["PosDilution"] = meas.PosDilution
	}
	if meas.GeoidalSep != 0.0 {
		measTmp["GeoidalSep"] = meas.GeoidalSep
	}
	if meas.Type != "" {
		measTmp["Type"] = meas.Type
	}
	if meas.MethodQuality != "" {
		measTmp["MethodQuality"] = meas.MethodQuality
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *GNSS) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("gnss", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *GNSS) GetLogEnabled() bool {
	return SharedSubscriptionConfig.GNSSLogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *GNSS) GetMeasurementName() string {
	return "gnss"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *GNSS) GetTopicPrefix() string {
	return "gnss"
}
