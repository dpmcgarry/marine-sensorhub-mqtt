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
	"fmt"
	"math"
	"reflect"
	"strconv"
)

func ParseFloat64(a any) (float64, error) {
	var floatType = reflect.TypeOf(float64(0))
	var stringType = reflect.TypeOf("")
	switch i := a.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		v := reflect.ValueOf(a)
		v = reflect.Indirect(v)
		if !v.IsValid() {
			return math.NaN(), fmt.Errorf("can't convert %v to float64", v.Type())
		} else if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			return strconv.ParseFloat(s, 64)
		} else {
			return math.NaN(), fmt.Errorf("can't convert %v to float64", v.Type())
		}
	}
}

func ParseString(a any) (string, error) {
	if strVal, ok := a.(string); ok {
		return strVal, nil
	}
	return "", fmt.Errorf("type assertion to string failed for value: %v", a)
}

func ParseMapString(a any) (map[string]any, error) {
	if m, ok := a.(map[string]any); ok {
		return m, nil
	}
	return nil, fmt.Errorf("type assertion to map[string]any failed")
}
