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
	"context"
	"encoding/json"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/rs/zerolog/log"
)

// SensorData defines the common interface for all sensor data types
type SensorData interface {
	// ToJSON serializes the data to JSON
	ToJSON() string
	// LogJSON logs the JSON representation of the data
	LogJSON()
	// IsEmpty checks if the data has any meaningful values
	IsEmpty() bool
	// GetInfluxTags returns tags for InfluxDB
	GetInfluxTags() map[string]string
	// GetInfluxFields returns fields for InfluxDB
	GetInfluxFields() map[string]interface{}
	// ToInfluxPoint creates an InfluxDB point
	ToInfluxPoint() *write.Point
	// GetSource returns the source of the data
	GetSource() string
	// GetTimestamp returns the timestamp of the data
	GetTimestamp() time.Time
	// SetSource sets the source of the data
	SetSource(source string)
	// SetTimestamp sets the timestamp of the data
	SetTimestamp(timestamp time.Time)
	// GetLogEnabled returns whether logging is enabled for this data type
	GetLogEnabled() bool
	// GetMeasurementName returns the measurement name for InfluxDB
	GetMeasurementName() string
	// GetTopicPrefix returns the topic prefix for MQTT publishing
	GetTopicPrefix() string
}

// BaseSensorData contains common fields and methods for all sensor data types
type BaseSensorData struct {
	Source    string    `json:"Source,omitempty"`
	Timestamp time.Time `json:"Timestamp,omitempty"`
}

// GetSource returns the source of the data
func (b BaseSensorData) GetSource() string {
	return b.Source
}

// GetTimestamp returns the timestamp of the data
func (b BaseSensorData) GetTimestamp() time.Time {
	return b.Timestamp
}

// SetSource sets the source of the data
func (b *BaseSensorData) SetSource(source string) {
	b.Source = source
}

// SetTimestamp sets the timestamp of the data
func (b *BaseSensorData) SetTimestamp(timestamp time.Time) {
	b.Timestamp = timestamp
}

// GetInfluxTags returns tags for InfluxDB
func (b BaseSensorData) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	if b.Source != "" {
		tagTmp["Source"] = b.Source
	}
	return tagTmp
}

// ParseCommonFields parses common fields from raw data
func ParseCommonFields(rawData map[string]any, data SensorData) {
	// Parse source
	strtmp, err := ParseString(rawData["$source"])
	if err != nil {
		log.Warn().Msgf("Error parsing string: %v", err.Error())
	} else {
		data.SetSource(strtmp)
	}

	// Parse timestamp
	strtmp, err = ParseString(rawData["timestamp"])
	if err != nil {
		log.Warn().Msgf("Error parsing string: %v", err.Error())
	} else {
		timestamp, err := time.Parse(ISOTimeLayout, strtmp)
		if err != nil {
			log.Warn().Msgf("Error parsing time string: %v", err.Error())
		} else {
			data.SetTimestamp(timestamp)
		}
	}

	// Set default timestamp if not provided
	if data.GetTimestamp().IsZero() {
		data.SetTimestamp(time.Now())
	}

	// Map source name if available
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(data.GetSource())]
	if ok {
		data.SetSource(name)
	} else {
		log.Warn().Msgf("Name not found for Source %v", data.GetSource())
	}
}

// HandleSensorMessage handles a sensor message
func HandleSensorMessage(client MQTT.Client, message MQTT.Message, data SensorData, handler func(map[string]any, string, SensorData)) {
	logEnabled := data.GetLogEnabled()

	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if logEnabled {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}

	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if logEnabled {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}

	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
		return
	}

	// Parse common fields
	ParseCommonFields(rawData, data)

	// Call the specific handler for this data type
	handler(rawData, measurement, data)

	// Skip empty data
	if data.IsEmpty() {
		return
	}

	// Log the data
	data.LogJSON()

	// Publish to MQTT if enabled
	if SharedSubscriptionConfig.Repost {
		PublishClientMessage(client,
			SharedSubscriptionConfig.RepostRootTopic+"vessel/"+data.GetTopicPrefix()+"/"+data.GetSource()+"/"+measurement,
			data.ToJSON(), true)
	}

	// Write to InfluxDB if enabled
	if SharedSubscriptionConfig.InfluxEnabled {
		p := data.ToInfluxPoint()
		err := SharedInfluxWriteAPI.WritePoint(context.Background(), p)
		if err != nil {
			log.Warn().Msgf("Error writing to influx: %v", err.Error())
		}
		log.Trace().Msg("Wrote Point")
		if logEnabled {
			log.Debug().Msg("Wrote Point")
		}
	}
}
