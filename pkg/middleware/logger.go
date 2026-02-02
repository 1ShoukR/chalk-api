package middleware

import (
	"io"
	"log/slog"
)

func SetupLogger(w io.Writer) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
	l := slog.New(h)
	l = l.With("app", slog.GroupValue(
		slog.String("name", "Chalk-API"),
	))
	return l
}
