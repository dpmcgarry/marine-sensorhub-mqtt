/*
Copyright © 2024 Don P. McGarry

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
	"time"

	"github.com/rs/zerolog/log"
)

type SensorMeasurement struct {
	Name         string
	Tags         map[string]string
	FloatFields  map[string]float64
	IntFields    map[string]int64
	UintFields   map[string]uint64
	StringFields map[string]string
	BoolFields   map[string]bool
	Timestamp    time.Time
}

func (meas SensorMeasurement) Log() {
	log.Debug().Msgf("Name: %v Tags: %v Floats: %v Ints: %v UInts: %v Strings: %v Bools: %v Time: %v", meas.Name, meas.Tags, meas.FloatFields, meas.IntFields, meas.UintFields, meas.StringFields, meas.BoolFields, meas.Timestamp)
}
