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

import "math"

func RadiansToDegrees(rad float64) float64 {
	return rad * (180 / math.Pi)
}

func MetersPerSecondToKnots(mps float64) float64 {
	return mps * 1.943844
}

func MetersToFeet(m float64) float64 {
	return m * 3.28084
}

func KelvinToFarenheit(tempk float64) float64 {
	return (tempk-273.15)*1.8 + 32
}

func MillibarToInHg(mb float64) float64 {
	return mb / 33.8639
}

func PascalToPSI(pascal float64) float64 {
	return pascal * 0.000145038
}

func CubicMetersPerSecondToGallonsPerHour(cumps float64) float64 {
	return cumps * 951019.4
}

func CubicMetersPerSecondToGallonsPerMinute(cumps float64) float64 {
	return cumps * 15850.323
}

func CubicMetersPerSecondToGallonsPerSecond(cumps float64) float64 {
	return cumps * 264.172056
}
