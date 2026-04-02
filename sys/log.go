// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

func logSetup() error {
	var w io.Writer = io.Discard
	var level slog.Level

	if Options.LogLevel > 0 {
		w = os.Stderr
		if Options.LogFile != "" {
			f, err := os.OpenFile(Options.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			w = f
		}

		switch Options.LogLevel {
		case 1:
			level = slog.LevelError
		case 2:
			level = slog.LevelWarn
		case 3:
			level = slog.LevelInfo
		default:
			level = slog.LevelDebug
		}
	}

	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   level <= slog.LevelDebug,
		ReplaceAttr: logFormat,
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(w, opts)))
	return nil
}

func logFormat(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		fn := source.Function

		if lastSlash := strings.LastIndex(fn, "/"); lastSlash != -1 {
			if secondLastSlash := strings.LastIndex(fn[:lastSlash], "/"); secondLastSlash != -1 {
				fn = fn[secondLastSlash+1:]
			}
		}

		if start := strings.Index(fn, "("); start != -1 {
			if end := strings.Index(fn, ")."); end != -1 {
				fn = fn[:start] + fn[end+2:]
			}
		}

		return slog.String("fn", fn)
	}
	return a
}
