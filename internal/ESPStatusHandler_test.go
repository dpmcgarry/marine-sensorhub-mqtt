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

func TestESPStatusStruct(t *testing.T) {
	// Create an ESPStatus instance
	now := time.Now()
	espStatus := ESPStatus{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		MAC:                "AA:BB:CC:DD:EE:FF",
		Location:           "Engine Room",
		IPAddress:          "192.168.1.100",
		MSHVersion:         "1.2.3",
		FreeSRAM:           50000,
		FreeHeap:           40000,
		FreePSRAM:          30000,
		WiFiReconnectCount: 5,
		MQTTReconnectCount: 3,
		BLEEnabled:         true,
		RTDEnabled:         true,
		WiFiRSSI:           -65,
		HasTime:            true,
		HasResetMQTT:       false,
	}

	// Test ToJSON
	jsonData := espStatus.ToJSON()
	var parsedESPStatus ESPStatus
	err := json.Unmarshal([]byte(jsonData), &parsedESPStatus)
	assert.NoError(t, err)
	assert.Equal(t, espStatus.MAC, parsedESPStatus.MAC)
	assert.Equal(t, espStatus.Location, parsedESPStatus.Location)
	assert.Equal(t, espStatus.IPAddress, parsedESPStatus.IPAddress)
	assert.Equal(t, espStatus.MSHVersion, parsedESPStatus.MSHVersion)
	assert.Equal(t, espStatus.FreeSRAM, parsedESPStatus.FreeSRAM)
	assert.Equal(t, espStatus.FreeHeap, parsedESPStatus.FreeHeap)
	assert.Equal(t, espStatus.FreePSRAM, parsedESPStatus.FreePSRAM)
	assert.Equal(t, espStatus.WiFiReconnectCount, parsedESPStatus.WiFiReconnectCount)
	assert.Equal(t, espStatus.MQTTReconnectCount, parsedESPStatus.MQTTReconnectCount)
	assert.Equal(t, espStatus.BLEEnabled, parsedESPStatus.BLEEnabled)
	assert.Equal(t, espStatus.RTDEnabled, parsedESPStatus.RTDEnabled)
	assert.Equal(t, espStatus.WiFiRSSI, parsedESPStatus.WiFiRSSI)
	assert.Equal(t, espStatus.HasTime, parsedESPStatus.HasTime)
	assert.Equal(t, espStatus.HasResetMQTT, parsedESPStatus.HasResetMQTT)

	// Test IsEmpty - ESP status data is always considered valid
	assert.False(t, espStatus.IsEmpty())

	emptyESPStatus := ESPStatus{}
	assert.False(t, emptyESPStatus.IsEmpty())

	// Test GetInfluxTags
	tags := espStatus.GetInfluxTags()
	assert.Equal(t, espStatus.MAC, tags["MAC"])
	assert.Equal(t, espStatus.Location, tags["Location"])
	assert.Equal(t, espStatus.IPAddress, tags["IPAddress"])
	assert.Equal(t, espStatus.MSHVersion, tags["MSHVersion"])

	// Test GetInfluxTags with empty fields
	espStatusNoLoc := ESPStatus{
		MAC: "AA:BB:CC:DD:EE:FF",
	}
	tagsNoLoc := espStatusNoLoc.GetInfluxTags()
	assert.Equal(t, espStatusNoLoc.MAC, tagsNoLoc["MAC"])
	_, hasLocation := tagsNoLoc["Location"]
	assert.False(t, hasLocation)
	_, hasIPAddress := tagsNoLoc["IPAddress"]
	assert.False(t, hasIPAddress)
	_, hasMSHVersion := tagsNoLoc["MSHVersion"]
	assert.False(t, hasMSHVersion)

	// Test GetInfluxFields
	fields := espStatus.GetInfluxFields()
	assert.Equal(t, espStatus.FreeSRAM, fields["FreeSRAM"])
	assert.Equal(t, espStatus.FreeHeap, fields["FreeHeap"])
	assert.Equal(t, espStatus.FreePSRAM, fields["FreePSRAM"])
	assert.Equal(t, espStatus.WiFiReconnectCount, fields["WiFiReconnectCount"])
	assert.Equal(t, espStatus.MQTTReconnectCount, fields["MQTTReconnectCount"])
	assert.Equal(t, 1, fields["BLEEnabled"])
	assert.Equal(t, 1, fields["RTDEnabled"])
	assert.Equal(t, espStatus.WiFiRSSI, fields["WiFiRSSI"])
	assert.Equal(t, 1, fields["HasTime"])
	assert.Equal(t, 0, fields["HasResetMQTT"])

	// Test GetInfluxFields with different boolean values
	espStatusBooleans := ESPStatus{
		BLEEnabled:   false,
		RTDEnabled:   false,
		HasTime:      false,
		HasResetMQTT: true,
	}
	boolFields := espStatusBooleans.GetInfluxFields()
	assert.Equal(t, 0, boolFields["BLEEnabled"])
	assert.Equal(t, 0, boolFields["RTDEnabled"])
	assert.Equal(t, 0, boolFields["HasTime"])
	assert.Equal(t, 1, boolFields["HasResetMQTT"])

	// Test GetInfluxFields with zero values
	espStatusZeros := ESPStatus{
		FreeSRAM:  0,
		FreeHeap:  0,
		FreePSRAM: 0,
		WiFiRSSI:  0,
	}
	zeroFields := espStatusZeros.GetInfluxFields()
	_, hasFreeSRAM := zeroFields["FreeSRAM"]
	assert.False(t, hasFreeSRAM)
	_, hasFreeHeap := zeroFields["FreeHeap"]
	assert.False(t, hasFreeHeap)
	_, hasFreePSRAM := zeroFields["FreePSRAM"]
	assert.False(t, hasFreePSRAM)
	_, hasWiFiRSSI := zeroFields["WiFiRSSI"]
	assert.False(t, hasWiFiRSSI)
	// These should always be included even if zero
	assert.Equal(t, int64(0), zeroFields["WiFiReconnectCount"])
	assert.Equal(t, int64(0), zeroFields["MQTTReconnectCount"])

	// Test GetSource and SetSource
	assert.Equal(t, espStatus.MAC, espStatus.GetSource())

	espStatusSetSource := ESPStatus{}
	espStatusSetSource.SetSource("BB:CC:DD:EE:FF:AA")
	assert.Equal(t, "BB:CC:DD:EE:FF:AA", espStatusSetSource.MAC)
	assert.Equal(t, "BB:CC:DD:EE:FF:AA", espStatusSetSource.GetSource())

	// Test GetMeasurementName
	assert.Equal(t, "espStatus", espStatus.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "esp/status", espStatus.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.ESPLogEn, espStatus.GetLogEnabled())

	// Test ToInfluxPoint
	point := espStatus.ToInfluxPoint()
	assert.NotNil(t, point)
}

func TestHandleESPStatusMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data
	espData := map[string]any{
		"MAC":                "AA:BB:CC:DD:EE:FF",
		"IPAddress":          "192.168.1.100",
		"MSHVersion":         "1.2.3",
		"FreeSRAM":           50000,
		"FreeHeap":           40000,
		"FreePSRAM":          30000,
		"WiFiReconnectCount": 5,
		"MQTTReconnectCount": 3,
		"BLEEnabled":         true,
		"RTDEnabled":         true,
		"WiFiRSSI":           -65,
		"HasTime":            true,
		"HasResetMQTT":       false,
		"timestamp":          "2025-01-01T12:00:00.000Z",
	}
	espPayload, _ := json.Marshal(espData)
	espMessage := NewMockMessage("esp/status", espPayload)

	// Call OnESPStatusMessage
	OnESPStatusMessage(client, espMessage)

	// Test with MAC that has a location mapping
	// First add a mapping to the SharedSubscriptionConfig
	SharedSubscriptionConfig.MACtoLocation = map[string]string{
		"BB:CC:DD:EE:FF:AA": "Engine Room",
	}

	espDataWithMapping := map[string]any{
		"MAC":                "BB:CC:DD:EE:FF:AA",
		"IPAddress":          "192.168.1.101",
		"MSHVersion":         "1.2.4",
		"FreeSRAM":           55000,
		"FreeHeap":           45000,
		"FreePSRAM":          35000,
		"WiFiReconnectCount": 2,
		"MQTTReconnectCount": 1,
		"BLEEnabled":         false,
		"RTDEnabled":         false,
		"WiFiRSSI":           -70,
		"HasTime":            false,
		"HasResetMQTT":       true,
		"timestamp":          "2025-01-01T12:00:00.000Z",
	}
	espPayloadWithMapping, _ := json.Marshal(espDataWithMapping)
	espMessageWithMapping := NewMockMessage("esp/status", espPayloadWithMapping)

	// Call OnESPStatusMessage
	OnESPStatusMessage(client, espMessageWithMapping)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("esp/status", []byte("invalid json"))
	OnESPStatusMessage(client, invalidMessage)

	// Test with empty MAC
	espDataEmptyMAC := map[string]any{
		"MAC":        "",
		"IPAddress":  "192.168.1.100",
		"MSHVersion": "1.2.3",
		"timestamp":  "2025-01-01T12:00:00.000Z",
	}
	espPayloadEmptyMAC, _ := json.Marshal(espDataEmptyMAC)
	espMessageEmptyMAC := NewMockMessage("esp/status", espPayloadEmptyMAC)

	// Call OnESPStatusMessage
	OnESPStatusMessage(client, espMessageEmptyMAC)

	// Test with unknown MAC
	espDataUnknownMAC := map[string]any{
		"MAC":        "unknown-mac",
		"IPAddress":  "192.168.1.100",
		"MSHVersion": "1.2.3",
		"timestamp":  "2025-01-01T12:00:00.000Z",
	}
	espPayloadUnknownMAC, _ := json.Marshal(espDataUnknownMAC)
	espMessageUnknownMAC := NewMockMessage("esp/status", espPayloadUnknownMAC)

	// Call OnESPStatusMessage
	OnESPStatusMessage(client, espMessageUnknownMAC)
}
