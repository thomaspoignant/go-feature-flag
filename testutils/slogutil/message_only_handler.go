package slogutil

import (
	"context"
	"log/slog"
	"os"
)

type MessageOnlyHandler struct {
	Writer *os.File
}

// Handle implements the slog.Handler interface.
func (h *MessageOnlyHandler) Handle(_ context.Context, r slog.Record) error {
	_, err := h.Writer.WriteString(r.Message + "\n")
	return err
}

// Enabled implements the slog.Handler interface.
func (h *MessageOnlyHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// WithAttrs implements the slog.Handler interface.
func (h *MessageOnlyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup implements the slog.Handler interface.
func (h *MessageOnlyHandler) WithGroup(_ string) slog.Handler {
	return h
}
