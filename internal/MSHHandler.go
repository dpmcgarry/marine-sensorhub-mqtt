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
	"encoding/json"
	"reflect"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

func OnMSHMessage(client MQTT.Client, message MQTT.Message) {
	log.Debug().Msgf("Got a message from: %v", message.Topic())
	msg := string(message.Payload())
	log.Trace().Msgf("Message: %v", msg)
	jsonData := make(map[string]any)
	err := json.Unmarshal(message.Payload(), &jsonData)
	if err != nil {
		log.Warn().Msgf("Error unmarshalling JSON for topic: %v error: %v", message.Topic(), err.Error())
	}
	log.Debug().Msgf("JSON: %v", jsonData)
	meas := SensorMeasurement{}
	for k, v := range jsonData {
		log.Debug().Msgf("Key: %v", k)
		valType := reflect.TypeOf(v)
		log.Debug().Msgf("TypeOf: %v", valType)
		switch valType.Kind() {
		case reflect.Bool:
			log.Debug().Msg("Bool")
			meas.BoolFields[k] = v.(bool)
		case reflect.String:
			log.Debug().Msg("String")
			meas.StringFields[k] = v.(string)
		case reflect.Float64:
			if v == float64(int64(v.(float64))) {
				log.Debug().Msg("Int64")
				meas.IntFields[k] = v.(int64)
			} else {
				log.Debug().Msg("Float64")
				meas.FloatFields[k] = v.(float64)
			}
		default:
			log.Warn().Msgf("Ruh Roh")
		}
	}
	if meas.Timestamp.IsZero() {
		log.Warn().Msg("Timestamp is nil")
	}
}
