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

func TestGNSSStruct(t *testing.T) {
	// Create a GNSS instance
	now := time.Now()
	gnss := GNSS{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		AntennaAlt:    123.45,
		Satellites:    12,
		HozDilution:   1.2,
		PosDilution:   2.3,
		GeoidalSep:    34.5,
		Type:          "GPS",
		MethodQuality: "3D",
	}

	// Test ToJSON
	jsonData := gnss.ToJSON()
	var parsedGNSS GNSS
	err := json.Unmarshal([]byte(jsonData), &parsedGNSS)
	assert.NoError(t, err)
	assert.Equal(t, gnss.Source, parsedGNSS.Source)
	assert.Equal(t, gnss.AntennaAlt, parsedGNSS.AntennaAlt)
	assert.Equal(t, gnss.Satellites, parsedGNSS.Satellites)
	assert.Equal(t, gnss.HozDilution, parsedGNSS.HozDilution)
	assert.Equal(t, gnss.PosDilution, parsedGNSS.PosDilution)
	assert.Equal(t, gnss.GeoidalSep, parsedGNSS.GeoidalSep)
	assert.Equal(t, gnss.Type, parsedGNSS.Type)
	assert.Equal(t, gnss.MethodQuality, parsedGNSS.MethodQuality)

	// Test IsEmpty
	assert.False(t, gnss.IsEmpty())

	emptyGNSS := GNSS{}
	assert.True(t, emptyGNSS.IsEmpty())

	// Test GetInfluxFields
	fields := gnss.GetInfluxFields()
	assert.Equal(t, gnss.AntennaAlt, fields["AntennaAlt"])
	assert.Equal(t, gnss.Satellites, fields["Satellites"])
	assert.Equal(t, gnss.HozDilution, fields["HozDilution"])
	assert.Equal(t, gnss.PosDilution, fields["PosDilution"])
	assert.Equal(t, gnss.GeoidalSep, fields["GeoidalSep"])
	assert.Equal(t, gnss.Type, fields["Type"])
	assert.Equal(t, gnss.MethodQuality, fields["MethodQuality"])

	// Test GetMeasurementName
	assert.Equal(t, "gnss", gnss.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "gnss", gnss.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.GNSSLogEn, gnss.GetLogEnabled())
}

func TestProcessGNSSData(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	tests := []struct {
		name        string
		measurement string
		rawData     map[string]any
		expected    *GNSS
	}{
		{
			name:        "antennaAltitude measurement",
			measurement: "antennaAltitude",
			rawData: map[string]any{
				"value": 123.45,
			},
			expected: &GNSS{
				AntennaAlt: 123.45,
			},
		},
		{
			name:        "satellites measurement",
			measurement: "satellites",
			rawData: map[string]any{
				"value": 12.0,
			},
			expected: &GNSS{
				Satellites: 12,
			},
		},
		{
			name:        "horizontalDilution measurement",
			measurement: "horizontalDilution",
			rawData: map[string]any{
				"value": 1.2,
			},
			expected: &GNSS{
				HozDilution: 1.2,
			},
		},
		{
			name:        "positionDilution measurement",
			measurement: "positionDilution",
			rawData: map[string]any{
				"value": 2.3,
			},
			expected: &GNSS{
				PosDilution: 2.3,
			},
		},
		{
			name:        "geoidalSeparation measurement",
			measurement: "geoidalSeparation",
			rawData: map[string]any{
				"value": 34.5,
			},
			expected: &GNSS{
				GeoidalSep: 34.5,
			},
		},
		{
			name:        "type measurement",
			measurement: "type",
			rawData: map[string]any{
				"value": "GPS",
			},
			expected: &GNSS{
				Type: "GPS",
			},
		},
		{
			name:        "methodQuality measurement",
			measurement: "methodQuality",
			rawData: map[string]any{
				"value": "3D",
			},
			expected: &GNSS{
				MethodQuality: "3D",
			},
		},
		{
			name:        "integrity measurement (ignored)",
			measurement: "integrity",
			rawData: map[string]any{
				"value": "test",
			},
			expected: &GNSS{},
		},
		{
			name:        "satellitesInView measurement (ignored)",
			measurement: "satellitesInView",
			rawData: map[string]any{
				"value": 15,
			},
			expected: &GNSS{},
		},
		{
			name:        "unknown measurement",
			measurement: "unknown",
			rawData: map[string]any{
				"value": 42.0,
			},
			expected: &GNSS{},
		},
		{
			name:        "invalid numeric value",
			measurement: "antennaAltitude",
			rawData: map[string]any{
				"value": "not a number",
			},
			expected: &GNSS{},
		},
		{
			name:        "invalid string value",
			measurement: "type",
			rawData: map[string]any{
				"value": 123,
			},
			expected: &GNSS{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gnss := &GNSS{}
			processGNSSData(tt.rawData, tt.measurement, gnss)

			switch tt.measurement {
			case "antennaAltitude":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.AntennaAlt, gnss.AntennaAlt, 0.001)
				} else {
					assert.Equal(t, tt.expected.AntennaAlt, gnss.AntennaAlt)
				}
			case "satellites":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.Equal(t, tt.expected.Satellites, gnss.Satellites)
				} else {
					assert.Equal(t, tt.expected.Satellites, gnss.Satellites)
				}
			case "horizontalDilution":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.HozDilution, gnss.HozDilution, 0.001)
				} else {
					assert.Equal(t, tt.expected.HozDilution, gnss.HozDilution)
				}
			case "positionDilution":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.PosDilution, gnss.PosDilution, 0.001)
				} else {
					assert.Equal(t, tt.expected.PosDilution, gnss.PosDilution)
				}
			case "geoidalSeparation":
				if _, ok := tt.rawData["value"].(float64); ok {
					assert.InDelta(t, tt.expected.GeoidalSep, gnss.GeoidalSep, 0.001)
				} else {
					assert.Equal(t, tt.expected.GeoidalSep, gnss.GeoidalSep)
				}
			case "type":
				assert.Equal(t, tt.expected.Type, gnss.Type)
			case "methodQuality":
				assert.Equal(t, tt.expected.MethodQuality, gnss.MethodQuality)
			case "integrity", "satellitesInView":
				// These are ignored in the implementation
				assert.Equal(t, tt.expected.AntennaAlt, gnss.AntennaAlt)
				assert.Equal(t, tt.expected.Satellites, gnss.Satellites)
				assert.Equal(t, tt.expected.HozDilution, gnss.HozDilution)
				assert.Equal(t, tt.expected.PosDilution, gnss.PosDilution)
				assert.Equal(t, tt.expected.GeoidalSep, gnss.GeoidalSep)
				assert.Equal(t, tt.expected.Type, gnss.Type)
				assert.Equal(t, tt.expected.MethodQuality, gnss.MethodQuality)
			default:
				assert.Equal(t, tt.expected.AntennaAlt, gnss.AntennaAlt)
				assert.Equal(t, tt.expected.Satellites, gnss.Satellites)
				assert.Equal(t, tt.expected.HozDilution, gnss.HozDilution)
				assert.Equal(t, tt.expected.PosDilution, gnss.PosDilution)
				assert.Equal(t, tt.expected.GeoidalSep, gnss.GeoidalSep)
				assert.Equal(t, tt.expected.Type, gnss.Type)
				assert.Equal(t, tt.expected.MethodQuality, gnss.MethodQuality)
			}
		})
	}
}

func TestHandleGNSSMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data for antenna altitude
	altData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     123.45,
	}
	altPayload, _ := json.Marshal(altData)
	altMessage := NewMockMessage("vessels/test/gnss/antennaAltitude", altPayload)

	// Call OnGNSSMessage
	OnGNSSMessage(client, altMessage)

	// Create test data for type
	typeData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
		"value":     "GPS",
	}
	typePayload, _ := json.Marshal(typeData)
	typeMessage := NewMockMessage("vessels/test/gnss/type", typePayload)

	// Call OnGNSSMessage
	OnGNSSMessage(client, typeMessage)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("vessels/test/gnss/antennaAltitude", []byte("invalid json"))
	OnGNSSMessage(client, invalidMessage)
}
