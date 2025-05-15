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

func TestBLETemperatureStruct(t *testing.T) {
	// Create a BLETemperature instance
	now := time.Now()
	bleTemp := BLETemperature{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		MAC:            "AA:BB:CC:DD:EE:FF",
		Location:       "Living Room",
		TempF:          72.5,
		BatteryPercent: 85.0,
		Humidity:       45.0,
		RSSI:           -65,
	}

	// Test ToJSON
	jsonData := bleTemp.ToJSON()
	var parsedBLETemp BLETemperature
	err := json.Unmarshal([]byte(jsonData), &parsedBLETemp)
	assert.NoError(t, err)
	assert.Equal(t, bleTemp.MAC, parsedBLETemp.MAC)
	assert.Equal(t, bleTemp.Location, parsedBLETemp.Location)
	assert.Equal(t, bleTemp.TempF, parsedBLETemp.TempF)
	assert.Equal(t, bleTemp.BatteryPercent, parsedBLETemp.BatteryPercent)
	assert.Equal(t, bleTemp.Humidity, parsedBLETemp.Humidity)
	assert.Equal(t, bleTemp.RSSI, parsedBLETemp.RSSI)

	// Test IsEmpty - BLE temperature data is always considered valid
	assert.False(t, bleTemp.IsEmpty())

	emptyBLETemp := BLETemperature{}
	assert.False(t, emptyBLETemp.IsEmpty())

	// Test GetInfluxTags
	tags := bleTemp.GetInfluxTags()
	assert.Equal(t, bleTemp.MAC, tags["MAC"])
	assert.Equal(t, bleTemp.Location, tags["Location"])

	// Test GetInfluxTags with empty location
	bleTempNoLoc := BLETemperature{
		MAC: "AA:BB:CC:DD:EE:FF",
	}
	tagsNoLoc := bleTempNoLoc.GetInfluxTags()
	assert.Equal(t, bleTempNoLoc.MAC, tagsNoLoc["MAC"])
	_, hasLocation := tagsNoLoc["Location"]
	assert.False(t, hasLocation)

	// Test GetInfluxFields
	fields := bleTemp.GetInfluxFields()
	assert.Equal(t, bleTemp.TempF, fields["TempF"])
	assert.Equal(t, bleTemp.BatteryPercent, fields["BatteryPercent"])
	assert.Equal(t, bleTemp.Humidity, fields["Humidity"])
	assert.Equal(t, bleTemp.RSSI, fields["RSSI"])

	// Test GetInfluxFields with empty fields
	emptyFields := emptyBLETemp.GetInfluxFields()
	assert.Empty(t, emptyFields)

	// Test GetSource and SetSource
	assert.Equal(t, bleTemp.MAC, bleTemp.GetSource())

	bleTempSetSource := BLETemperature{}
	bleTempSetSource.SetSource("BB:CC:DD:EE:FF:AA")
	assert.Equal(t, "BB:CC:DD:EE:FF:AA", bleTempSetSource.MAC)
	assert.Equal(t, "BB:CC:DD:EE:FF:AA", bleTempSetSource.GetSource())

	// Test GetMeasurementName
	assert.Equal(t, "bleTemperature", bleTemp.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "ble/temperature", bleTemp.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.BLELogEn, bleTemp.GetLogEnabled())

	// Test ToInfluxPoint
	point := bleTemp.ToInfluxPoint()
	assert.NotNil(t, point)
}

func TestHandleBLETemperatureMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data
	bleData := map[string]any{
		"MAC":        "AA:BB:CC:DD:EE:FF",
		"TempF":      72.5,
		"BatteryPct": 85.0,
		"Humidity":   45.0,
		"RSSI":       -65,
		"timestamp":  "2025-01-01T12:00:00.000Z",
	}
	blePayload, _ := json.Marshal(bleData)
	bleMessage := NewMockMessage("ble/temperature", blePayload)

	// Call OnBLETemperatureMessage
	OnBLETemperatureMessage(client, bleMessage)

	// Test with MAC that has a location mapping
	// First add a mapping to the SharedSubscriptionConfig
	SharedSubscriptionConfig.MACtoLocation = map[string]string{
		"BB:CC:DD:EE:FF:AA": "Kitchen",
	}

	bleDataWithMapping := map[string]any{
		"MAC":        "BB:CC:DD:EE:FF:AA",
		"TempF":      68.5,
		"BatteryPct": 90.0,
		"Humidity":   50.0,
		"RSSI":       -70,
		"timestamp":  "2025-01-01T12:00:00.000Z",
	}
	blePayloadWithMapping, _ := json.Marshal(bleDataWithMapping)
	bleMessageWithMapping := NewMockMessage("ble/temperature", blePayloadWithMapping)

	// Call OnBLETemperatureMessage
	OnBLETemperatureMessage(client, bleMessageWithMapping)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("ble/temperature", []byte("invalid json"))
	OnBLETemperatureMessage(client, invalidMessage)

	// Test with empty MAC
	bleDataEmptyMAC := map[string]any{
		"MAC":        "",
		"TempF":      72.5,
		"BatteryPct": 85.0,
		"timestamp":  "2025-01-01T12:00:00.000Z",
	}
	blePayloadEmptyMAC, _ := json.Marshal(bleDataEmptyMAC)
	bleMessageEmptyMAC := NewMockMessage("ble/temperature", blePayloadEmptyMAC)

	// Call OnBLETemperatureMessage
	OnBLETemperatureMessage(client, bleMessageEmptyMAC)

	// Test with unknown MAC
	bleDataUnknownMAC := map[string]any{
		"MAC":        "unknown-mac",
		"TempF":      72.5,
		"BatteryPct": 85.0,
		"timestamp":  "2025-01-01T12:00:00.000Z",
	}
	blePayloadUnknownMAC, _ := json.Marshal(bleDataUnknownMAC)
	bleMessageUnknownMAC := NewMockMessage("ble/temperature", blePayloadUnknownMAC)

	// Call OnBLETemperatureMessage
	OnBLETemperatureMessage(client, bleMessageUnknownMAC)
}
