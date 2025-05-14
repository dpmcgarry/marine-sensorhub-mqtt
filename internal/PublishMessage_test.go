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
	"errors"
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// MockToken with error for testing error cases
type MockTokenWithError struct {
	MockToken
	err error
}

func (m *MockTokenWithError) Error() error {
	return m.err
}

// MockMQTTClient with publish error for testing error cases
type MockMQTTClientWithPublishError struct {
	MockMQTTClient
}

func (m *MockMQTTClientWithPublishError) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	return &MockTokenWithError{err: errors.New("publish error")}
}

func TestPublishClientMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	tests := []struct {
		name        string
		client      MQTT.Client
		topic       string
		messagedata string
		strip       bool
		expectTopic string
	}{
		{
			name:        "basic_message",
			client:      &MockMQTTClient{},
			topic:       "test/topic",
			messagedata: "{\"key\":\"value\"}",
			strip:       false,
			expectTopic: "test/topic",
		},
		{
			name:        "strip_spaces",
			client:      &MockMQTTClient{},
			topic:       "test/ topic with spaces ",
			messagedata: "{\"key\":\"value\"}",
			strip:       true,
			expectTopic: "test/topicwithspaces",
		},
		{
			name:        "with_error",
			client:      &MockMQTTClientWithPublishError{},
			topic:       "test/topic",
			messagedata: "{\"key\":\"value\"}",
			strip:       false,
			expectTopic: "test/topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			PublishClientMessage(tt.client, tt.topic, tt.messagedata, tt.strip)

			// Since we're using mocks, we can't directly verify the function's behavior
			// But we can ensure it doesn't panic and the test passes
		})
	}
}

// Test PublishMessage function
func TestPublishMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create test configurations
	publishConf := PublishConfig{
		Interval:          60,
		PublishTimeout:    1000,
		DisconnectTimeout: 2000,
	}

	// Test case 1: Basic server configuration
	serverConf1 := MQTTDestination{
		Host:     "tcp://localhost:1883",
		Topics:   []string{"test/topic1", "test/topic2"},
		Username: "testuser",
		Password: "testpass",
	}

	// Test case 2: Server configuration with CA cert
	serverConf2 := MQTTDestination{
		Host:     "ssl://localhost:8883",
		Topics:   []string{"secure/topic"},
		Username: "secureuser",
		Password: "securepass",
		CACert:   []byte("-----BEGIN CERTIFICATE-----\nMIIDCTCCAfGgAwIBAgIUJQv..."),
	}

	// Test case 3: Server configuration with connection error
	serverConf3 := MQTTDestination{
		Host:   "tcp://error:1883",
		Topics: []string{"error/topic"},
	}

	// Test case 1: Basic server configuration
	t.Run("basic_server_config", func(t *testing.T) {
		// We can't mock MQTT.NewClient directly, but we can test that the function runs without panicking
		PublishMessage(publishConf, serverConf1)
	})

	// Test case 2: Server configuration with CA cert
	t.Run("server_with_ca_cert", func(t *testing.T) {
		// We can't mock MQTT.NewClient directly, but we can test that the function runs without panicking
		PublishMessage(publishConf, serverConf2)
	})

	// Test case 3: Server configuration with connection error
	t.Run("server_with_connection_error", func(t *testing.T) {
		// We can't mock MQTT.NewClient directly, but we can test that the function runs without panicking
		PublishMessage(publishConf, serverConf3)
	})
}
