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

func TestOutsideStruct(t *testing.T) {
	// Create an Outside instance
	now := time.Now()
	outside := Outside{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		TempF:        72.5,
		Pressure:     1013.25,
		PressureInHg: 29.92,
	}

	// Test ToJSON
	jsonData := outside.ToJSON()
	var parsedOutside Outside
	err := json.Unmarshal([]byte(jsonData), &parsedOutside)
	assert.NoError(t, err)
	assert.Equal(t, outside.Source, parsedOutside.Source)
	assert.Equal(t, outside.TempF, parsedOutside.TempF)
	assert.Equal(t, outside.Pressure, parsedOutside.Pressure)
	assert.Equal(t, outside.PressureInHg, parsedOutside.PressureInHg)

	// Test IsEmpty
	assert.False(t, outside.IsEmpty())

	emptyOutside := Outside{}
	assert.True(t, emptyOutside.IsEmpty())

	// Test GetInfluxFields
	fields := outside.GetInfluxFields()
	assert.Equal(t, outside.TempF, fields["TempF"])
	assert.Equal(t, outside.Pressure, fields["Pressure"])
	assert.Equal(t, outside.PressureInHg, fields["PressureInHg"])

	// Test GetMeasurementName
	assert.Equal(t, "outside", outside.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "environment/outside", outside.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.OutsideLogEn, outside.GetLogEnabled())

	// Test ToInfluxPoint
	point := outside.ToInfluxPoint()
	assert.NotNil(t, point)
}

func TestProcessOutsideData(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	tests := []struct {
		name        string
		measurement string
		rawData     map[string]any
		expected    *Outside
	}{
		{
			name:        "temperature measurement",
			measurement: "temperature",
			rawData: map[string]any{
				"value": 300.15, // 300.15K = 80.6°F
			},
			expected: &Outside{
				TempF: 80.6, // KelvinToFarenheit(300.15) = 80.6
			},
		},
		{
			name:        "pressure measurement",
			measurement: "pressure",
			rawData: map[string]any{
				"value": 101325.0, // 101325 Pa = 1013.25 mbar = 29.92 inHg
			},
			expected: &Outside{
				Pressure:     1013.25, // 101325 / 100 = 1013.25
				PressureInHg: 29.92,   // MillibarToInHg(1013.25) = 29.92
			},
		},
		{
			name:        "unknown measurement",
			measurement: "unknown",
			rawData: map[string]any{
				"value": 42.0,
			},
			expected: &Outside{},
		},
		{
			name:        "invalid temperature value",
			measurement: "temperature",
			rawData: map[string]any{
				"value": "not a number",
			},
			expected: &Outside{},
		},
		{
			name:        "invalid pressure value",
			measurement: "pressure",
			rawData: map[string]any{
				"value": "not a number",
			},
			expected: &Outside{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outside := &Outside{}
			processOutsideData(tt.rawData, tt.measurement, outside)

			switch tt.measurement {
			case "temperature":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.TempF, outside.TempF, 0.1)
				} else {
					assert.Equal(t, tt.expected.TempF, outside.TempF)
				}
			case "pressure":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.Pressure, outside.Pressure, 0.01)
					assert.InDelta(t, tt.expected.PressureInHg, outside.PressureInHg, 0.01)
				} else {
					assert.Equal(t, tt.expected.Pressure, outside.Pressure)
					assert.Equal(t, tt.expected.PressureInHg, outside.PressureInHg)
				}
			default:
				assert.Equal(t, tt.expected.TempF, outside.TempF)
				assert.Equal(t, tt.expected.Pressure, outside.Pressure)
				assert.Equal(t, tt.expected.PressureInHg, outside.PressureInHg)
			}
		})
	}
}

func TestHandleOutsideMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data for temperature
	tempData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     300.15, // 300.15K = 80.6°F
	}
	tempPayload, _ := json.Marshal(tempData)
	tempMessage := NewMockMessage("vessels/test/environment/outside/temperature", tempPayload)

	// Call OnOutsideMessage
	OnOutsideMessage(client, tempMessage)

	// Create test data for pressure
	pressureData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     101325.0, // 101325 Pa = 1013.25 mbar = 29.92 inHg
	}
	pressurePayload, _ := json.Marshal(pressureData)
	pressureMessage := NewMockMessage("vessels/test/environment/outside/pressure", pressurePayload)

	// Call OnOutsideMessage
	OnOutsideMessage(client, pressureMessage)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("vessels/test/environment/outside/temperature", []byte("invalid json"))
	OnOutsideMessage(client, invalidMessage)
}
