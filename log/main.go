// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>
//
// DISCLAIMER: tibula is now using slog, this package is only for backward compatibility

package log

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
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

var Level = LevelInfo
var logStderr = true
var logLevelChecked = false

func Init(level int, filename string) error {
	if level < LevelFatal || level > LevelTrace {
		return fmt.Errorf("invalid log level")
	}
	Level = level
	logLevelChecked = true

	if filename != "" {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		var sLevel slog.Level
		switch level {
		case LevelError:
			sLevel = slog.LevelError
		case LevelWarn:
			sLevel = slog.LevelWarn
		case LevelInfo:
			sLevel = slog.LevelInfo
		default:
			sLevel = slog.LevelDebug
		}

		opts := &slog.HandlerOptions{Level: sLevel}
		slog.SetDefault(slog.New(slog.NewJSONHandler(file, opts)))
		logStderr = false
	}

	return nil
}

func Log(level int, args ...any) {
	if !logLevelChecked {
		if f := flag.Lookup("log-level"); f != nil {
			if v, err := strconv.Atoi(f.Value.String()); err == nil {
				Level = v
			}
		} else {
			ctx := context.Background()
			logger := slog.Default()

			if logger.Enabled(ctx, slog.LevelDebug) {
				Level = LevelDebug
			} else if logger.Enabled(ctx, slog.LevelInfo) {
				Level = LevelInfo
			} else if logger.Enabled(ctx, slog.LevelWarn) {
				Level = LevelWarn
			} else if logger.Enabled(ctx, slog.LevelError) {
				Level = LevelError
			} else {
				Level = LevelFatal
			}
		}
		logLevelChecked = true
	}

	if level > Level {
		return
	}

	msg := ""
	for _, arg := range args {
		if str, ok := arg.(string); ok {
			arg = regexp.MustCompile(`[\n\t\s]+`).ReplaceAllString(str, " ")
			arg = strings.Trim(arg.(string), " ")
		}
		msg += fmt.Sprintf(" %v", arg)
	}
	msg = strings.TrimSpace(msg)

	legacyAttr := slog.Bool("legacy_logger", true)

	switch level {
	case LevelFatal:
		slog.Error(msg, slog.Bool("fatal", true), legacyAttr)
		os.Exit(1)
	case LevelError:
		slog.Error(msg, legacyAttr)
	case LevelWarn:
		slog.Warn(msg, legacyAttr)
	case LevelInfo:
		slog.Info(msg, legacyAttr)
	case LevelDebug:
		slog.Debug(msg, legacyAttr)
	case LevelTrace:
		slog.LogAttrs(context.Background(), slog.LevelDebug, msg, slog.Bool("trace", true), legacyAttr)
	}
}

func Fatal(args ...any) {
	Log(LevelFatal, args...)
}

func Error(args ...any) {
	Log(LevelError, args...)
}

func Warn(args ...any) {
	Log(LevelWarn, args...)
}

func Info(args ...any) {
	Log(LevelInfo, args...)
}

func Debug(args ...any) {
	Log(LevelDebug, args...)
}

func Trace(args ...any) {
	Log(LevelTrace, args...)
}
