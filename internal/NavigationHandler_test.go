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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNavigationStruct(t *testing.T) {
	// Create a Navigation instance
	now := time.Now()
	nav := Navigation{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		Lat:          37.7749,
		Lon:          -122.4194,
		Alt:          100.5,
		SOG:          10.5,
		ROT:          2.5,
		COGTrue:      180.0,
		HeadingMag:   175.0,
		MagVariation: 15.0,
		MagDeviation: 2.0,
		Yaw:          5.0,
		Pitch:        1.5,
		Roll:         3.0,
		HeadingTrue:  178.0,
		STW:          9.8,
	}

	// Test ToJSON
	jsonData := nav.ToJSON()
	var parsedNav Navigation
	err := json.Unmarshal([]byte(jsonData), &parsedNav)
	assert.NoError(t, err)
	assert.Equal(t, nav.Source, parsedNav.Source)
	assert.Equal(t, nav.Lat, parsedNav.Lat)
	assert.Equal(t, nav.Lon, parsedNav.Lon)
	assert.Equal(t, nav.Alt, parsedNav.Alt)
	assert.Equal(t, nav.SOG, parsedNav.SOG)
	assert.Equal(t, nav.ROT, parsedNav.ROT)
	assert.Equal(t, nav.COGTrue, parsedNav.COGTrue)
	assert.Equal(t, nav.HeadingMag, parsedNav.HeadingMag)
	assert.Equal(t, nav.MagVariation, parsedNav.MagVariation)
	assert.Equal(t, nav.MagDeviation, parsedNav.MagDeviation)
	assert.Equal(t, nav.Yaw, parsedNav.Yaw)
	assert.Equal(t, nav.Pitch, parsedNav.Pitch)
	assert.Equal(t, nav.Roll, parsedNav.Roll)
	assert.Equal(t, nav.HeadingTrue, parsedNav.HeadingTrue)
	assert.Equal(t, nav.STW, parsedNav.STW)

	// Test IsEmpty
	assert.False(t, nav.IsEmpty())

	emptyNav := Navigation{}
	assert.True(t, emptyNav.IsEmpty())

	// Test GetInfluxFields
	fields := nav.GetInfluxFields()
	assert.Equal(t, nav.Lat, fields["Latitude"])
	assert.Equal(t, nav.Lon, fields["Longitude"])
	assert.Equal(t, nav.Alt, fields["Altitude"])
	assert.Equal(t, nav.SOG, fields["SpeedOverGround"])
	assert.Equal(t, nav.ROT, fields["RateOfTurn"])
	assert.Equal(t, nav.COGTrue, fields["CourseOverGroundTrue"])
	assert.Equal(t, nav.HeadingMag, fields["HeadingMagnetic"])
	assert.Equal(t, nav.MagVariation, fields["MagneticVariation"])
	assert.Equal(t, nav.MagDeviation, fields["MagneticDeviation"])
	assert.Equal(t, nav.Yaw, fields["Yaw"])
	assert.Equal(t, nav.Pitch, fields["Pitch"])
	assert.Equal(t, nav.Roll, fields["Roll"])
	assert.Equal(t, nav.HeadingTrue, fields["HeadingTrue"])
	assert.Equal(t, nav.STW, fields["SpeedThroughWater"])

	// Test GetMeasurementName
	assert.Equal(t, "navigation", nav.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "navigation", nav.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.NavLogEn, nav.GetLogEnabled())

	// Test ToInfluxPoint
	point := nav.ToInfluxPoint()
	assert.NotNil(t, point)
}

func TestProcessNavigationData(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	tests := []struct {
		name        string
		measurement string
		rawData     map[string]any
		source      string
		expected    *Navigation
	}{
		{
			name:        "headingMagnetic measurement",
			measurement: "headingMagnetic",
			rawData: map[string]any{
				"value": 3.14159, // π radians = 180 degrees
			},
			source: "test-source",
			expected: &Navigation{
				HeadingMag: 180.0, // RadiansToDegrees(3.14159) = 180.0
			},
		},
		{
			name:        "rateOfTurn measurement",
			measurement: "rateOfTurn",
			rawData: map[string]any{
				"value": 0.0872665, // 0.0872665 radians = 5 degrees
			},
			source: "test-source",
			expected: &Navigation{
				ROT: 5.0, // RadiansToDegrees(0.0872665) = 5.0
			},
		},
		{
			name:        "speedOverGround measurement",
			measurement: "speedOverGround",
			rawData: map[string]any{
				"value": 5.0, // 5 m/s = 9.71922 knots
			},
			source: "test-source",
			expected: &Navigation{
				SOG: 9.71922, // MetersPerSecondToKnots(5.0) = 9.71922
			},
		},
		{
			name:        "position measurement",
			measurement: "position",
			rawData: map[string]any{
				"value": map[string]any{
					"latitude":  37.7749,
					"longitude": -122.4194,
					"altitude":  30.48, // 30.48 meters = 100 feet
				},
			},
			source: "test-source",
			expected: &Navigation{
				Lat: 37.7749,
				Lon: -122.4194,
				Alt: 100.0, // MetersToFeet(30.48) = 100.0
			},
		},
		{
			name:        "position measurement without altitude",
			measurement: "position",
			rawData: map[string]any{
				"value": map[string]any{
					"latitude":  37.7749,
					"longitude": -122.4194,
				},
			},
			source: "test-source",
			expected: &Navigation{
				Lat: 37.7749,
				Lon: -122.4194,
				Alt: 0.0,
			},
		},
		{
			name:        "headingTrue measurement",
			measurement: "headingTrue",
			rawData: map[string]any{
				"value": 3.14159, // π radians = 180 degrees
			},
			source: "test-source",
			expected: &Navigation{
				HeadingTrue: 180.0, // RadiansToDegrees(3.14159) = 180.0
			},
		},
		{
			name:        "magneticVariation measurement",
			measurement: "magneticVariation",
			rawData: map[string]any{
				"value": 0.261799, // 0.261799 radians = 15 degrees
			},
			source: "test-source",
			expected: &Navigation{
				MagVariation: 15.0, // RadiansToDegrees(0.261799) = 15.0
			},
		},
		{
			name:        "magneticDeviation measurement",
			measurement: "magneticDeviation",
			rawData: map[string]any{
				"value": 0.0349066, // 0.0349066 radians = 2 degrees
			},
			source: "test-source",
			expected: &Navigation{
				MagDeviation: 2.0, // RadiansToDegrees(0.0349066) = 2.0
			},
		},
		{
			name:        "datetime measurement (ignored)",
			measurement: "datetime",
			rawData: map[string]any{
				"value": "2025-01-01T12:00:00Z",
			},
			source:   "test-source",
			expected: &Navigation{},
		},
		{
			name:        "courseOverGroundTrue measurement",
			measurement: "courseOverGroundTrue",
			rawData: map[string]any{
				"value": 3.14159, // π radians = 180 degrees
			},
			source: "test-source",
			expected: &Navigation{
				COGTrue: 180.0, // RadiansToDegrees(3.14159) = 180.0
			},
		},
		{
			name:        "attitude measurement",
			measurement: "attitude",
			rawData: map[string]any{
				"value": map[string]any{
					"yaw":   0.0872665, // 0.0872665 radians = 5 degrees
					"pitch": 0.0261799, // 0.0261799 radians = 1.5 degrees
					"roll":  0.0523599, // 0.0523599 radians = 3 degrees
				},
			},
			source: "test-source",
			expected: &Navigation{
				Yaw:   5.0, // RadiansToDegrees(0.0872665) = 5.0
				Pitch: 1.5, // RadiansToDegrees(0.0261799) = 1.5
				Roll:  3.0, // RadiansToDegrees(0.0523599) = 3.0
			},
		},
		{
			name:        "attitude measurement without yaw",
			measurement: "attitude",
			rawData: map[string]any{
				"value": map[string]any{
					"pitch": 0.0261799, // 0.0261799 radians = 1.5 degrees
					"roll":  0.0523599, // 0.0523599 radians = 3 degrees
				},
			},
			source: "test-source",
			expected: &Navigation{
				Yaw:   0.0,
				Pitch: 1.5, // RadiansToDegrees(0.0261799) = 1.5
				Roll:  3.0, // RadiansToDegrees(0.0523599) = 3.0
			},
		},
		{
			name:        "speedThroughWater measurement",
			measurement: "speedThroughWater",
			rawData: map[string]any{
				"value": 5.0, // 5 m/s = 9.71922 knots
			},
			source: "test-source",
			expected: &Navigation{
				STW: 9.71922, // MetersPerSecondToKnots(5.0) = 9.71922
			},
		},
		{
			name:        "speedThroughWaterReferenceType measurement (ignored)",
			measurement: "speedThroughWaterReferenceType",
			rawData: map[string]any{
				"value": "test",
			},
			source:   "test-source",
			expected: &Navigation{},
		},
		{
			name:        "log measurement (ignored)",
			measurement: "log",
			rawData: map[string]any{
				"value": 42.0,
			},
			source:   "test-source",
			expected: &Navigation{},
		},
		{
			name:        "unknown measurement",
			measurement: "unknown",
			rawData: map[string]any{
				"value": 42.0,
			},
			source:   "test-source",
			expected: &Navigation{},
		},
		{
			name:        "invalid numeric value",
			measurement: "headingMagnetic",
			rawData: map[string]any{
				"value": "not a number",
			},
			source:   "test-source",
			expected: &Navigation{},
		},
		{
			name:        "invalid position value",
			measurement: "position",
			rawData: map[string]any{
				"value": "not a map",
			},
			source:   "test-source",
			expected: &Navigation{},
		},
		{
			name:        "invalid attitude value",
			measurement: "attitude",
			rawData: map[string]any{
				"value": "not a map",
			},
			source:   "test-source",
			expected: &Navigation{},
		},
		{
			name:        "victron source (skipped)",
			measurement: "position",
			rawData: map[string]any{
				"value": map[string]any{
					"latitude":  37.7749,
					"longitude": -122.4194,
				},
			},
			source:   "venus.com.victronenergy.gps.123",
			expected: &Navigation{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nav := &Navigation{
				BaseSensorData: BaseSensorData{
					Source: tt.source,
				},
			}
			processNavigationData(tt.rawData, tt.measurement, nav)

			// Skip further checks if source is from Victron
			if strings.Contains(tt.source, "venus.com.victronenergy.gps.") {
				return
			}

			switch tt.measurement {
			case "headingMagnetic":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.HeadingMag, nav.HeadingMag, 0.001)
				} else {
					assert.Equal(t, tt.expected.HeadingMag, nav.HeadingMag)
				}
			case "rateOfTurn":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.ROT, nav.ROT, 0.001)
				} else {
					assert.Equal(t, tt.expected.ROT, nav.ROT)
				}
			case "speedOverGround":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.SOG, nav.SOG, 0.001)
				} else {
					assert.Equal(t, tt.expected.SOG, nav.SOG)
				}
			case "position":
				if _, ok := tt.rawData["value"].(map[string]any); ok {
					assert.InDelta(t, tt.expected.Lat, nav.Lat, 0.001)
					assert.InDelta(t, tt.expected.Lon, nav.Lon, 0.001)
					assert.InDelta(t, tt.expected.Alt, nav.Alt, 0.001)
				} else {
					assert.Equal(t, tt.expected.Lat, nav.Lat)
					assert.Equal(t, tt.expected.Lon, nav.Lon)
					assert.Equal(t, tt.expected.Alt, nav.Alt)
				}
			case "headingTrue":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.HeadingTrue, nav.HeadingTrue, 0.001)
				} else {
					assert.Equal(t, tt.expected.HeadingTrue, nav.HeadingTrue)
				}
			case "magneticVariation":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.MagVariation, nav.MagVariation, 0.001)
				} else {
					assert.Equal(t, tt.expected.MagVariation, nav.MagVariation)
				}
			case "magneticDeviation":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.MagDeviation, nav.MagDeviation, 0.001)
				} else {
					assert.Equal(t, tt.expected.MagDeviation, nav.MagDeviation)
				}
			case "courseOverGroundTrue":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.COGTrue, nav.COGTrue, 0.001)
				} else {
					assert.Equal(t, tt.expected.COGTrue, nav.COGTrue)
				}
			case "attitude":
				if _, ok := tt.rawData["value"].(map[string]any); ok {
					assert.InDelta(t, tt.expected.Yaw, nav.Yaw, 0.001)
					assert.InDelta(t, tt.expected.Pitch, nav.Pitch, 0.001)
					assert.InDelta(t, tt.expected.Roll, nav.Roll, 0.001)
				} else {
					assert.Equal(t, tt.expected.Yaw, nav.Yaw)
					assert.Equal(t, tt.expected.Pitch, nav.Pitch)
					assert.Equal(t, tt.expected.Roll, nav.Roll)
				}
			case "speedThroughWater":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.STW, nav.STW, 0.001)
				} else {
					assert.Equal(t, tt.expected.STW, nav.STW)
				}
			case "datetime", "speedThroughWaterReferenceType", "log":
				// These are ignored in the implementation
				assert.Equal(t, tt.expected.Lat, nav.Lat)
				assert.Equal(t, tt.expected.Lon, nav.Lon)
				assert.Equal(t, tt.expected.Alt, nav.Alt)
				assert.Equal(t, tt.expected.SOG, nav.SOG)
				assert.Equal(t, tt.expected.ROT, nav.ROT)
				assert.Equal(t, tt.expected.COGTrue, nav.COGTrue)
				assert.Equal(t, tt.expected.HeadingMag, nav.HeadingMag)
				assert.Equal(t, tt.expected.MagVariation, nav.MagVariation)
				assert.Equal(t, tt.expected.MagDeviation, nav.MagDeviation)
				assert.Equal(t, tt.expected.Yaw, nav.Yaw)
				assert.Equal(t, tt.expected.Pitch, nav.Pitch)
				assert.Equal(t, tt.expected.Roll, nav.Roll)
				assert.Equal(t, tt.expected.HeadingTrue, nav.HeadingTrue)
				assert.Equal(t, tt.expected.STW, nav.STW)
			default:
				assert.Equal(t, tt.expected.Lat, nav.Lat)
				assert.Equal(t, tt.expected.Lon, nav.Lon)
				assert.Equal(t, tt.expected.Alt, nav.Alt)
				assert.Equal(t, tt.expected.SOG, nav.SOG)
				assert.Equal(t, tt.expected.ROT, nav.ROT)
				assert.Equal(t, tt.expected.COGTrue, nav.COGTrue)
				assert.Equal(t, tt.expected.HeadingMag, nav.HeadingMag)
				assert.Equal(t, tt.expected.MagVariation, nav.MagVariation)
				assert.Equal(t, tt.expected.MagDeviation, nav.MagDeviation)
				assert.Equal(t, tt.expected.Yaw, nav.Yaw)
				assert.Equal(t, tt.expected.Pitch, nav.Pitch)
				assert.Equal(t, tt.expected.Roll, nav.Roll)
				assert.Equal(t, tt.expected.HeadingTrue, nav.HeadingTrue)
				assert.Equal(t, tt.expected.STW, nav.STW)
			}
		})
	}
}

func TestHandleNavigationMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data for heading magnetic
	headingData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     3.14159, // π radians = 180 degrees
	}
	headingPayload, _ := json.Marshal(headingData)
	headingMessage := NewMockMessage("vessels/test/navigation/headingMagnetic", headingPayload)

	// Call OnNavigationMessage
	OnNavigationMessage(client, headingMessage)

	// Create test data for position
	positionData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value": map[string]any{
			"latitude":  37.7749,
			"longitude": -122.4194,
			"altitude":  30.48, // 30.48 meters = 100 feet
		},
	}
	positionPayload, _ := json.Marshal(positionData)
	positionMessage := NewMockMessage("vessels/test/navigation/position", positionPayload)

	// Call OnNavigationMessage
	OnNavigationMessage(client, positionMessage)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("vessels/test/navigation/headingMagnetic", []byte("invalid json"))
	OnNavigationMessage(client, invalidMessage)

	// Test with Victron source
	victronData := map[string]any{
		"$source":   "venus.com.victronenergy.gps.123",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value": map[string]any{
			"latitude":  37.7749,
			"longitude": -122.4194,
		},
	}
	victronPayload, _ := json.Marshal(victronData)
	victronMessage := NewMockMessage("vessels/test/navigation/position", victronPayload)

	// Call OnNavigationMessage
	OnNavigationMessage(client, victronMessage)
}
