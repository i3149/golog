// logger.go: extends the log package so that each log message is
// prefixed with a string
//
// Copyright (c) 2011-2012 CloudFlare, Inc.

package logger

import (
	"sync/atomic"
)

type Level int

var (
	// offLogger is a dummy no-op logger.
	OffLogger = New(Levels.Off)

	// Levels is a singleton that represents possible log levels.
	Levels = struct {
		Off    Level
		Panic  Level
		Error  Level
		Warn   Level
		Info   Level
		Debug  Level
		Access Level
	}{
		Access: (-1),
		Off:    (0),
		Panic:  (1),
		Error:  (2),
		Warn:   (3),
		Info:   (4),
		Debug:  (5),
	}

	// levelMap maps Level objects to the pretty printed name
	levelMap = map[Level]string{
		Levels.Access: "Access",
		Levels.Off:    "Off",
		Levels.Panic:  "Panic",
		Levels.Error:  "Error",
		Levels.Warn:   "Warn",
		Levels.Info:   "Info",
		Levels.Debug:  "Debug",
	}

	// CfgLevels maps strings to Level. The intent is to use this during config
	// time.
	CfgLevels = map[string]Level{
		"access": Levels.Access,
		"off":    Levels.Off,
		"panic":  Levels.Panic,
		"error":  Levels.Error,
		"warn":   Levels.Warn,
		"info":   Levels.Info,
		"debug":  Levels.Debug,
	}

	logCount  uint64 // number of messages attemped on all loggers
	dropCount uint64 // number of messages dropped on all loggers
	errCount  uint64 // number of errors seen across all loggers
)

/* Stats returns the current status of the logger. It reports:
 * logs: number of logs attempted to be written since startup
 * pending: number of logs queued to be written
 * drop: numer of logs that have been dropped, because the write queue is full, since startup
 * errs: number of errors seen while trying to write logs since startup
 */
func Stats() (logs, pending, drop, errs uint64) {
	return atomic.LoadUint64(&logCount), uint64(len(messages)), atomic.LoadUint64(&dropCount), atomic.LoadUint64(&errCount)
}

type Logger struct {
	level               Level
	sample, sampleCount uint64 // counters to allow us to sample every "sample" access logs
}

func (level Level) String() string {
	return levelMap[level]
}

func New(level Level) (l *Logger) {
	l = new(Logger)
	l.level = level
	l.sample = 1

	return
}

func (l *Logger) Printf(level Level, prefix, format string, v ...interface{}) {
	switch {
	case level == Levels.Access:
		count := atomic.AddUint64(&l.sampleCount, 1)
		if l.sample == 0 || count%l.sample != 0 {
			return
		}
	case level > l.level, level == Levels.Off:
		return
	}

	queueMsg(level, prefix, format, v...)
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) Level() Level {
	return l.level
}

func (l *Logger) SetAccessLogSample(sample uint64) {
	atomic.StoreUint64(&l.sample, sample)
}
