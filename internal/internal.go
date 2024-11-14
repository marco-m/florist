// Package internal contains useful functions.
package internal

import (
	"fmt"
	"log/slog"
)

func MakeErrorfAndLog(prefix string, log *slog.Logger,
) (func(format string, a ...any) error, *slog.Logger) {
	return MakeErrorf(prefix), MakeLog(prefix, log)
}

func MakeErrorf(prefix string) func(format string, a ...any) error {
	return func(format string, a ...any) error {
		return fmt.Errorf(prefix+": "+format, a...)
	}
}

func MakeLog(prefix string, log *slog.Logger) *slog.Logger {
	return log.With(slog.String("fn", prefix))
}
