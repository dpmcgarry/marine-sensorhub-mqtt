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

func TestSteeringStruct(t *testing.T) {
	// Create a Steering instance
	now := time.Now()
	steering := Steering{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		RudderAngle:      15.5,
		AutopilotState:   "engaged",
		TargetHeadingMag: 180.0,
	}

	// Test ToJSON
	jsonData := steering.ToJSON()
	var parsedSteering Steering
	err := json.Unmarshal([]byte(jsonData), &parsedSteering)
	assert.NoError(t, err)
	assert.Equal(t, steering.Source, parsedSteering.Source)
	assert.Equal(t, steering.RudderAngle, parsedSteering.RudderAngle)
	assert.Equal(t, steering.AutopilotState, parsedSteering.AutopilotState)
	assert.Equal(t, steering.TargetHeadingMag, parsedSteering.TargetHeadingMag)

	// Test IsEmpty
	assert.False(t, steering.IsEmpty())

	emptySteering := Steering{}
	assert.True(t, emptySteering.IsEmpty())

	// Test GetInfluxFields
	fields := steering.GetInfluxFields()
	assert.Equal(t, steering.RudderAngle, fields["RudderAngle"])
	assert.Equal(t, steering.AutopilotState, fields["AutopilotState"])
	assert.Equal(t, steering.TargetHeadingMag, fields["TargetHeading"])

	// Test GetMeasurementName
	assert.Equal(t, "steering", steering.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "steering", steering.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.SteerLogEn, steering.GetLogEnabled())

	// Test ToInfluxPoint
	point := steering.ToInfluxPoint()
	assert.NotNil(t, point)
}

func TestProcessSteeringData(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	tests := []struct {
		name        string
		measurement string
		rawData     map[string]any
		expected    *Steering
	}{
		{
			name:        "rudderAngle measurement",
			measurement: "rudderAngle",
			rawData: map[string]any{
				"value": 0.2705, // 0.2705 radians = 15.5 degrees
			},
			expected: &Steering{
				RudderAngle: 15.5, // RadiansToDegrees(0.2705) = 15.5
			},
		},
		{
			name:        "autopilot measurement (container topic)",
			measurement: "autopilot",
			rawData: map[string]any{
				"value": "test",
			},
			expected: &Steering{},
		},
		{
			name:        "state measurement",
			measurement: "state",
			rawData: map[string]any{
				"value": "engaged",
			},
			expected: &Steering{
				AutopilotState: "engaged",
			},
		},
		{
			name:        "target measurement (container topic)",
			measurement: "target",
			rawData: map[string]any{
				"value": "test",
			},
			expected: &Steering{},
		},
		{
			name:        "headingMagnetic measurement",
			measurement: "headingMagnetic",
			rawData: map[string]any{
				"value": 3.14159, // π radians = 180 degrees
			},
			expected: &Steering{
				TargetHeadingMag: 180.0, // RadiansToDegrees(3.14159) = 180.0
			},
		},
		{
			name:        "unknown measurement",
			measurement: "unknown",
			rawData: map[string]any{
				"value": 42.0,
			},
			expected: &Steering{},
		},
		{
			name:        "invalid rudderAngle value",
			measurement: "rudderAngle",
			rawData: map[string]any{
				"value": "not a number",
			},
			expected: &Steering{},
		},
		{
			name:        "invalid state value",
			measurement: "state",
			rawData: map[string]any{
				"value": 123,
			},
			expected: &Steering{},
		},
		{
			name:        "invalid headingMagnetic value",
			measurement: "headingMagnetic",
			rawData: map[string]any{
				"value": "not a number",
			},
			expected: &Steering{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steering := &Steering{}
			processSteeringData(tt.rawData, tt.measurement, steering)

			switch tt.measurement {
			case "rudderAngle":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.RudderAngle, steering.RudderAngle, 0.1)
				} else {
					assert.Equal(t, tt.expected.RudderAngle, steering.RudderAngle)
				}
			case "state":
				assert.Equal(t, tt.expected.AutopilotState, steering.AutopilotState)
			case "headingMagnetic":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.TargetHeadingMag, steering.TargetHeadingMag, 0.1)
				} else {
					assert.Equal(t, tt.expected.TargetHeadingMag, steering.TargetHeadingMag)
				}
			case "autopilot", "target":
				// These are container topics, no data to process
				assert.Equal(t, tt.expected.RudderAngle, steering.RudderAngle)
				assert.Equal(t, tt.expected.AutopilotState, steering.AutopilotState)
				assert.Equal(t, tt.expected.TargetHeadingMag, steering.TargetHeadingMag)
			default:
				assert.Equal(t, tt.expected.RudderAngle, steering.RudderAngle)
				assert.Equal(t, tt.expected.AutopilotState, steering.AutopilotState)
				assert.Equal(t, tt.expected.TargetHeadingMag, steering.TargetHeadingMag)
			}
		})
	}
}

func TestHandleSteeringMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data for rudderAngle
	rudderData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     0.2705, // 0.2705 radians = 15.5 degrees
	}
	rudderPayload, _ := json.Marshal(rudderData)
	rudderMessage := NewMockMessage("vessels/test/steering/rudderAngle", rudderPayload)

	// Call OnSteeringMessage
	OnSteeringMessage(client, rudderMessage)

	// Create test data for state (special case with reposting)
	stateData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     "engaged",
	}
	statePayload, _ := json.Marshal(stateData)
	stateMessage := NewMockMessage("vessels/test/steering/state", statePayload)

	// Call OnSteeringMessage
	OnSteeringMessage(client, stateMessage)

	// Create test data for headingMagnetic
	headingData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     3.14159, // π radians = 180 degrees
	}
	headingPayload, _ := json.Marshal(headingData)
	headingMessage := NewMockMessage("vessels/test/steering/headingMagnetic", headingPayload)

	// Call OnSteeringMessage
	OnSteeringMessage(client, headingMessage)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("vessels/test/steering/rudderAngle", []byte("invalid json"))
	OnSteeringMessage(client, invalidMessage)
}
