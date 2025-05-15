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

// Steering represents steering sensor data
type Steering struct {
	BaseSensorData
	RudderAngle      float64 `json:"RudderAngle,omitempty"`
	AutopilotState   string  `json:"AutoPilotState,omitempty"`
	TargetHeadingMag float64 `json:"TargetHeading,omitempty"`
}

// OnSteeringMessage is called when a steering message is received
func OnSteeringMessage(client MQTT.Client, message MQTT.Message) {
	go handleSteeringMessage(client, message)
}

// handleSteeringMessage processes steering messages
func handleSteeringMessage(client MQTT.Client, message MQTT.Message) {
	steer := &Steering{}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]

	// Store the original measurement for potential modification later
	originalMeasurement := measurement

	// Use the common handler
	HandleSensorMessage(client, message, steer, func(rawData map[string]any, measurement string, data SensorData) {
		processSteeringData(rawData, measurement, data)
	})

	// Special case for state measurement - rename to autopilotState for reposting
	if SharedSubscriptionConfig.Repost && !steer.IsEmpty() && originalMeasurement == "state" {
		PublishClientMessage(client,
			SharedSubscriptionConfig.RepostRootTopic+"vessel/steering/"+steer.Source+"/autopilotState",
			steer.ToJSON(), true)
	}
}

// processSteeringData processes specific steering data fields
func processSteeringData(rawData map[string]any, measurement string, data SensorData) {
	steer, ok := data.(*Steering)
	if !ok {
		log.Error().Msg("Failed to cast data to Steering type")
		return
	}

	var err error
	var floatTmp float64
	var strtmp string

	switch measurement {
	case "rudderAngle":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			steer.RudderAngle = RadiansToDegrees(floatTmp)
		}
	case "autopilot":
		// Just a container topic, no data to process
		break
	case "state":
		strtmp, err = ParseString(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing string: %v", err.Error())
		} else {
			steer.AutopilotState = strtmp
		}
	case "target":
		// Just a container topic, no data to process
		break
	case "headingMagnetic":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			steer.TargetHeadingMag = RadiansToDegrees(floatTmp)
		}
	default:
		log.Warn().Msgf("Unknown measurement %v", measurement)
	}
}

// ToJSON serializes the data to JSON
func (meas *Steering) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

// LogJSON logs the JSON representation of the data
func (meas *Steering) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Steering: %v", json)
	if SharedSubscriptionConfig.SteerLogEn {
		log.Info().Msgf("Steering: %v", json)
	}
}

// IsEmpty checks if the data has any meaningful values
func (meas *Steering) IsEmpty() bool {
	if meas.RudderAngle == 0.0 && meas.AutopilotState == "" && meas.TargetHeadingMag == 0.0 {
		return true
	}
	return false
}

// GetInfluxFields returns fields for InfluxDB
func (meas *Steering) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.RudderAngle != 0.0 {
		measTmp["RudderAngle"] = meas.RudderAngle
	}
	if meas.AutopilotState != "" {
		measTmp["AutopilotState"] = meas.AutopilotState
	}
	if meas.TargetHeadingMag != 0.0 {
		measTmp["TargetHeading"] = meas.TargetHeadingMag
	}
	return measTmp
}

// ToInfluxPoint creates an InfluxDB point
func (meas *Steering) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("steering", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}

// GetLogEnabled returns whether logging is enabled for this data type
func (meas *Steering) GetLogEnabled() bool {
	return SharedSubscriptionConfig.SteerLogEn
}

// GetMeasurementName returns the measurement name for InfluxDB
func (meas *Steering) GetMeasurementName() string {
	return "steering"
}

// GetTopicPrefix returns the topic prefix for MQTT publishing
func (meas *Steering) GetTopicPrefix() string {
	return "steering"
}
