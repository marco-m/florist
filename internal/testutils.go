package internal

import "log/slog"

// RemoveTime removes the "time" attribute from the output of a slog.Logger.
func RemoveTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

func MakeTestLog() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
