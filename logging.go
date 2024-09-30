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

package main

import (
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zerolog.Logger
}

func configureLogging() *Logger {
	var writers []io.Writer

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	writers = append(writers, newRollingFile())
	mw := io.MultiWriter(writers...)

	logger := zerolog.New(mw).With().Timestamp().Logger()
	// Adds the ability to change the logging level using an environment variable
	level := strings.ToUpper(os.Getenv("LOGLEVEL"))
	switch level {
	case "TRACE":
		log.Logger = zerolog.New(mw).
			Level(zerolog.TraceLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	case "DEBUG":
		log.Logger = zerolog.New(mw).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	case "INFO":
		log.Logger = zerolog.New(mw).
			Level(zerolog.InfoLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	case "WARN":
		log.Logger = zerolog.New(mw).
			Level(zerolog.WarnLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	case "ERROR":
		log.Logger = zerolog.New(mw).
			Level(zerolog.ErrorLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	default: // Default to Debug level
		log.Logger = zerolog.New(mw).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	return &Logger{
		Logger: &logger,
	}
}

func newRollingFile() io.Writer {
	return &lumberjack.Logger{
		Filename:   path.Join(viper.GetString("logdir") + "mqtt-keepalive.log"),
		MaxBackups: 5,
		MaxSize:    20,
		MaxAge:     30,
	}
}
