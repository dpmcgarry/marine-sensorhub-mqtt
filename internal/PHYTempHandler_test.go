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

func TestPHYTemperatureStruct(t *testing.T) {
	// Create a PHYTemperature instance
	now := time.Now()
	phyTemp := PHYTemperature{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		MAC:       "AA:BB:CC:DD:EE:FF",
		Location:  "Engine Room",
		Device:    "ESP32",
		Component: "CPU",
		TempF:     85.5,
	}

	// Test ToJSON
	jsonData := phyTemp.ToJSON()
	var parsedPHYTemp PHYTemperature
	err := json.Unmarshal([]byte(jsonData), &parsedPHYTemp)
	assert.NoError(t, err)
	assert.Equal(t, phyTemp.MAC, parsedPHYTemp.MAC)
	assert.Equal(t, phyTemp.Location, parsedPHYTemp.Location)
	assert.Equal(t, phyTemp.Device, parsedPHYTemp.Device)
	assert.Equal(t, phyTemp.Component, parsedPHYTemp.Component)
	assert.Equal(t, phyTemp.TempF, parsedPHYTemp.TempF)

	// Test IsEmpty
	assert.False(t, phyTemp.IsEmpty())

	emptyPHYTemp := PHYTemperature{}
	assert.True(t, emptyPHYTemp.IsEmpty())

	// Test GetInfluxTags
	tags := phyTemp.GetInfluxTags()
	assert.Equal(t, phyTemp.MAC, tags["MAC"])
	assert.Equal(t, phyTemp.Location, tags["Location"])
	assert.Equal(t, phyTemp.Device, tags["Device"])
	assert.Equal(t, phyTemp.Component, tags["Component"])

	// Test GetInfluxTags with empty fields
	phyTempNoLoc := PHYTemperature{
		MAC: "AA:BB:CC:DD:EE:FF",
	}
	tagsNoLoc := phyTempNoLoc.GetInfluxTags()
	assert.Equal(t, phyTempNoLoc.MAC, tagsNoLoc["MAC"])
	_, hasLocation := tagsNoLoc["Location"]
	assert.False(t, hasLocation)
	_, hasDevice := tagsNoLoc["Device"]
	assert.False(t, hasDevice)
	_, hasComponent := tagsNoLoc["Component"]
	assert.False(t, hasComponent)

	// Test GetInfluxFields
	fields := phyTemp.GetInfluxFields()
	assert.Equal(t, phyTemp.TempF, fields["TempF"])

	// Test GetInfluxFields with zero values
	phyTempZeros := PHYTemperature{
		TempF: 0.0,
	}
	zeroFields := phyTempZeros.GetInfluxFields()
	_, hasTempF := zeroFields["TempF"]
	assert.False(t, hasTempF)

	// Test GetSource and SetSource
	assert.Equal(t, phyTemp.MAC, phyTemp.GetSource())

	phyTempSetSource := PHYTemperature{}
	phyTempSetSource.SetSource("BB:CC:DD:EE:FF:AA")
	assert.Equal(t, "BB:CC:DD:EE:FF:AA", phyTempSetSource.MAC)
	assert.Equal(t, "BB:CC:DD:EE:FF:AA", phyTempSetSource.GetSource())

	// Test GetMeasurementName
	assert.Equal(t, "phyTemperature", phyTemp.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "rtd/temperature", phyTemp.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.PHYLogEn, phyTemp.GetLogEnabled())

	// Test ToInfluxPoint
	point := phyTemp.ToInfluxPoint()
	assert.NotNil(t, point)
}

func TestHandlePHYTemperatureMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data
	phyTempData := map[string]any{
		"MAC":       "AA:BB:CC:DD:EE:FF",
		"Device":    "ESP32",
		"Component": "CPU",
		"TempF":     85.5,
		"timestamp": "2025-01-01T12:00:00.000Z",
	}
	phyTempPayload, _ := json.Marshal(phyTempData)
	phyTempMessage := NewMockMessage("rtd/temperature", phyTempPayload)

	// Call OnPHYTemperatureMessage
	OnPHYTemperatureMessage(client, phyTempMessage)

	// Test with MAC that has a location mapping
	// First add a mapping to the SharedSubscriptionConfig
	SharedSubscriptionConfig.MACtoLocation = map[string]string{
		"BB:CC:DD:EE:FF:AA": "Engine Room",
	}

	phyTempDataWithMapping := map[string]any{
		"MAC":       "BB:CC:DD:EE:FF:AA",
		"Device":    "ESP32",
		"Component": "CPU",
		"TempF":     90.5,
		"timestamp": "2025-01-01T12:00:00.000Z",
	}
	phyTempPayloadWithMapping, _ := json.Marshal(phyTempDataWithMapping)
	phyTempMessageWithMapping := NewMockMessage("rtd/temperature", phyTempPayloadWithMapping)

	// Call OnPHYTemperatureMessage
	OnPHYTemperatureMessage(client, phyTempMessageWithMapping)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("rtd/temperature", []byte("invalid json"))
	OnPHYTemperatureMessage(client, invalidMessage)

	// Test with empty MAC
	phyTempDataEmptyMAC := map[string]any{
		"MAC":       "",
		"Device":    "ESP32",
		"Component": "CPU",
		"TempF":     85.5,
		"timestamp": "2025-01-01T12:00:00.000Z",
	}
	phyTempPayloadEmptyMAC, _ := json.Marshal(phyTempDataEmptyMAC)
	phyTempMessageEmptyMAC := NewMockMessage("rtd/temperature", phyTempPayloadEmptyMAC)

	// Call OnPHYTemperatureMessage
	OnPHYTemperatureMessage(client, phyTempMessageEmptyMAC)

	// Test with unknown MAC
	phyTempDataUnknownMAC := map[string]any{
		"MAC":       "unknown-mac",
		"Device":    "ESP32",
		"Component": "CPU",
		"TempF":     85.5,
		"timestamp": "2025-01-01T12:00:00.000Z",
	}
	phyTempPayloadUnknownMAC, _ := json.Marshal(phyTempDataUnknownMAC)
	phyTempMessageUnknownMAC := NewMockMessage("rtd/temperature", phyTempPayloadUnknownMAC)

	// Call OnPHYTemperatureMessage
	OnPHYTemperatureMessage(client, phyTempMessageUnknownMAC)
}
