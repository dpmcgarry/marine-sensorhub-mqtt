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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaterStruct(t *testing.T) {
	// Create a Water instance
	now := time.Now()
	water := Water{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		TempF:                  72.5,
		DepthUnderTransducerFt: 15.3,
	}

	// Test ToJSON
	jsonData := water.ToJSON()
	var parsedWater Water
	err := json.Unmarshal([]byte(jsonData), &parsedWater)
	assert.NoError(t, err)
	assert.Equal(t, water.Source, parsedWater.Source)
	assert.Equal(t, water.TempF, parsedWater.TempF)
	assert.Equal(t, water.DepthUnderTransducerFt, parsedWater.DepthUnderTransducerFt)

	// Test IsEmpty
	assert.False(t, water.IsEmpty())

	emptyWater := Water{}
	assert.True(t, emptyWater.IsEmpty())

	// Test GetInfluxFields
	fields := water.GetInfluxFields()
	assert.Equal(t, water.TempF, fields["TempF"])
	assert.Equal(t, water.DepthUnderTransducerFt, fields["DepthUnderTransducerFt"])

	// Test GetMeasurementName
	assert.Equal(t, "water", water.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "environment/water", water.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.WaterLogEn, water.GetLogEnabled())
}

func TestProcessWaterData(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	tests := []struct {
		name        string
		measurement string
		rawData     map[string]any
		expected    *Water
	}{
		{
			name:        "temperature measurement",
			measurement: "temperature",
			rawData: map[string]any{
				"value": 300.15, // 27°C or 80.6°F in Kelvin
			},
			expected: &Water{
				TempF: 27.0, // KelvinToCelsius(300.15) = 27.0
			},
		},
		{
			name:        "belowTransducer measurement",
			measurement: "belowTransducer",
			rawData: map[string]any{
				"value": 5.0, // 5 meters = 16.4042 feet
			},
			expected: &Water{
				DepthUnderTransducerFt: 16.4042, // MetersToFeet(5.0) = 16.4042
			},
		},
		{
			name:        "unknown measurement",
			measurement: "unknown",
			rawData: map[string]any{
				"value": 42.0,
			},
			expected: &Water{},
		},
		{
			name:        "invalid value type",
			measurement: "temperature",
			rawData: map[string]any{
				"value": "not a number",
			},
			expected: &Water{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			water := &Water{}
			processWaterData(tt.rawData, tt.measurement, water)

			if tt.measurement == "temperature" {
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.TempF, water.TempF, 0.001)
				} else {
					assert.Equal(t, tt.expected.TempF, water.TempF)
				}
			} else if tt.measurement == "belowTransducer" {
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.DepthUnderTransducerFt, water.DepthUnderTransducerFt, 0.001)
				} else {
					assert.Equal(t, tt.expected.DepthUnderTransducerFt, water.DepthUnderTransducerFt)
				}
			}
		})
	}
}

func TestHandleWaterMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data for temperature
	tempData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     300.15, // 27°C or 80.6°F in Kelvin
	}
	tempPayload, _ := json.Marshal(tempData)
	tempMessage := NewMockMessage("vessels/test/environment/water/temperature", tempPayload)

	// Call OnWaterMessage
	OnWaterMessage(client, tempMessage)

	// Create test data for depth
	depthData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     5.0, // 5 meters = 16.4042 feet
	}
	depthPayload, _ := json.Marshal(depthData)
	depthMessage := NewMockMessage("vessels/test/environment/water/belowTransducer", depthPayload)

	// Call OnWaterMessage
	OnWaterMessage(client, depthMessage)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("vessels/test/environment/water/temperature", []byte("invalid json"))
	OnWaterMessage(client, invalidMessage)
}
