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
	"encoding/json"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/rs/zerolog/log"
)

type Navigation struct {
	Source       string    `json:"Source,omitempty"`
	Lat          float64   `json:"Latitude,omitempty"`
	Lon          float64   `json:"Longitude,omitempty"`
	Alt          float64   `json:"Altitude,omitempty"`
	SOG          float64   `json:"SpeedOverGround,omitempty"`
	ROT          float64   `json:"RateOfTurn,omitempty"`
	COGTrue      float64   `json:"CourseOverGroundTrue,omitempty"`
	HeadingMag   float64   `json:"HeadingMagnetic,omitempty"`
	MagVariation float64   `json:"MagneticVariation,omitempty"`
	MagDeviation float64   `json:"MagneticDeviation,omitempty"`
	Yaw          float64   `json:"Yaw,omitempty"`
	Pitch        float64   `json:"Pitch,omitempty"`
	Roll         float64   `json:"Roll,omitempty"`
	HeadingTrue  float64   `json:"HeadingTrue,omitempty"`
	STW          float64   `json:"SpeedThroughWater,omitempty"`
	Timestamp    time.Time `json:"Timestamp,omitempty"`
}

func OnNavigationMessage(client MQTT.Client, message MQTT.Message) {
	go handleNavigationMessage(client, message)
}

func handleNavigationMessage(client MQTT.Client, message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.NavLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.NavLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	nav := Navigation{}
	var strtmp string
	strtmp, err = ParseString(rawData["$source"])

	if err != nil {
		log.Warn().Msgf("Error parsing string: %v", err.Error())
	} else {
		nav.Source = strtmp
		if strings.Contains(nav.Source, "venus.com.victronenergy.gps.") {
			return
		}
	}
	strtmp, err = ParseString(rawData["timestamp"])
	if err != nil {
		log.Warn().Msgf("Error parsing string: %v", err.Error())
	} else {
		nav.Timestamp, err = time.Parse(ISOTimeLayout, strtmp)
		if err != nil {
			log.Warn().Msgf("Error parsing time string: %v", err.Error())
		}
	}
	var floatTmp float64
	switch measurement {
	case "headingMagnetic":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			nav.HeadingMag = RadiansToDegrees(floatTmp)
		}

	case "rateOfTurn":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			nav.ROT = RadiansToDegrees(floatTmp)
		}

	case "speedOverGround":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			nav.SOG = MetersPerSecondToKnots(floatTmp)
		}

	case "position":
		postmp, err := ParseMapString(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing map[string]: %v", err.Error())
		} else {
			floatTmp, err = ParseFloat64(postmp["latitude"])
			if err != nil {
				log.Warn().Msgf("Error parsing float64: %v", err.Error())
			} else {
				nav.Lat = floatTmp
			}

			floatTmp, err = ParseFloat64(postmp["longitude"])
			if err != nil {
				log.Warn().Msgf("Error parsing float64: %v", err.Error())
			} else {
				nav.Lon = floatTmp
			}

			floatTmp, err = ParseFloat64(postmp["altitude"])
			if err != nil {
				// Altitude isn't always a thing so just leaving this as a trace
				log.Trace().Msgf("Error parsing float64: %v", err.Error())
			} else {
				nav.Alt = MetersToFeet(floatTmp)
			}
		}
	case "headingTrue":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			nav.HeadingTrue = RadiansToDegrees(floatTmp)
		}
	case "magneticVariation":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			nav.MagVariation = RadiansToDegrees(floatTmp)
		}
	case "magneticDeviation":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			nav.MagDeviation = RadiansToDegrees(floatTmp)
		}
	case "datetime":
		break
	case "courseOverGroundTrue":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			nav.COGTrue = RadiansToDegrees(floatTmp)
		}
	case "attitude":
		atttmp, err := ParseMapString(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing map[string]: %v", err.Error())
		} else {
			floatTmp, err = ParseFloat64(atttmp["yaw"])
			if err != nil {
				log.Warn().Msgf("Error parsing float64: %v", err.Error())
			} else {
				nav.Yaw = RadiansToDegrees(floatTmp)
			}
			floatTmp, err = ParseFloat64(atttmp["pitch"])
			if err != nil {
				log.Warn().Msgf("Error parsing float64: %v", err.Error())
			} else {
				nav.Pitch = RadiansToDegrees(floatTmp)
			}

			floatTmp, err = ParseFloat64(atttmp["roll"])
			if err != nil {
				log.Warn().Msgf("Error parsing float64: %v", err.Error())
			} else {
				nav.Roll = RadiansToDegrees(floatTmp)
			}
		}
	case "speedThroughWater":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			nav.STW = MetersPerSecondToKnots(floatTmp)
		}
	case "speedThroughWaterReferenceType":
		break
	case "log":
		break
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(nav.Source)]
	if ok {
		nav.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", nav.Source)
	}
	if nav.Timestamp.IsZero() {
		nav.Timestamp = time.Now()
	}
	if !nav.IsEmpty() {
		nav.LogJSON()
		if SharedSubscriptionConfig.Repost {
			PublishClientMessage(client,
				SharedSubscriptionConfig.RepostRootTopic+"vessel/navigation/"+nav.Source+"/"+measurement, nav.ToJSON(), true)
		}
		if SharedSubscriptionConfig.InfluxEnabled {
			p := nav.ToInfluxPoint()
			err := SharedInfluxWriteAPI.WritePoint(context.Background(), p)
			if err != nil {
				log.Warn().Msgf("Error writing to influx: %v", err.Error())
			}
			log.Trace().Msg("Wrote Point")
			if SharedSubscriptionConfig.NavLogEn {
				log.Debug().Msg("Wrote Point")
			}
		}
	}
}

func (meas Navigation) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas Navigation) LogJSON() {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	log.Trace().Msgf("Navigation: %v", string(jsonData))
	if SharedSubscriptionConfig.NavLogEn {
		log.Info().Msgf("Navigation: %v", string(jsonData))
	}
}

// Since we are dropping fields we can end up with messages that are just a Source and a Timestamp
// This allows us to drop those messages
func (meas Navigation) IsEmpty() bool {
	if meas.Lat == 0.0 && meas.Lon == 0.0 && meas.Alt == 0.0 && meas.SOG == 0.0 && meas.ROT == 0.0 && meas.COGTrue == 0.0 &&
		meas.HeadingMag == 0.0 && meas.MagVariation == 0.0 && meas.MagDeviation == 0.0 && meas.Yaw == 0.0 &&
		meas.Pitch == 0.0 && meas.Roll == 0.0 && meas.HeadingTrue == 0.0 && meas.STW == 0.0 {
		return true
	}
	return false
}

func (meas Navigation) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	if meas.Source != "" {
		tagTmp["Source"] = meas.Source
	}
	return tagTmp
}

func (meas Navigation) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.Lat != 0.0 {
		measTmp["Latitude"] = meas.Lat
	}
	if meas.Lon != 0.0 {
		measTmp["Longitude"] = meas.Lon
	}
	if meas.Alt != 0.0 {
		measTmp["Altitude"] = meas.Alt
	}
	if meas.SOG != 0.0 {
		measTmp["SpeedOverGround"] = meas.SOG
	}
	if meas.ROT != 0.0 {
		measTmp["RateOfTurn"] = meas.ROT
	}
	if meas.COGTrue != 0.0 {
		measTmp["CourseOverGroundTrue"] = meas.COGTrue
	}
	if meas.HeadingMag != 0.0 {
		measTmp["HeadingMagnetic"] = meas.HeadingMag
	}
	if meas.MagVariation != 0.0 {
		measTmp["MagneticVariation"] = meas.MagVariation
	}
	if meas.MagDeviation != 0.0 {
		measTmp["MagneticDeviation"] = meas.MagDeviation
	}
	if meas.Yaw != 0.0 {
		measTmp["Yaw"] = meas.Yaw
	}
	if meas.Pitch != 0.0 {
		measTmp["Pitch"] = meas.Pitch
	}
	if meas.Roll != 0.0 {
		measTmp["Roll"] = meas.Roll
	}
	if meas.HeadingTrue != 0.0 {
		measTmp["HeadingTrue"] = meas.HeadingTrue
	}
	if meas.STW != 0.0 {
		measTmp["SpeedThroughWater"] = meas.STW
	}

	return measTmp
}

func (meas Navigation) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("navigation", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}
