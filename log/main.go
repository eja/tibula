// Package log provides a simple logging mechanism with different log levels.
//
// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package log

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	LevelFatal = 0
	LevelError = 1
	LevelWarn  = 2
	LevelInfo  = 3
	LevelDebug = 4
	LevelTrace = 5
)

var logLevel = LevelInfo
var logStderr = true

func Init(level int, filename string) error {
	if level < LevelFatal || level > LevelTrace {
		return fmt.Errorf("invalid log level")
	}
	logLevel = level
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	if filename != "" {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			return err
		}
		log.SetOutput(file)
		logStderr = false
	}

	return nil
}

func Log(level int, args ...interface{}) {
	msg := ""
	switch level {
	case LevelFatal:
		msg = "[F]"
	case LevelError:
		msg = "[E]"
	case LevelWarn:
		msg = "[W]"
	case LevelInfo:
		msg = "[I]"
	case LevelDebug:
		msg = "[D]"
	case LevelTrace:
		msg = "[T]"
	}

	for _, arg := range args {
		if str, ok := arg.(string); ok {
			arg = regexp.MustCompile(`[\n\t\s]+`).ReplaceAllString(str, " ")
			arg = strings.Trim(arg.(string), " ")
		}
		msg += fmt.Sprintf(" %v", arg)
	}

	if level == LevelFatal {
		if logStderr {
			log.SetFlags(0)
			log.Fatal(args...)
		} else {
			log.Fatal(msg)
		}
	}
	if level <= logLevel && level >= LevelError && level <= LevelTrace {
		log.Println(msg)
	}
}

func Fatal(args ...interface{}) {
	Log(LevelFatal, args...)
}

func Error(args ...interface{}) {
	Log(LevelError, args...)
}

func Warn(args ...interface{}) {
	Log(LevelWarn, args...)
}

func Info(args ...interface{}) {
	Log(LevelInfo, args...)
}

func Debug(args ...interface{}) {
	Log(LevelDebug, args...)
}

func Trace(args ...interface{}) {
	Log(LevelTrace, args...)
}
