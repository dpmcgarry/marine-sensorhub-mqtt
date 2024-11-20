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
	"strconv"

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

type SubscriptionConfig struct {
	Host             string
	Username         string
	Password         string
	CACert           []byte
	ESPMSHRootTopic  string
	SignalKRootTopic string
	CerboRootTopic   string
	Repost           bool
	RepostRootTopic  string
	PublishTimeout   uint
	MACtoLocation    map[string]string
	N2KtoName        map[string]string
	BLESubEn         bool
	GNSSSubEn        bool
	ESPSubEn         bool
	NavSubEn         bool
	OutsideSubEn     bool
	PHYSubEn         bool
	PropSubEn        bool
	SteerSubEn       bool
	WaterSubEn       bool
	WindSubEn        bool
	BLELogEn         bool
	GNSSLogEn        bool
	ESPLogEn         bool
	NavLogEn         bool
	OutsideLogEn     bool
	PHYLogEn         bool
	PropLogEn        bool
	SteerLogEn       bool
	WaterLogEn       bool
	WindLogEn        bool
}

type PublishConfig struct {
	Interval          int
	PublishTimeout    int
	DisconnectTimeout int
}

func LoadPublishConfig() (PublishConfig, error) {
	publishConf := PublishConfig{}
	if !viper.IsSet("publish.interval") {
		log.Error().Msg("Interval not configured")
		return PublishConfig{}, errors.New("interval not set")
	}
	publishConf.Interval = viper.GetInt("publish.interval")
	if !(publishConf.Interval > 0) {
		log.Error().Msgf("Interval set to invalid value: %v", publishConf.Interval)
		return PublishConfig{}, fmt.Errorf("interval set to invalid value %v", publishConf.Interval)
	}
	log.Debug().Msgf("Interval Set to: %v", publishConf.Interval)
	if !viper.IsSet("publish.timeout") {
		log.Error().Msg("Publish Timeout not configured")
		return PublishConfig{}, errors.New("publishtimeout not set")
	}
	publishConf.PublishTimeout = viper.GetInt("publish.timeout")
	if !(publishConf.PublishTimeout > 0) {
		log.Error().Msgf("Publish Timeout set to invalid value: %v", publishConf.PublishTimeout)
		return PublishConfig{}, fmt.Errorf("publishtimeout set to invalid value %v", publishConf.PublishTimeout)
	}
	log.Debug().Msgf("Publish Timeout Set to: %v", publishConf.PublishTimeout)
	if !viper.IsSet("publish.disconnecttimeout") {
		log.Error().Msg("Disconnect Timeout not configured")
		return PublishConfig{}, errors.New("disconnecttimeout not set")
	}
	publishConf.DisconnectTimeout = viper.GetInt("publish.disconnecttimeout")
	if !(publishConf.DisconnectTimeout > 0) {
		log.Error().Msgf("Disconnect Timeout set to invalid value: %v", publishConf.DisconnectTimeout)
		return PublishConfig{}, fmt.Errorf("disconnecttimeout set to invalid value %v", publishConf.DisconnectTimeout)
	}
	log.Debug().Msgf("Disconnect Timeout Set to: %v", publishConf.DisconnectTimeout)
	return publishConf, nil
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

func LoadSubscribeServerConfig() (SubscriptionConfig, error) {
	subConf := SubscriptionConfig{}
	subConf.BLESubEn = true
	subConf.GNSSSubEn = true
	subConf.ESPSubEn = true
	subConf.NavSubEn = true
	subConf.OutsideSubEn = true
	subConf.PHYSubEn = true
	subConf.PropSubEn = true
	subConf.SteerSubEn = true
	subConf.WaterSubEn = true
	subConf.WindSubEn = true
	subConf.BLELogEn = false
	subConf.GNSSLogEn = false
	subConf.ESPLogEn = false
	subConf.NavLogEn = false
	subConf.OutsideLogEn = false
	subConf.PHYLogEn = false
	subConf.PropLogEn = false
	subConf.SteerLogEn = false
	subConf.WaterLogEn = false
	subConf.WindLogEn = false
	if !viper.IsSet("subscription") {
		log.Error().Msg("No Subscription Information Configured")
		return SubscriptionConfig{}, errors.New("no subscription information set in viper config")
	}
	subscriptionMap := viper.GetStringMapString("subscription")
	v, ok := subscriptionMap["server"]
	if ok {
		log.Debug().Msgf("Setting host: %v", v)
		subConf.Host = v
	} else {
		log.Error().Msg("Server is required but is not configured")
		return SubscriptionConfig{}, errors.New("server is required but is not set in viper config")
	}
	var configItems = []string{"esp-msh-root-topic", "signalk-root-topic", "cerbo-root-topic", "username", "password", "cafile", "repost", "repost-root-topic", "publish-timeout"}
	for _, confItem := range configItems {
		v, ok = subscriptionMap[confItem]
		if ok {
			log.Debug().Msgf("Setting %v: %v", confItem, v)
			switch confItem {
			case "esp-msh-root-topic":
				subConf.ESPMSHRootTopic = v
			case "signalk-root-topic":
				subConf.SignalKRootTopic = v
			case "cerbo-root-topic":
				subConf.CerboRootTopic = v
			case "username":
				subConf.Username = v
			case "password":
				subConf.Password = v
			case "cafile":
				log.Debug().Msgf("Using CA File %v", v)
				cabytes, err := os.ReadFile(v)
				if err != nil {
					log.Warn().Msgf("Error Reading Defined CA File: %v", err.Error())
				}
				log.Debug().Msgf("Loaded CAFile %v", v)
				subConf.CACert = cabytes
			case "repost":
				tmpbool, err := strconv.ParseBool(v)
				if err != nil {
					log.Warn().Msgf("Error parsing boolean from config: %v", err.Error())
				} else {
					subConf.Repost = tmpbool
				}
			case "repost-root-topic":
				subConf.RepostRootTopic = v
			case "publish-timeout":
				inttmp, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					log.Warn().Msgf("Error parsing publish-timeout will use default: %v", err.Error())
				} else {
					subConf.PublishTimeout = uint(inttmp)
				}
			}
		} else {
			log.Trace().Msgf("%v not found. Continuing", confItem)
		}
	}

	if !viper.IsSet("subscription.MACtoName") {
		log.Warn().Msg("MAC to Location Mappings not found")
	} else {
		log.Debug().Msg("Loading MAC to Location Mappings")
		subConf.MACtoLocation = viper.GetStringMapString("subscription.MACtoName")
	}

	if !viper.IsSet("subscription.topic-overrides") {
		log.Debug().Msg("Subscription topic overrides not found")
		return subConf, nil
	} else {
		log.Debug().Msg("Subscription topics overrides found")

		tmpmap := viper.GetStringMap("subscription.topic-overrides")
		for k, v := range tmpmap {
			switch k {
			case "ble":
				subConf.BLESubEn = v.(bool)
			case "gnss":
				subConf.GNSSSubEn = v.(bool)
			case "esp":
				subConf.ESPSubEn = v.(bool)
			case "nav":
				subConf.NavSubEn = v.(bool)
			case "outside":
				subConf.OutsideSubEn = v.(bool)
			case "phy":
				subConf.PHYSubEn = v.(bool)
			case "propulsion":
				subConf.PropSubEn = v.(bool)
			case "steering":
				subConf.SteerSubEn = v.(bool)
			case "water":
				subConf.WaterSubEn = v.(bool)
			case "wind":
				subConf.WindSubEn = v.(bool)
			default:
				log.Warn().Msgf("Invalid Key %v found in EnableSubscriptions", k)
			}
		}
	}

	if !viper.IsSet("subscription.verbose-topic-logging") {
		log.Debug().Msg("Subscription logging overrides not found")
		return subConf, nil
	} else {
		log.Debug().Msg("Subscription logging overrides found")

		tmpmap := viper.GetStringMap("subscription.verbose-topic-logging")
		for k, v := range tmpmap {
			switch k {
			case "ble":
				subConf.BLELogEn = v.(bool)
			case "gnss":
				subConf.GNSSLogEn = v.(bool)
			case "esp":
				subConf.ESPLogEn = v.(bool)
			case "nav":
				subConf.NavLogEn = v.(bool)
			case "outside":
				subConf.OutsideLogEn = v.(bool)
			case "phy":
				subConf.PHYLogEn = v.(bool)
			case "propulsion":
				subConf.PropLogEn = v.(bool)
			case "steering":
				subConf.SteerLogEn = v.(bool)
			case "water":
				subConf.WaterLogEn = v.(bool)
			case "wind":
				subConf.WindLogEn = v.(bool)
			default:
				log.Warn().Msgf("Invalid Key %v found in VerboseSubscriptionLogging", k)
			}
		}
	}

	if !viper.IsSet("subscription.N2KtoName") {
		log.Warn().Msg("N2K to Name Mappings not found")
		return subConf, nil
	} else {
		log.Debug().Msg("Loading N2K to Device Name Mappings")
		subConf.N2KtoName = viper.GetStringMapString("subscription.N2KtoName")
	}

	return subConf, nil
}
