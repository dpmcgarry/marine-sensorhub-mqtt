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

type Wind struct {
	Source        string    `json:"Source,omitempty"`
	SpeedApp      float64   `json:"SpeedApp,omitempty"`
	AngleApp      float64   `json:"AngleApp,omitempty"`
	SOG           float64   `json:"SOG,omitempty"`
	DirectionTrue float64   `json:"DirectionTrue,omitempty"`
	Timestamp     time.Time `json:"Timestamp,omitempty"`
}

func OnWindMessage(client MQTT.Client, message MQTT.Message) {
	go handleWindMessage(client, message)
}

func handleWindMessage(client MQTT.Client, message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.WindLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.WindLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	wind := Wind{}
	var strtmp string
	strtmp, err = ParseString(rawData["$source"])
	if err != nil {
		log.Warn().Msgf("Error parsing string: %v", err.Error())
	} else {
		wind.Source = strtmp
	}
	strtmp, err = ParseString(rawData["timestamp"])
	if err != nil {
		log.Warn().Msgf("Error parsing string: %v", err.Error())
	} else {
		wind.Timestamp, err = time.Parse(ISOTimeLayout, strtmp)
		if err != nil {
			log.Warn().Msgf("Error parsing time string: %v", err.Error())
		}
	}
	var floatTmp float64
	switch measurement {
	case "speedOverGround":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			wind.SOG = MetersPerSecondToKnots(floatTmp)
		}
	case "directionTrue":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			wind.DirectionTrue = RadiansToDegrees(floatTmp)
		}
	case "speedApparent":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			wind.SpeedApp = MetersPerSecondToKnots(floatTmp)
		}
	case "angleApparent":
		floatTmp, err = ParseFloat64(rawData["value"])
		if err != nil {
			log.Warn().Msgf("Error parsing float64: %v", err.Error())
		} else {
			wind.AngleApp = RadiansToDegrees(floatTmp)
		}
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(wind.Source)]
	if ok {
		wind.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", wind.Source)
	}
	if wind.Timestamp.IsZero() {
		wind.Timestamp = time.Now()
	}
	if !wind.IsEmpty() {
		wind.LogJSON()
		if SharedSubscriptionConfig.Repost {
			PublishClientMessage(client,
				SharedSubscriptionConfig.RepostRootTopic+"vessel/environment/wind/"+wind.Source+"/"+measurement, wind.ToJSON(), true)
		}
		if SharedSubscriptionConfig.InfluxEnabled {
			p := wind.ToInfluxPoint()
			err := SharedInfluxWriteAPI.WritePoint(context.Background(), p)
			if err != nil {
				log.Warn().Msgf("Error writing to influx: %v", err.Error())
			}
			log.Trace().Msg("Wrote Point")
			if SharedSubscriptionConfig.WindLogEn {
				log.Debug().Msg("Wrote Point")
			}
		}
	}
}

func (meas Wind) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas Wind) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Wind: %v", json)
	if SharedSubscriptionConfig.WindLogEn {
		log.Info().Msgf("Wind: %v", json)
	}
}

// Since we are dropping fields we can end up with messages that are just a Source and a Timestamp
// This allows us to drop those messages
func (meas Wind) IsEmpty() bool {
	if meas.SpeedApp == 0.0 && meas.AngleApp == 0.0 && meas.SOG == 0.0 && meas.DirectionTrue == 0.0 {
		return true
	}
	return false
}

func (meas Wind) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	if meas.Source != "" {
		tagTmp["Source"] = meas.Source
	}
	return tagTmp
}

func (meas Wind) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.SpeedApp != 0.0 {
		measTmp["SpeedApp"] = meas.SpeedApp
	}
	if meas.AngleApp != 0.0 {
		measTmp["AngleApp"] = meas.AngleApp
	}
	if meas.SOG != 0.0 {
		measTmp["SOG"] = meas.SOG
	}
	if meas.DirectionTrue != 0.0 {
		measTmp["DirectionTrue"] = meas.DirectionTrue
	}

	return measTmp
}

func (meas Wind) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("wind", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}
