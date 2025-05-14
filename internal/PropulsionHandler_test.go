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

func TestPropulsionStruct(t *testing.T) {
	// Create a Propulsion instance
	now := time.Now()
	prop := Propulsion{
		BaseSensorData: BaseSensorData{
			Source:    "test-source",
			Timestamp: now,
		},
		Device:           "Engine1",
		RPM:              2500,
		BoostPSI:         15.5,
		OilTempF:         180.5,
		OilPressure:      45.2,
		CoolantTempF:     195.3,
		RunTime:          3600,
		EngineLoad:       75.0,
		EngineTorque:     80.0,
		TransOilTempF:    160.2,
		TransOilPressure: 30.5,
		AltVoltage:       14.2,
		FuelRate:         3.5,
	}

	// Test ToJSON
	jsonData := prop.ToJSON()
	var parsedProp Propulsion
	err := json.Unmarshal([]byte(jsonData), &parsedProp)
	assert.NoError(t, err)
	assert.Equal(t, prop.Device, parsedProp.Device)
	assert.Equal(t, prop.RPM, parsedProp.RPM)
	assert.Equal(t, prop.BoostPSI, parsedProp.BoostPSI)
	assert.Equal(t, prop.OilTempF, parsedProp.OilTempF)
	assert.Equal(t, prop.OilPressure, parsedProp.OilPressure)
	assert.Equal(t, prop.CoolantTempF, parsedProp.CoolantTempF)
	assert.Equal(t, prop.RunTime, parsedProp.RunTime)
	assert.Equal(t, prop.EngineLoad, parsedProp.EngineLoad)
	assert.Equal(t, prop.EngineTorque, parsedProp.EngineTorque)
	assert.Equal(t, prop.TransOilTempF, parsedProp.TransOilTempF)
	assert.Equal(t, prop.TransOilPressure, parsedProp.TransOilPressure)
	assert.Equal(t, prop.AltVoltage, parsedProp.AltVoltage)
	assert.Equal(t, prop.FuelRate, parsedProp.FuelRate)

	// Test IsEmpty
	assert.False(t, prop.IsEmpty())

	emptyProp := Propulsion{}
	assert.True(t, emptyProp.IsEmpty())

	// Test GetInfluxTags
	tags := prop.GetInfluxTags()
	assert.Equal(t, prop.Source, tags["Source"])
	assert.Equal(t, prop.Device, tags["Device"])

	// Test GetInfluxTags with empty fields
	propNoDevice := Propulsion{
		BaseSensorData: BaseSensorData{
			Source: "test-source",
		},
	}
	tagsNoDevice := propNoDevice.GetInfluxTags()
	assert.Equal(t, propNoDevice.Source, tagsNoDevice["Source"])
	_, hasDevice := tagsNoDevice["Device"]
	assert.False(t, hasDevice)

	// Test GetInfluxFields
	fields := prop.GetInfluxFields()
	assert.Equal(t, prop.RPM, fields["RPM"])
	assert.Equal(t, prop.BoostPSI, fields["BoostPSI"])
	assert.Equal(t, prop.OilTempF, fields["OilTempF"])
	assert.Equal(t, prop.OilPressure, fields["OilPressure"])
	assert.Equal(t, prop.CoolantTempF, fields["CoolantTempF"])
	assert.Equal(t, prop.RunTime, fields["RunTime"])
	assert.Equal(t, prop.EngineLoad, fields["EngineLoad"])
	assert.Equal(t, prop.EngineTorque, fields["EngineTorque"])
	assert.Equal(t, prop.TransOilTempF, fields["TransOilTempF"])
	assert.Equal(t, prop.TransOilPressure, fields["TransOilPressure"])
	assert.Equal(t, prop.AltVoltage, fields["AlternatorVoltage"])
	assert.Equal(t, prop.FuelRate, fields["FuelRate"])

	// Test GetInfluxFields with zero values
	propZeros := Propulsion{}
	zeroFields := propZeros.GetInfluxFields()
	_, hasRPM := zeroFields["RPM"]
	assert.False(t, hasRPM)
	_, hasBoostPSI := zeroFields["BoostPSI"]
	assert.False(t, hasBoostPSI)
	// ... and so on for other fields

	// Test GetMeasurementName
	assert.Equal(t, "propulsion", prop.GetMeasurementName())

	// Test GetTopicPrefix
	assert.Equal(t, "propulsion", prop.GetTopicPrefix())

	// Test GetLogEnabled
	cleanup := SetupTestEnvironment()
	defer cleanup()
	assert.Equal(t, SharedSubscriptionConfig.PropLogEn, prop.GetLogEnabled())

	// Test ToInfluxPoint
	point := prop.ToInfluxPoint()
	assert.NotNil(t, point)
}

func TestProcessPropulsionData(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	tests := []struct {
		name        string
		measurement string
		rawData     map[string]any
		context     map[string]interface{}
		checkFunc   func(*testing.T, *Propulsion)
	}{
		{
			name:        "revolutions_measurement",
			measurement: "revolutions",
			rawData:     map[string]any{"value": float64(30)},
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.Equal(t, int64(1800), prop.RPM) // 30 * 60
			},
		},
		{
			name:        "boostPressure_measurement",
			measurement: "boostPressure",
			rawData:     map[string]any{"value": float64(100000)}, // 100 kPa
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.InDelta(t, 14.5038, prop.BoostPSI, 0.001) // Converted to PSI
			},
		},
		{
			name:        "oilTemperature_measurement_engine",
			measurement: "oilTemperature",
			rawData:     map[string]any{"value": float64(350)}, // 350 K
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.InDelta(t, 170.33, prop.OilTempF, 0.01) // Converted to F
				assert.Equal(t, float64(0), prop.TransOilTempF)
			},
		},
		{
			name:        "oilTemperature_measurement_transmission",
			measurement: "oilTemperature",
			rawData:     map[string]any{"value": float64(350)}, // 350 K
			context:     map[string]interface{}{"isTranny": true},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.InDelta(t, 170.33, prop.TransOilTempF, 0.01) // Converted to F
				assert.Equal(t, float64(0), prop.OilTempF)
			},
		},
		{
			name:        "oilPressure_measurement_engine",
			measurement: "oilPressure",
			rawData:     map[string]any{"value": float64(300000)}, // 300 kPa
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.InDelta(t, 43.511, prop.OilPressure, 0.001) // Converted to PSI
				assert.Equal(t, float64(0), prop.TransOilPressure)
			},
		},
		{
			name:        "oilPressure_measurement_transmission",
			measurement: "oilPressure",
			rawData:     map[string]any{"value": float64(300000)}, // 300 kPa
			context:     map[string]interface{}{"isTranny": true},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.InDelta(t, 43.511, prop.TransOilPressure, 0.001) // Converted to PSI
				assert.Equal(t, float64(0), prop.OilPressure)
			},
		},
		{
			name:        "temperature_measurement",
			measurement: "temperature",
			rawData:     map[string]any{"value": float64(360)}, // 360 K
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.InDelta(t, 188.33, prop.CoolantTempF, 0.01) // Converted to F
			},
		},
		{
			name:        "alternatorVoltage_measurement",
			measurement: "alternatorVoltage",
			rawData:     map[string]any{"value": float64(14.2)},
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.Equal(t, 14.2, prop.AltVoltage)
			},
		},
		{
			name:        "transmission_container_topic",
			measurement: "transmission",
			rawData:     map[string]any{"value": "some-value"},
			context:     map[string]interface{}{"isTranny": true},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				// No fields should be set for container topics
				assert.True(t, prop.IsEmpty())
			},
		},
		{
			name:        "fuel_container_topic",
			measurement: "fuel",
			rawData:     map[string]any{"value": "some-value"},
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				// No fields should be set for container topics
				assert.True(t, prop.IsEmpty())
			},
		},
		{
			name:        "fuel_rate_measurement",
			measurement: "rate",
			rawData:     map[string]any{"value": float64(0.0001)}, // 0.0001 m³/s
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.InDelta(t, 95.1019, prop.FuelRate, 0.001) // Converted to gal/hr
			},
		},
		{
			name:        "runTime_measurement",
			measurement: "runTime",
			rawData:     map[string]any{"value": float64(3600)},
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.Equal(t, int64(3600), prop.RunTime)
			},
		},
		{
			name:        "engineLoad_measurement",
			measurement: "engineLoad",
			rawData:     map[string]any{"value": float64(0.75)},
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.Equal(t, 75.0, prop.EngineLoad) // Converted to percentage
			},
		},
		{
			name:        "engineTorque_measurement",
			measurement: "engineTorque",
			rawData:     map[string]any{"value": float64(0.8)},
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.Equal(t, 80.0, prop.EngineTorque) // Converted to percentage
			},
		},
		{
			name:        "unknown_measurement",
			measurement: "unknown",
			rawData:     map[string]any{"value": float64(123)},
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				// No fields should be set for unknown measurements
				assert.True(t, prop.IsEmpty())
			},
		},
		{
			name:        "invalid_revolutions_value",
			measurement: "revolutions",
			rawData:     map[string]any{"value": "not a number"},
			context:     map[string]interface{}{"isTranny": false},
			checkFunc: func(t *testing.T, prop *Propulsion) {
				assert.Equal(t, int64(0), prop.RPM)
			},
		},
		{
			name:        "invalid_context",
			measurement: "oilTemperature",
			rawData:     map[string]any{"value": float64(350)},
			context:     map[string]interface{}{}, // Missing isTranny key
			checkFunc: func(t *testing.T, prop *Propulsion) {
				// Should default to engine oil temp
				assert.InDelta(t, 170.33, prop.OilTempF, 0.01)
				assert.Equal(t, float64(0), prop.TransOilTempF)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := &Propulsion{}
			processPropulsionData(tt.rawData, tt.measurement, prop, tt.context)
			tt.checkFunc(t, prop)
		})
	}
}

func TestHandlePropulsionMessage(t *testing.T) {
	// Set up test environment
	cleanup := SetupTestEnvironment()
	defer cleanup()

	// Create a mock client
	client := &MockMQTTClient{}

	// Test with engine topic
	engineTopic := "vessels/test/propulsion/port/revolutions"
	engineData := map[string]any{
		"value":     float64(30),
		"timestamp": "2025-01-01T12:00:00.000Z",
	}
	enginePayload, _ := json.Marshal(engineData)
	engineMessage := NewMockMessage(engineTopic, enginePayload)

	// Call OnPropulsionMessage
	OnPropulsionMessage(client, engineMessage)

	// Test with transmission topic
	trannyTopic := "vessels/test/propulsion/port/transmission/oilTemperature"
	trannyData := map[string]any{
		"value":     float64(350),
		"timestamp": "2025-01-01T12:00:00.000Z",
	}
	trannyPayload, _ := json.Marshal(trannyData)
	trannyMessage := NewMockMessage(trannyTopic, trannyPayload)

	// Call OnPropulsionMessage
	OnPropulsionMessage(client, trannyMessage)

	// Test with invalid JSON
	invalidMessage := NewMockMessage("vessels/test/propulsion/port/revolutions", []byte("invalid json"))
	OnPropulsionMessage(client, invalidMessage)
}
