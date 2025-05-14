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

func TestRadiansToDegrees(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"pi", math.Pi, 180},
		{"2pi", 2 * math.Pi, 360},
		{"pi/2", math.Pi / 2, 90},
		{"pi/4", math.Pi / 4, 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RadiansToDegrees(tt.input)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestMetersPerSecondToKnots(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"one", 1, 1.943844},
		{"ten", 10, 19.43844},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MetersPerSecondToKnots(tt.input)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestMetersToFeet(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"one", 1, 3.28084},
		{"ten", 10, 32.8084},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MetersToFeet(tt.input)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestKelvinToFarenheit(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"absolute zero", 0, -459.67},
		{"freezing", 273.15, 32},
		{"boiling", 373.15, 212},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KelvinToFarenheit(tt.input)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestKelvinToCelsius(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"absolute zero", 0, -273.15},
		{"freezing", 273.15, 0},
		{"boiling", 373.15, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KelvinToCelsius(tt.input)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestMillibarToInHg(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"standard pressure", 1013.25, 29.92},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MillibarToInHg(tt.input)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestPascalToPSI(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"one atmosphere", 101325, 14.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PascalToPSI(tt.input)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestCubicMetersPerSecondToGallonsPerHour(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"one", 1, 951019.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CubicMetersPerSecondToGallonsPerHour(tt.input)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestCubicMetersPerSecondToGallonsPerMinute(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"one", 1, 15850.323},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CubicMetersPerSecondToGallonsPerMinute(tt.input)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestCubicMetersPerSecondToGallonsPerSecond(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"one", 1, 264.172056},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CubicMetersPerSecondToGallonsPerSecond(tt.input)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}
