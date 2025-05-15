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

// TestHandleJSONMessage tests the HandleJSONMessage function
func TestHandleJSONMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Create test data
	testData := struct {
		BaseSensorData
		Value float64 `json:"value"`
	}{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: parseTime("2025-01-01T12:00:00Z"),
		},
		Value: 42.0,
	}

	// Convert test data to JSON
	payload, _ := json.Marshal(testData)

	// Create a mock message
	message := NewMockMessage("test/topic/measurement", payload)

	// Create a mock sensor data object
	mockData := &MockSensorData{
		MockIsEmpty:         false,
		MockLogEnabled:      true,
		MockMeasurementName: "test",
		MockTopicPrefix:     "test",
		MockFields:          make(map[string]interface{}),
	}

	// Call HandleJSONMessage
	HandleJSONMessage(client, message, mockData)

	// Verify the data was unmarshalled correctly
	assert.Equal(t, "test-source", mockData.Source, "Source should be set correctly")
	assert.Equal(t, parseTime("2025-01-01T12:00:00Z"), mockData.Timestamp, "Timestamp should be set correctly")
}

// TestMapMACToLocation tests the MapMACToLocation function
func TestMapMACToLocation(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock sensor data object
	mockData := &MockSensorData{
		MockIsEmpty:         false,
		MockLogEnabled:      true,
		MockMeasurementName: "test",
		MockTopicPrefix:     "test",
		MockFields:          make(map[string]interface{}),
	}

	// Test with a known MAC address
	location := MapMACToLocation(mockData, "test-mac")
	assert.Equal(t, "test-location", location, "Should return the mapped location for a known MAC")

	// Test with an unknown MAC address
	location = MapMACToLocation(mockData, "unknown-mac")
	assert.Equal(t, "", location, "Should return an empty string for an unknown MAC")
}

// Helper function to parse time strings
func parseTime(timeStr string) time.Time {
	t, _ := time.Parse(ISOTimeLayout, timeStr)
	return t
}
