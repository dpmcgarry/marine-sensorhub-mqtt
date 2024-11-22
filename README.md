# marine-sensorhub-mqtt

Go Daemon to process MQTT messages from Marine Sensorhub / SignalK / Victron Cerbo

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/BAoeA8hPhLZPsitWVrJzUa/R3FbSRfu28FC3tEmqFMr83/tree/main.svg?style=shield)](https://dl.circleci.com/status-badge/redirect/circleci/BAoeA8hPhLZPsitWVrJzUa/R3FbSRfu28FC3tEmqFMr83/tree/main)

## InfluxDB Schema

Ref: <https://docs.influxdata.com/influxdb/v2/write-data/best-practices/schema-design/>

| measurement | tag keys | field keys |
| -------- | ------- | ------- |
| bleTemperature | MAC, Location, | TempF, BatteryPct, Humidity, RSSI |
| phyTemperature | MAC, Location, Device, Component | TempF |
| espStatus | MAC, Location, IPAddress, MSHVersion | FreeSRAM, FreeHeap, FreePSRAM, WiFiReconnectCount, MQTTReconnectCount, BLEEnabled, RTDEnabled, WiFiRSSI, HasTime, HasResetMQTT |
| navigation | Source | latitude, longitude, SOG, ROT, COGTrue, HeadingMag, MagVariation, MagDeviation, Attitude, HeadingTrue, STW |
| gnss | Source | AntennaAlt, Satellites, HozDilution, PosDilution, GeoidalSep, Type, MethodQuality, SatsInView |
| steering | Source | RudderAngle, AutopilotState, TargetHeadingMag |
| wind | Source | SpeedApp, AngApp, SOG, DirectionTrue |
| water | Source | TempF, DepthUnderTransducerFt |
| outside | Source | TempF, Pressure |
| propulsion | Device, Source | RPM, BoostPSI, OilTempF, OilPressure, CoolantTempF, RunTime, EngineLoad, EngineTorque, TransOilTempF, TransOilPressure, AltVoltage, FuelRate |

TBD: Notifications

## TODO

* ~~Add logging overrides~~
* ~~Cleanup MAC / N2K in Global and Subscription~~
* Unit tests
* ~~Change Handlers to Async~~
* ~~Add InfluxDB~~
* Check to see if InfluxClient can be shared across threads
* Check on what provides altitude and if it is getting lost somehow
* ~~Add MQTT Repost~~
* Look at trimming down number of items
* ~~Handle Unit Conversions~~
* Look at having two log files (Warn+ and Info/Debug)
* Add a message archiving capability
