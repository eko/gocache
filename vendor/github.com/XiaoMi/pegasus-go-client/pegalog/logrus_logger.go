// Copyright (c) 2017, Xiaomi, Inc.
// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pegalog

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogrusConfig is used to configure the generation of log files.
type LogrusConfig struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

// callerPrettifier simplifies the caller info
func callerPrettifier(f *runtime.Frame) (function string, file string) {
	function = f.Function[strings.LastIndex(f.Function, "/")+1:]
	file = fmt.Sprint(f.File[strings.LastIndex(f.File, "/")+1:], ":", f.Line)
	return function, file
}

// NewLogrusLogger creates a new LogrusLogger.
func NewLogrusLogger(cfg *LogrusConfig) Logger {
	l := logrus.New()
	l.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		FullTimestamp:    true,
		CallerPrettyfier: callerPrettifier,
	}
	l.Out = &lumberjack.Logger{
		Filename:  cfg.Filename,
		MaxSize:   cfg.MaxSize,
		MaxAge:    cfg.MaxAge,
		LocalTime: true,
	}
	l.ReportCaller = true
	return l
}

// DefaultLogrusLogger is a LogrusLogger instance with default configurations.
var DefaultLogrusLogger = NewLogrusLogger(&LogrusConfig{
	MaxSize:  500, // megabytes
	MaxAge:   5,   // days
	Filename: "./pegasus.log",
})
