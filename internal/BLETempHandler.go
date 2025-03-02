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

type BLETemperature struct {
	MAC            string    `json:"MAC,omitempty"`
	Location       string    `json:"Location,omitempty"`
	TempF          float64   `json:"TempF,omitempty"`
	BatteryPercent float64   `json:"BatteryPct,omitempty"`
	Humidity       float64   `json:"Humidity,omitempty"`
	RSSI           int64     `json:"RSSI,omitempty"`
	Timestamp      time.Time `json:"Timestamp,omitempty"`
}

func OnBLETemperatureMessage(client MQTT.Client, message MQTT.Message) {
	go handleBLETemperatureMessage(client, message)
}

func handleBLETemperatureMessage(client MQTT.Client, message MQTT.Message) {
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.BLELogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	bleTemp := BLETemperature{}
	err := json.Unmarshal(message.Payload(), &bleTemp)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}

	loc, ok := SharedSubscriptionConfig.MACtoLocation[strings.ToLower(bleTemp.MAC)]
	if ok {
		bleTemp.Location = loc
	} else {
		log.Warn().Msgf("Location not found for MAC %v", bleTemp.MAC)
	}
	if bleTemp.Timestamp.IsZero() {
		bleTemp.Timestamp = time.Now()
	}
	bleTemp.LogJSON()
	if SharedSubscriptionConfig.Repost {
		PublishClientMessage(client, SharedSubscriptionConfig.RepostRootTopic+"ble/temperature/"+bleTemp.Location, bleTemp.ToJSON(), true)
	}
	if SharedSubscriptionConfig.InfluxEnabled {
		log.Trace().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
			SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
		if SharedSubscriptionConfig.BLELogEn {
			log.Debug().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
				SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
		}
		writeAPI := SharedInfluxClient.WriteAPIBlocking(SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
		p := bleTemp.ToInfluxPoint()
		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			log.Warn().Msgf("Error writing to influx: %v", err.Error())
		}
		log.Trace().Msg("Wrote Point")
		if SharedSubscriptionConfig.BLELogEn {
			log.Debug().Msg("Wrote Point")
		}
	}
}

func (meas BLETemperature) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas BLETemperature) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("BLETemp: %v", json)
	if SharedSubscriptionConfig.BLELogEn {
		log.Info().Msgf("BLETemp: %v", json)
	}
}

func (meas BLETemperature) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	tagTmp["MAC"] = meas.MAC
	if meas.Location != "" {
		tagTmp["Location"] = meas.Location
	}
	return tagTmp
}

func (meas BLETemperature) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.TempF != 0.0 {
		measTmp["TempF"] = meas.TempF
	}
	if meas.BatteryPercent != 0.0 {
		measTmp["BatteryPercent"] = meas.BatteryPercent
	}
	if meas.Humidity != 0.0 {
		measTmp["Humidity"] = meas.Humidity
	}
	if meas.RSSI != 0 {
		measTmp["RSSI"] = meas.RSSI
	}
	return measTmp
}

func (meas BLETemperature) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("bleTemperature", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}
