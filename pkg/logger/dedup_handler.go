package logger

import (
	"context"
	"log/slog"
)

type dedupHandler struct {
	next  slog.Handler
	attrs []slog.Attr
}

func newDedupHandler(next slog.Handler) slog.Handler {
	return &dedupHandler{next: next}
}

func (h *dedupHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *dedupHandler) Handle(ctx context.Context, r slog.Record) error { //nolint:gocritic // interface implementation
	rec := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	merged := mergeAttrs(h.attrs, recordAttrs(&r))
	for _, a := range merged {
		rec.AddAttrs(a)
	}

	return h.next.Handle(ctx, rec)
}

func (h *dedupHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &dedupHandler{
		next:  h.next,
		attrs: mergeAttrs(h.attrs, attrs),
	}
}

func (h *dedupHandler) WithGroup(name string) slog.Handler {
	return &dedupHandler{
		next:  h.next.WithGroup(name),
		attrs: h.attrs,
	}
}

func recordAttrs(r *slog.Record) []slog.Attr {
	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	return attrs
}

func mergeAttrs(base, extra []slog.Attr) []slog.Attr {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}

	out := make([]slog.Attr, 0, len(base)+len(extra))
	index := make(map[string]int, len(base)+len(extra))

	appendOrReplace := func(a slog.Attr) {
		if i, ok := index[a.Key]; ok {
			out[i] = a
			return
		}
		index[a.Key] = len(out)
		out = append(out, a)
	}

	for _, a := range base {
		appendOrReplace(a)
	}
	for _, a := range extra {
		appendOrReplace(a)
	}

	return out
}
