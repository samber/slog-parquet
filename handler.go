package slogparquet

import (
	"context"

	"log/slog"
)

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// parquet rows buffer
	Buffer ParquetBuffer

	// optional: customize json payload builder
	Converter Converter
}

func (o Option) NewParquetHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	if o.Buffer == nil {
		panic("missing buffer configuration")
	}

	return &ParquetHandler{
		option: o,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*ParquetHandler)(nil)

type ParquetHandler struct {
	option Option
	attrs  []slog.Attr
	groups []string
}

func (h *ParquetHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *ParquetHandler) Handle(ctx context.Context, record slog.Record) error {
	converter := DefaultConverter
	if h.option.Converter != nil {
		converter = h.option.Converter
	}

	attrs := converter(h.attrs, &record)

	return h.option.Buffer.Append(
		record.Time,
		record.Level,
		record.Message,
		attrs,
	)
}

func (h *ParquetHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ParquetHandler{
		option: h.option,
		attrs:  appendAttrsToGroup(h.groups, h.attrs, attrs),
		groups: h.groups,
	}
}

func (h *ParquetHandler) WithGroup(name string) slog.Handler {
	return &ParquetHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}
