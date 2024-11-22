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

type Outside struct {
	Source       string    `json:"Source,omitempty"`
	TempF        float64   `json:"TempF,omitempty"`
	Pressure     float64   `json:"Pressure,omitempty"`
	PressureInHg float64   `json:"PressureInHg,omitempty"`
	Timestamp    time.Time `json:"Timestamp,omitempty"`
}

func OnOutsideMessage(client MQTT.Client, message MQTT.Message) {
	go handleOutsideMessage(client, message)
}

func handleOutsideMessage(client MQTT.Client, message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.OutsideLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.OutsideLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	out := Outside{}
	out.Source = rawData["$source"].(string)
	out.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "temperature":
		out.TempF = KelvinToFarenheit(rawData["value"].(float64))
	case "pressure":
		out.Pressure = rawData["value"].(float64) / 100
		out.PressureInHg = MillibarToInHg(out.Pressure)
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(out.Source)]
	if ok {
		out.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", out.Source)
	}
	if out.Timestamp.IsZero() {
		out.Timestamp = time.Now()
	}
	if !out.IsEmpty() {
		out.LogJSON()
		if SharedSubscriptionConfig.Repost {
			PublishClientMessage(client,
				SharedSubscriptionConfig.RepostRootTopic+"vessel/environment/outside/"+out.Source+"/"+measurement, out.ToJSON())
		}
		if SharedSubscriptionConfig.InfluxEnabled {
			log.Trace().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
				SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			if SharedSubscriptionConfig.OutsideLogEn {
				log.Debug().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
					SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			}
			// Sharing a client across threads did not seem to work
			// So will create a client each time for now
			client := influxdb2.NewClient(SharedSubscriptionConfig.InfluxUrl, SharedSubscriptionConfig.InfluxToken)
			writeAPI := client.WriteAPIBlocking(SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			p := out.ToInfluxPoint()
			writeAPI.WritePoint(context.Background(), p)
			client.Close()
			log.Trace().Msg("Wrote Point")
			if SharedSubscriptionConfig.OutsideLogEn {
				log.Debug().Msg("Wrote Point")
			}
		}
	}
}

func (meas Outside) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas Outside) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Outside: %v", json)
	if SharedSubscriptionConfig.OutsideLogEn {
		log.Info().Msgf("Outside: %v", json)
	}
}

// Since we are dropping fields we can end up with messages that are just a Source and a Timestamp
// This allows us to drop those messages
func (meas Outside) IsEmpty() bool {
	if meas.TempF == 0.0 && meas.Pressure == 0.0 && meas.PressureInHg == 0.0 {
		return true
	}
	return false
}

func (meas Outside) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	if meas.Source != "" {
		tagTmp["Source"] = meas.Source
	}
	return tagTmp
}

func (meas Outside) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.TempF != 0.0 {
		measTmp["TempF"] = meas.TempF
	}
	if meas.Pressure != 0.0 {
		measTmp["Pressure"] = meas.Pressure
	}

	return measTmp
}

func (meas Outside) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("outside", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}
