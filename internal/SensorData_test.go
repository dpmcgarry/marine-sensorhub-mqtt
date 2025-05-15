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
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Initialize the test environment
	SharedSubscriptionConfig = TestConfig()
	SharedInfluxWriteAPI = NewMockInfluxWriteAPI()
}

// TestBaseSensorData tests the BaseSensorData struct methods
func TestBaseSensorData(t *testing.T) {
	// Create a BaseSensorData instance
	now := time.Now()
	base := BaseSensorData{
		Source:    "test-source",
		Timestamp: now,
	}

	// Test GetSource
	assert.Equal(t, "test-source", base.GetSource(), "GetSource should return the source")

	// Test GetTimestamp
	assert.Equal(t, now, base.GetTimestamp(), "GetTimestamp should return the timestamp")

	// Test SetSource
	base.SetSource("new-source")
	assert.Equal(t, "new-source", base.Source, "SetSource should update the source")

	// Test SetTimestamp
	newTime := time.Now().Add(time.Hour)
	base.SetTimestamp(newTime)
	assert.Equal(t, newTime, base.Timestamp, "SetTimestamp should update the timestamp")

	// Test GetInfluxTags
	tags := base.GetInfluxTags()
	assert.Equal(t, "new-source", tags["Source"], "GetInfluxTags should include the source")
}

// MockSensorData is a mock implementation of the SensorData interface for testing
type MockSensorData struct {
	BaseSensorData
	MockIsEmpty         bool
	MockLogEnabled      bool
	MockMeasurementName string
	MockTopicPrefix     string
	MockFields          map[string]interface{}
}

func (m *MockSensorData) ToJSON() string {
	return `{"Source":"` + m.Source + `"}`
}

func (m *MockSensorData) LogJSON() {
	// Do nothing for testing
}

func (m *MockSensorData) IsEmpty() bool {
	return m.MockIsEmpty
}

func (m *MockSensorData) GetInfluxFields() map[string]interface{} {
	return m.MockFields
}

func (m *MockSensorData) ToInfluxPoint() *write.Point {
	return nil // Not needed for testing
}

func (m *MockSensorData) GetLogEnabled() bool {
	return m.MockLogEnabled
}

func (m *MockSensorData) GetMeasurementName() string {
	return m.MockMeasurementName
}

func (m *MockSensorData) GetTopicPrefix() string {
	return m.MockTopicPrefix
}

// TestParseCommonFields tests the ParseCommonFields function
func TestParseCommonFields(t *testing.T) {
	// Create test data
	rawData := map[string]any{
		"$source":   "test-source",
		"timestamp": "2025-01-01T12:00:00.000Z",
	}

	// Create a mock sensor data
	mockData := &MockSensorData{
		MockIsEmpty:         false,
		MockLogEnabled:      true,
		MockMeasurementName: "test",
		MockTopicPrefix:     "test",
		MockFields:          make(map[string]interface{}),
	}

	// Set up N2KtoName mapping for testing
	originalMapping := SharedSubscriptionConfig.N2KtoName
	SharedSubscriptionConfig.N2KtoName = map[string]string{
		"test-source": "mapped-source",
	}
	defer func() {
		SharedSubscriptionConfig.N2KtoName = originalMapping
	}()

	// Call ParseCommonFields
	ParseCommonFields(rawData, mockData)

	// Verify source was parsed and mapped
	assert.Equal(t, "mapped-source", mockData.Source, "Source should be parsed and mapped")

	// Verify timestamp was parsed
	expectedTime, _ := time.Parse(ISOTimeLayout, "2025-01-01T12:00:00.000Z")
	assert.Equal(t, expectedTime, mockData.Timestamp, "Timestamp should be parsed correctly")

	// Test with missing fields
	mockData = &MockSensorData{}
	ParseCommonFields(map[string]any{}, mockData)
	assert.NotEqual(t, time.Time{}, mockData.Timestamp, "Default timestamp should be set if not provided")
}
