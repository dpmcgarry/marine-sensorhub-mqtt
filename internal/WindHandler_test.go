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

func TestWindStruct(t *testing.T) {
	// Create a Wind instance
	now := time.Now()
	wind := Wind{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		SpeedApp:      15.5,
		AngleApp:      45.0,
		SOG:           10.2,
		DirectionTrue: 180.0,
	}

	// Test ToJSON
	jsonData := wind.ToJSON()
	var parsedWind Wind
	err := json.Unmarshal([]byte(jsonData), &parsedWind)
	assert.NoError(t, err)
	assert.Equal(t, wind.Source, parsedWind.Source)
	assert.Equal(t, wind.SpeedApp, parsedWind.SpeedApp)
	assert.Equal(t, wind.AngleApp, parsedWind.AngleApp)
	assert.Equal(t, wind.SOG, parsedWind.SOG)
	assert.Equal(t, wind.DirectionTrue, parsedWind.DirectionTrue)

	// Test IsEmpty
	assert.False(t, wind.IsEmpty())

	emptyWind := Wind{}
	assert.True(t, emptyWind.IsEmpty())

	// Test GetInfluxFields
	fields := wind.GetInfluxFields()
	assert.Equal(t, wind.SpeedApp, fields["SpeedApp"])
	assert.Equal(t, wind.AngleApp, fields["AngleApp"])
	assert.Equal(t, wind.SOG, fields["SOG"])
	assert.Equal(t, wind.DirectionTrue, fields["DirectionTrue"])

	// Test GetMeasurementName
	assert.Equal(t, "wind", wind.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "environment/wind", wind.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.WindLogEn, wind.GetLogEnabled())
}

func TestProcessWindData(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	tests := []struct {
		name        string
		measurement string
		rawData     map[string]any
		expected    *Wind
	}{
		{
			name:        "speedOverGround measurement",
			measurement: "speedOverGround",
			rawData: map[string]any{
				"value": 5.0, // 5 m/s = 9.71922 knots
			},
			expected: &Wind{
				SOG: 9.71922, // MetersPerSecondToKnots(5.0) = 9.71922
			},
		},
		{
			name:        "directionTrue measurement",
			measurement: "directionTrue",
			rawData: map[string]any{
				"value": 3.14159, // π radians = 180 degrees
			},
			expected: &Wind{
				DirectionTrue: 180.0, // RadiansToDegrees(3.14159) = 180.0
			},
		},
		{
			name:        "speedApparent measurement",
			measurement: "speedApparent",
			rawData: map[string]any{
				"value": 10.0, // 10 m/s = 19.43844 knots
			},
			expected: &Wind{
				SpeedApp: 19.43844, // MetersPerSecondToKnots(10.0) = 19.43844
			},
		},
		{
			name:        "angleApparent measurement",
			measurement: "angleApparent",
			rawData: map[string]any{
				"value": 1.5708, // π/2 radians = 90 degrees
			},
			expected: &Wind{
				AngleApp: 90.0, // RadiansToDegrees(1.5708) = 90.0
			},
		},
		{
			name:        "unknown measurement",
			measurement: "unknown",
			rawData: map[string]any{
				"value": 42.0,
			},
			expected: &Wind{},
		},
		{
			name:        "invalid value type",
			measurement: "speedApparent",
			rawData: map[string]any{
				"value": "not a number",
			},
			expected: &Wind{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wind := &Wind{}
			processWindData(tt.rawData, tt.measurement, wind)

			switch tt.measurement {
			case "speedOverGround":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.SOG, wind.SOG, 0.001)
				} else {
					assert.Equal(t, tt.expected.SOG, wind.SOG)
				}
			case "directionTrue":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.DirectionTrue, wind.DirectionTrue, 0.001)
				} else {
					assert.Equal(t, tt.expected.DirectionTrue, wind.DirectionTrue)
				}
			case "speedApparent":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.SpeedApp, wind.SpeedApp, 0.001)
				} else {
					assert.Equal(t, tt.expected.SpeedApp, wind.SpeedApp)
				}
			case "angleApparent":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.AngleApp, wind.AngleApp, 0.001)
				} else {
					assert.Equal(t, tt.expected.AngleApp, wind.AngleApp)
				}
			}
		})
	}
}

func TestHandleWindMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data for speed apparent
	speedData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     10.0, // 10 m/s = 19.43844 knots
	}
	speedPayload, _ := json.Marshal(speedData)
	speedMessage := NewMockMessage("vessels/test/environment/wind/speedApparent", speedPayload)

	// Call OnWindMessage
	OnWindMessage(client, speedMessage)

	// Create test data for angle apparent
	angleData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     1.5708, // π/2 radians = 90 degrees
	}
	anglePayload, _ := json.Marshal(angleData)
	angleMessage := NewMockMessage("vessels/test/environment/wind/angleApparent", anglePayload)

	// Call OnWindMessage
	OnWindMessage(client, angleMessage)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("vessels/test/environment/wind/speedApparent", []byte("invalid json"))
	OnWindMessage(client, invalidMessage)
}
