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
	"log"
	"os"
	"sync"
)

// The logger module in this file is inspired by etcd/clientv3/logger

// Logger is the internal logger served for pegasus go client.
// WARN: Don't use this logger for your application.
type Logger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
}

var (
	_logger settableLogger
)

type settableLogger struct {
	l  Logger
	mu sync.RWMutex
}

// StderrLogger is an implementation of Logger that outputs logs to stderr.
// WARN: Don't use it in your production environment. Lack of logs after failures will make it
// 		 significantly difficult to track the root cause.
var StderrLogger = log.New(os.Stderr, "", log.LstdFlags)

func init() {
	// by default we use stderr for logging
	_logger.set(DefaultLogrusLogger)
}

// SetLogger sets client-side Logger. By default, logs are disabled.
func SetLogger(l Logger) {
	_logger.set(l)
}

// GetLogger returns the current logger.
func GetLogger() Logger {
	return _logger.get()
}

func (s *settableLogger) set(l Logger) {
	s.mu.Lock()
	_logger.l = l
	s.mu.Unlock()
}

func (s *settableLogger) get() Logger {
	s.mu.RLock()
	l := _logger.l
	s.mu.RUnlock()
	return l
}

func (s *settableLogger) Fatal(args ...interface{})                 { s.get().Fatal(args...) }
func (s *settableLogger) Fatalf(format string, args ...interface{}) { s.get().Fatalf(format, args...) }
func (s *settableLogger) Print(args ...interface{})                 { s.get().Print(args...) }
func (s *settableLogger) Printf(format string, args ...interface{}) { s.get().Printf(format, args...) }
