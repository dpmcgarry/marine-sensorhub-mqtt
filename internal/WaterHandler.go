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
	"context"
	"encoding/json"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/rs/zerolog/log"
)

type Water struct {
	Source                 string    `json:"Source,omitempty"`
	TempF                  float64   `json:"TempF,omitempty"`
	DepthUnderTransducerFt float64   `json:"DepthUnderTransducerFt,omitempty"`
	Timestamp              time.Time `json:"Timestamp,omitempty"`
}

func OnWaterMessage(client MQTT.Client, message MQTT.Message) {
	go handleWaterMessage(client, message)
}

func handleWaterMessage(client MQTT.Client, message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.WaterLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.WaterLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	water := Water{}
	water.Source = rawData["$source"].(string)
	water.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "temperature":
		// My sensor reports in F but SK assumes it is C
		// So Converting from K to C actually gives F
		water.TempF = KelvinToCelsius(rawData["value"].(float64))
	case "belowTransducer":
		water.DepthUnderTransducerFt = MetersToFeet(rawData["value"].(float64))
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(water.Source)]
	if ok {
		water.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", water.Source)
	}
	if water.Timestamp.IsZero() {
		water.Timestamp = time.Now()
	}
	if !water.IsEmpty() {
		water.LogJSON()
		if SharedSubscriptionConfig.Repost {
			PublishClientMessage(client,
				SharedSubscriptionConfig.RepostRootTopic+"vessel/environment/water/"+water.Source+"/"+measurement, water.ToJSON(), true)
		}
		if SharedSubscriptionConfig.InfluxEnabled {
			log.Trace().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
				SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			if SharedSubscriptionConfig.WaterLogEn {
				log.Debug().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
					SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			}
			writeAPI := SharedInfluxClient.WriteAPIBlocking(SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			p := water.ToInfluxPoint()
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				log.Warn().Msgf("Error writing to influx: %v", err.Error())
			}
			log.Trace().Msg("Wrote Point")
			if SharedSubscriptionConfig.WaterLogEn {
				log.Debug().Msg("Wrote Point")
			}
		}
	}
}

func (meas Water) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas Water) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Water: %v", json)
	if SharedSubscriptionConfig.WaterLogEn {
		log.Info().Msgf("Water: %v", json)
	}
}

// Since we are dropping fields we can end up with messages that are just a Source and a Timestamp
// This allows us to drop those messages
func (meas Water) IsEmpty() bool {
	if meas.DepthUnderTransducerFt == 0.0 && meas.TempF == 0.0 {
		return true
	}
	return false
}

func (meas Water) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	if meas.Source != "" {
		tagTmp["Source"] = meas.Source
	}
	return tagTmp
}

func (meas Water) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.TempF != 0.0 {
		measTmp["TempF"] = meas.TempF
	}
	if meas.DepthUnderTransducerFt != 0.0 {
		measTmp["DepthUnderTransducerFt"] = meas.DepthUnderTransducerFt
	}

	return measTmp
}

func (meas Water) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("water", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}
