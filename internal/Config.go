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
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type MQTTDestination struct {
	Host     string
	Topics   []string
	Username string
	Password string
	CACert   []byte
}

type GlobalConfig struct {
	Interval          int
	PublishTimeout    int
	DisconnectTimeout int
}

func LoadPublishServerConfig() ([]MQTTDestination, error) {
	var destinations []MQTTDestination
	if !viper.IsSet("pubservers") {
		log.Error().Msg("No Publish Servers Configured")
		return nil, errors.New("no publish servers set in viper config")
	}
	genericConf := viper.Get("pubservers")
	// Type assertion to convert 'any' to 'map[string]interface{}'
	serverConf, ok := genericConf.(map[string]interface{})
	if !ok {
		log.Error().Msg("Conversion failed: Publish Server config is not formatted correctly")
		return nil, errors.New("publish server configuration formatting invalid")
	}

	for k := range serverConf {
		dest := MQTTDestination{}
		log.Debug().Msgf("Server: %v", k)
		dest.Host = k
		if !viper.IsSet("pubservers." + k + ".topics") {
			log.Error().Msgf("No Topics Configured for host %v", k)
			return nil, fmt.Errorf("no topics set for host %v", k)
		}
		topics := viper.GetStringSlice("pubservers." + k + ".topics")
		for _, topic := range topics {
			log.Debug().Msgf("Topic %v", topic)
			dest.Topics = append(dest.Topics, topic)
		}

		if viper.IsSet("pubservers." + k + ".username") {
			dest.Username = viper.GetString("pubservers." + k + ".username")
			log.Debug().Msgf("Username %v", dest.Username)
		}

		if viper.IsSet("pubservers." + k + ".password") {
			dest.Password = viper.GetString("pubservers." + k + ".password")
			log.Trace().Msgf("Password %v", dest.Password)
		}

		if viper.IsSet("pubservers." + k + ".cafile") {
			cafilename := viper.GetString("pubservers." + k + ".cafile")
			log.Debug().Msgf("Using CA File %v", cafilename)
			cabytes, err := os.ReadFile(cafilename)
			if err != nil {
				log.Error().Msgf("Error Reading Defined CA File: %v", err.Error())
				return nil, fmt.Errorf("unable to read CA file %v", err)
			}
			log.Debug().Msgf("Loaded CAFile %v", cafilename)
			dest.CACert = cabytes
		}
		destinations = append(destinations, dest)
	}
	return destinations, nil
}

func LoadGlobalConfig() (GlobalConfig, error) {
	globalConf := GlobalConfig{}
	if !viper.IsSet("interval") {
		log.Error().Msg("Interval not configured")
		return GlobalConfig{}, errors.New("interval not set")
	}
	globalConf.Interval = viper.GetInt("interval")
	if !(globalConf.Interval > 0) {
		log.Error().Msgf("Interval set to invalid value: %v", globalConf.Interval)
		return GlobalConfig{}, fmt.Errorf("interval set to invalid value %v", globalConf.Interval)
	}
	log.Debug().Msgf("Interval Set to: %v", globalConf.Interval)
	if !viper.IsSet("publishtimeout") {
		log.Error().Msg("Publish Timeout not configured")
		return GlobalConfig{}, errors.New("publishtimeout not set")
	}
	globalConf.PublishTimeout = viper.GetInt("publishtimeout")
	if !(globalConf.PublishTimeout > 0) {
		log.Error().Msgf("Publish Timeout set to invalid value: %v", globalConf.PublishTimeout)
		return GlobalConfig{}, fmt.Errorf("publishtimeout set to invalid value %v", globalConf.PublishTimeout)
	}
	log.Debug().Msgf("Publish Timeout Set to: %v", globalConf.PublishTimeout)
	if !viper.IsSet("disconnecttimeout") {
		log.Error().Msg("Disconnect Timeout not configured")
		return GlobalConfig{}, errors.New("disconnecttimeout not set")
	}
	globalConf.DisconnectTimeout = viper.GetInt("disconnecttimeout")
	if !(globalConf.DisconnectTimeout > 0) {
		log.Error().Msgf("Disconnect Timeout set to invalid value: %v", globalConf.DisconnectTimeout)
		return GlobalConfig{}, fmt.Errorf("disconnecttimeout set to invalid value %v", globalConf.DisconnectTimeout)
	}
	log.Debug().Msgf("Disconnect Timeout Set to: %v", globalConf.DisconnectTimeout)
	return globalConf, nil
}
