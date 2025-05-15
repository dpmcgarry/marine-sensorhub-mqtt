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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected float64
		hasError bool
	}{
		{"float64", float64(42.5), 42.5, false},
		{"float32", float32(42.5), 42.5, false},
		{"int64", int64(42), 42.0, false},
		{"int32", int32(42), 42.0, false},
		{"int", int(42), 42.0, false},
		{"uint64", uint64(42), 42.0, false},
		{"uint32", uint32(42), 42.0, false},
		{"uint", uint(42), 42.0, false},
		{"string", "42.5", 42.5, false},
		{"invalid string", "not a number", 0.0, true},
		{"nil", nil, math.NaN(), true},
		{"bool", true, 0.0, true}, // Not directly convertible to float64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFloat64(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				if tt.name == "nil" {
					assert.True(t, math.IsNaN(result))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
		hasError bool
	}{
		{"string", "test", "test", false},
		{"int", 42, "", true},
		{"nil", nil, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseString(tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseMapString(t *testing.T) {
	validMap := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	tests := []struct {
		name     string
		input    any
		expected map[string]any
		hasError bool
	}{
		{"valid map", validMap, validMap, false},
		{"int", 42, nil, true},
		{"nil", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMapString(tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
