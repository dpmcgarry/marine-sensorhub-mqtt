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

type Propulsion struct {
	Device           string    `json:"Device,omitempty"`
	Source           string    `json:"Source,omitempty"`
	RPM              int64     `json:"RPM,omitempty"`
	BoostPSI         float64   `json:"BoostPSI,omitempty"`
	OilTempF         float64   `json:"OilTempF,omitempty"`
	OilPressure      float64   `json:"OilPressure,omitempty"`
	CoolantTempF     float64   `json:"CoolantTempF,omitempty"`
	RunTime          int64     `json:"RunTime,omitempty"`
	EngineLoad       float64   `json:"EngineLoad,omitempty"`
	EngineTorque     float64   `json:"EngineTorque,omitempty"`
	TransOilTempF    float64   `json:"TransOilTemp,omitempty"`
	TransOilPressure float64   `json:"TransOilPressure,omitempty"`
	AltVoltage       float64   `json:"AlternatorVoltage,omitempty"`
	FuelRate         float64   `json:"FuelRate,omitempty"`
	Timestamp        time.Time `json:"Timestamp,omitempty"`
}

func OnPropulsionMessage(client MQTT.Client, message MQTT.Message) {
	go handlePropulsionMessage(client, message)
}

func handlePropulsionMessage(client MQTT.Client, message MQTT.Message) {
	// TODO: Support Multiple Engines
	log.Trace().Msgf("Got a message from: %v", message.Topic())
	if SharedSubscriptionConfig.PropLogEn {
		log.Info().Msgf("Got a message from: %v", message.Topic())
	}
	measurement := message.Topic()[strings.LastIndex(message.Topic(), "/")+1:]
	log.Trace().Msgf("Got Measurement: %v", measurement)
	if SharedSubscriptionConfig.PropLogEn {
		log.Info().Msgf("Got Measurement: %v", measurement)
	}
	isTranny := false
	if strings.Contains(message.Topic(), "/transmission/") {
		isTranny = true
	}
	var rawData map[string]any
	err := json.Unmarshal(message.Payload(), &rawData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	prop := Propulsion{}
	prop.Source = rawData["$source"].(string)
	prop.Timestamp, err = time.Parse(ISOTimeLayout, rawData["timestamp"].(string))
	if err != nil {
		log.Warn().Msgf("Error parsing time string: %v", err.Error())
	}
	switch measurement {
	case "revolutions":
		prop.RPM = int64(rawData["value"].(float64)) * 60
	case "boostPressure":
		prop.BoostPSI = PascalToPSI(rawData["value"].(float64))
	case "oilTemperature":
		if isTranny {
			prop.TransOilTempF = KelvinToFarenheit(rawData["value"].(float64))
		} else {
			prop.OilTempF = KelvinToFarenheit(rawData["value"].(float64))
		}
	case "oilPressure":
		if isTranny {
			prop.TransOilPressure = PascalToPSI(rawData["value"].(float64))
		} else {
			prop.OilPressure = PascalToPSI(rawData["value"].(float64))
		}
	case "temperature":
		prop.CoolantTempF = KelvinToFarenheit(rawData["value"].(float64))
	case "alternatorVoltage":
		prop.AltVoltage = rawData["value"].(float64)
	case "transmission":
		break
	case "fuel":
		break
	case "rate":
		prop.FuelRate = CubicMetersPerSecondToGallonsPerHour(rawData["value"].(float64))
	case "runTime":
		prop.RunTime = int64(rawData["value"].(float64))
	case "engineLoad":
		prop.EngineLoad = rawData["value"].(float64) * 100
	case "engineTorque":
		prop.EngineTorque = rawData["value"].(float64) * 100
	default:
		log.Warn().Msgf("Unknown measurement %v in %v", measurement, message.Topic())
	}
	name, ok := SharedSubscriptionConfig.N2KtoName[strings.ToLower(prop.Source)]
	if ok {
		prop.Source = name
	} else {
		log.Warn().Msgf("Name not found for Source %v", prop.Source)
	}
	if prop.Timestamp.IsZero() {
		prop.Timestamp = time.Now()
	}
	if !prop.IsEmpty() {
		prop.LogJSON()
		if SharedSubscriptionConfig.Repost {
			PublishClientMessage(client,
				SharedSubscriptionConfig.RepostRootTopic+"vessel/propulsion/"+prop.Source+"/"+measurement, prop.ToJSON())
		}
		if SharedSubscriptionConfig.InfluxEnabled {
			log.Trace().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
				SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			if SharedSubscriptionConfig.PropLogEn {
				log.Debug().Msgf("InfluxDB is enabled. URL: %v Org: %v Bucket:%v", SharedSubscriptionConfig.InfluxUrl,
					SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			}
			// Sharing a client across threads did not seem to work
			// So will create a client each time for now
			client := influxdb2.NewClient(SharedSubscriptionConfig.InfluxUrl, SharedSubscriptionConfig.InfluxToken)
			writeAPI := client.WriteAPIBlocking(SharedSubscriptionConfig.InfluxOrg, SharedSubscriptionConfig.InfluxBucket)
			p := prop.ToInfluxPoint()
			writeAPI.WritePoint(context.Background(), p)
			client.Close()
			log.Trace().Msg("Wrote Point")
			if SharedSubscriptionConfig.PropLogEn {
				log.Debug().Msg("Wrote Point")
			}
		}
	}
}

func (meas Propulsion) ToJSON() string {
	jsonData, err := json.Marshal(meas)
	if err != nil {
		log.Warn().Msgf("Error Serializing JSON: %v", err.Error())
	}
	return string(jsonData)
}

func (meas Propulsion) LogJSON() {
	json := meas.ToJSON()
	log.Trace().Msgf("Propulsion: %v", json)
	if SharedSubscriptionConfig.PropLogEn {
		log.Info().Msgf("Propulsion: %v", json)
	}
}

// Since we are dropping fields we can end up with messages that are just a Source and a Timestamp
// This allows us to drop those messages
func (meas Propulsion) IsEmpty() bool {
	if meas.RPM == 0 && meas.BoostPSI == 0.0 && meas.OilTempF == 0.0 && meas.OilPressure == 0.0 &&
		meas.CoolantTempF == 0.0 && meas.RunTime == 0 && meas.EngineLoad == 0.0 && meas.EngineTorque == 0.0 &&
		meas.TransOilTempF == 0.0 && meas.TransOilPressure == 0.0 && meas.AltVoltage == 0.0 && meas.FuelRate == 0.0 {
		return true
	}
	return false
}

func (meas Propulsion) GetInfluxTags() map[string]string {
	tagTmp := make(map[string]string)
	if meas.Source != "" {
		tagTmp["Source"] = meas.Source
	}
	if meas.Device != "" {
		tagTmp["Device"] = meas.Device
	}
	return tagTmp
}

func (meas Propulsion) GetInfluxFields() map[string]interface{} {
	measTmp := make(map[string]interface{})
	if meas.RPM != 0 {
		measTmp["RPM"] = meas.RPM
	}
	if meas.BoostPSI != 0.0 {
		measTmp["BoostPSI"] = meas.BoostPSI
	}
	if meas.OilTempF != 0.0 {
		measTmp["OilTempF"] = meas.OilTempF
	}
	if meas.OilPressure != 0.0 {
		measTmp["OilPressure"] = meas.OilPressure
	}
	if meas.CoolantTempF != 0.0 {
		measTmp["CoolantTempF"] = meas.CoolantTempF
	}
	if meas.RunTime != 0 {
		measTmp["RunTime"] = meas.RunTime
	}
	if meas.EngineLoad != 0.0 {
		measTmp["EngineLoad"] = meas.EngineLoad
	}
	if meas.EngineTorque != 0.0 {
		measTmp["EngineTorque"] = meas.EngineTorque
	}
	if meas.TransOilTempF != 0.0 {
		measTmp["TransOilTempF"] = meas.TransOilTempF
	}
	if meas.TransOilPressure != 0.0 {
		measTmp["TransOilPressure"] = meas.TransOilPressure
	}
	if meas.AltVoltage != 0.0 {
		measTmp["AlternatorVoltage"] = meas.AltVoltage
	}
	if meas.FuelRate != 0.0 {
		measTmp["FuelRate"] = meas.FuelRate
	}

	return measTmp
}

func (meas Propulsion) ToInfluxPoint() *write.Point {
	return influxdb2.NewPoint("propulsion", meas.GetInfluxTags(), meas.GetInfluxFields(), meas.Timestamp)
}
