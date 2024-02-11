package slogparquet

import (
	"context"

	"log/slog"

	slogcommon "github.com/samber/slog-common"
)

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// parquet rows buffer
	Buffer ParquetBuffer

	// optional: customize json payload builder
	Converter Converter

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

func (o Option) NewParquetHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	if o.Buffer == nil {
		panic("missing buffer configuration")
	}

	if o.Converter == nil {
		o.Converter = DefaultConverter
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
	attrs := h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, h.attrs, h.groups, &record)

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
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
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
