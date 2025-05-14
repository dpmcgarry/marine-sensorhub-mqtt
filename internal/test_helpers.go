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
	"context"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// Mock implementations for testing

// MockMQTTClient is a mock implementation of the MQTT.Client interface
type MockMQTTClient struct{}

func (m *MockMQTTClient) Connect() MQTT.Token {
	return &MockToken{}
}

func (m *MockMQTTClient) Disconnect(quiesce uint) {}

func (m *MockMQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	return &MockToken{}
}

func (m *MockMQTTClient) Subscribe(topic string, qos byte, callback MQTT.MessageHandler) MQTT.Token {
	return &MockToken{}
}

func (m *MockMQTTClient) SubscribeMultiple(filters map[string]byte, callback MQTT.MessageHandler) MQTT.Token {
	return &MockToken{}
}

func (m *MockMQTTClient) Unsubscribe(topics ...string) MQTT.Token {
	return &MockToken{}
}

func (m *MockMQTTClient) AddRoute(topic string, callback MQTT.MessageHandler) {}

func (m *MockMQTTClient) OptionsReader() MQTT.ClientOptionsReader {
	return MQTT.ClientOptionsReader{}
}

func (m *MockMQTTClient) IsConnected() bool {
	return true
}

func (m *MockMQTTClient) IsConnectionOpen() bool {
	return true
}

// MockToken is a mock implementation of the MQTT.Token interface
type MockToken struct{}

func (m *MockToken) Wait() bool {
	return true
}

func (m *MockToken) WaitTimeout(time.Duration) bool {
	return true
}

func (m *MockToken) Error() error {
	return nil
}

func (m *MockToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch) // Return a closed channel to indicate it's already done
	return ch
}

// MockMessage is a mock implementation of the MQTT.Message interface
type MockMessage struct {
	topic   string
	payload []byte
}

func NewMockMessage(topic string, payload []byte) *MockMessage {
	return &MockMessage{
		topic:   topic,
		payload: payload,
	}
}

func (m *MockMessage) Duplicate() bool {
	return false
}

func (m *MockMessage) Qos() byte {
	return 0
}

func (m *MockMessage) Retained() bool {
	return false
}

func (m *MockMessage) Topic() string {
	return m.topic
}

func (m *MockMessage) MessageID() uint16 {
	return 0
}

func (m *MockMessage) Payload() []byte {
	return m.payload
}

func (m *MockMessage) Ack() {}

// MockInfluxWriteAPI is a mock implementation of the influxdb2 WriteAPIBlocking
type MockInfluxWriteAPI struct {
	Points []*write.Point
}

func NewMockInfluxWriteAPI() *MockInfluxWriteAPI {
	return &MockInfluxWriteAPI{
		Points: make([]*write.Point, 0),
	}
}

func (m *MockInfluxWriteAPI) WritePoint(ctx context.Context, points ...*write.Point) error {
	m.Points = append(m.Points, points...)
	return nil
}

// Implement the rest of the WriteAPIBlocking interface
func (m *MockInfluxWriteAPI) WriteRecord(ctx context.Context, records ...string) error {
	return nil
}

func (m *MockInfluxWriteAPI) WriteRecords(ctx context.Context, records []string) error {
	return nil
}

func (m *MockInfluxWriteAPI) EnableBatching() {}

func (m *MockInfluxWriteAPI) Flush(ctx context.Context) error {
	return nil
}

// TestConfig creates a test configuration for testing
func TestConfig() *SubscriptionConfig {
	return &SubscriptionConfig{
		Repost:          true,
		RepostRootTopic: "test/",
		InfluxEnabled:   true,
		N2KtoName:       map[string]string{"test-source": "mapped-source"},
		MACtoLocation:   map[string]string{"test-mac": "test-location"},
		WaterLogEn:      true,
		NavLogEn:        true,
		WindLogEn:       true,
		GNSSLogEn:       true,
		OutsideLogEn:    true,
		BLELogEn:        true,
		SteerLogEn:      true,
		PropLogEn:       true,
		ESPLogEn:        true,
		PHYLogEn:        true,
		PublishTimeout:  1000,
	}
}

// SetupTestEnvironment sets up the test environment with mock implementations
// Returns a cleanup function to restore the original values
func SetupTestEnvironment() func() {
	// Save original values
	originalConfig := SharedSubscriptionConfig
	originalInfluxWriteAPI := SharedInfluxWriteAPI

	// Create mock implementations
	mockConfig := TestConfig()
	mockWriteAPI := NewMockInfluxWriteAPI()

	// Set mock values
	SharedSubscriptionConfig = mockConfig
	SharedInfluxWriteAPI = mockWriteAPI

	// Return a function to restore original values
	return func() {
		SharedSubscriptionConfig = originalConfig
		SharedInfluxWriteAPI = originalInfluxWriteAPI
	}
}
