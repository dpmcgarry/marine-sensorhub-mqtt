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
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/rs/zerolog/log"
)

// Propulsion represents propulsion system sensor data
type Propulsion struct {
	BaseSensorData
	Device           string  `json:"Device,omitempty"`
	RPM              int64   `json:"RPM,omitempty"`
	BoostPSI         float64 `json:"BoostPSI,omitempty"`
	OilTempF         float64 `json:"OilTempF,omitempty"`
	OilPressure      float64 `json:"OilPressure,omitempty"`
	CoolantTempF     float64 `json:"CoolantTempF,omitempty"`
	RunTime          int64   `json:"RunTime,omitempty"`
	EngineLoad       float64 `json:"EngineLoad,omitempty"`
	EngineTorque     float64 `json:"EngineTorque,omitempty"`
	TransOilTempF    float64 `json:"TransOilTemp,omitempty"`
	TransOilPressure float64 `json:"TransOilPressure,omitempty"`
	AltVoltage       float64 `json:"AlternatorVoltage,omitempty"`
	FuelRate         float64 `json:"FuelRate,omitempty"`
}

// OnPropulsionMessage is called when a propulsion message is received
func OnPropulsionMessage(client MQTT.Client, message MQTT.Message) {
	go handlePropulsionMessage(client, message)
}

// handlePropulsionMessage processes propulsion messages
func handlePropulsionMessage(client MQTT.Client, message MQTT.Message) {
	// TODO: Support Multiple Engines
	prop := &Propulsion{}

	// Check if this is a transmission message
	isTranny := false
	if strings.Contains(message.Topic(), "/transmission/") {
		isTranny = true
	}

	// Store this in a context that can be accessed by the processor
	context := map[string]interface{}{
		"isTranny": isTranny,
	}

	// Use the common handler with a custom processor that has access to the context
	HandleSensorMessage(client, message, prop, func(rawData map[string]any, measurement string, data SensorData) {
		processPropulsionData(rawData, measurement, data, context)
	})
}

// processPropulsionData processes specific propulsion data fields
func processPropulsionData(rawData map[string]any, measurement string, data SensorData, context map[string]interface{}) {
	prop, ok := data.(*Propulsion)
	if !ok {
		log.Error().Msg("Failed to cast data to Propulsion type")
		return
	}

	// Get the transmission flag from context
	isTranny, _ := context["isTranny"].(bool)

	var err error
	var floatTmp float64

	switch measurement {
	case "revolutions":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			prop.RPM = int64(floatTmp) * 60
		}
	case "boostPressure":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			prop.BoostPSI = PascalToPSI(floatTmp)
		}
	case "oilTemperature":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			if isTranny {
				prop.TransOilTempF = KelvinToFarenheit(floatTmp)
			} else {
				prop.OilTempF = KelvinToFarenheit(floatTmp)
			}
		}
	case "oilPressure":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			if isTranny {
				prop.TransOilPressure = PascalToPSI(floatTmp)
			} else {
				prop.OilPressure = PascalToPSI(floatTmp)
			}
		}
	case "temperature":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			prop.CoolantTempF = KelvinToFarenheit(floatTmp)
		}
	case "alternatorVoltage":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			prop.AltVoltage = floatTmp
		}
	case "transmission":
		// Just a container topic, no data to process
		break
	case "fuel":
		// Just a container topic, no data to process
		break
	case "rate":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			prop.FuelRate = CubicMetersPerSecondToGallonsPerHour(floatTmp)
		}
	case "runTime":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			prop.RunTime = int64(floatTmp)
		}
	case "engineLoad":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			prop.EngineLoad = floatTmp * 100
		}
	case "engineTorque":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			prop.EngineTorque = floatTmp * 100
		}
	default:
		log.Warn().Msgf("Unknown measurement %v", measurement)
	}
}

// ToJSON serializes the data to JSON
func (meas *Propulsion) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *Propulsion) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Propulsion: %v", json)
	if SharedSubscriptionConfig.PropLogEn {
		log.Info().Msgf("Propulsion: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *Propulsion) IsEmpty() bool {
	if meas.RPM == 0 && meas.BoostPSI == 0.0 && meas.OilTempF == 0.0 && meas.OilPressure == 0.0 &&
		meas.CoolantTempF == 0.0 && meas.RunTime == 0 && meas.EngineLoad == 0.0 && meas.EngineTorque == 0.0 &&
		meas.TransOilTempF == 0.0 && meas.TransOilPressure == 0.0 && meas.AltVoltage == 0.0 && meas.FuelRate == 0.0 {
		return true
	}
	return false
}

// GetInfluxTags returns tags for InfluxDB
func (meas *Propulsion) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	if meas.Source != "" {
		tagTmp["Source"] = meas.Source
	}
	if meas.Device != "" {
		tagTmp["Device"] = meas.Device
	}
	return tagTmp
}

// GetInfluxFields returns fields for InfluxDB
func (meas *Propulsion) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.RPM != 0 {
		measTmp["RPM"] = meas.RPM
	}
	if meas.BoostPSI != 0.0 {
		measTmp["BoostPSI"] = meas.BoostPSI
	}
	if meas.OilTempF != 0.0 {
		measTmp["OilTempF"] = meas.OilTempF
	}
	if meas.OilPressure != 0.0 {
		measTmp["OilPressure"] = meas.OilPressure
	}
	if meas.CoolantTempF != 0.0 {
		measTmp["CoolantTempF"] = meas.CoolantTempF
	}
	if meas.RunTime != 0 {
		measTmp["RunTime"] = meas.RunTime
	}
	if meas.EngineLoad != 0.0 {
		measTmp["EngineLoad"] = meas.EngineLoad
	}
	if meas.EngineTorque != 0.0 {
		measTmp["EngineTorque"] = meas.EngineTorque
	}
	if meas.TransOilTempF != 0.0 {
		measTmp["TransOilTempF"] = meas.TransOilTempF
	}
	if meas.TransOilPressure != 0.0 {
		measTmp["TransOilPressure"] = meas.TransOilPressure
	}
	if meas.AltVoltage != 0.0 {
		measTmp["AlternatorVoltage"] = meas.AltVoltage
	}
	if meas.FuelRate != 0.0 {
		measTmp["FuelRate"] = meas.FuelRate
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *Propulsion) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("propulsion", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *Propulsion) GetLogEnabled() bool {
	return SharedSubscriptionConfig.PropLogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *Propulsion) GetMeasurementName() string {
	return "propulsion"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *Propulsion) GetTopicPrefix() string {
	return "propulsion"
}
