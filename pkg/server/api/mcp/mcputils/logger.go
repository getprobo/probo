// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package mcputils

import (
	"context"
	"log/slog"
	"slices"

	"go.gearno.de/kit/log"
)

type (
	slogHandler struct {
		logger *log.Logger
		attrs  []slog.Attr
		groups []slogGroup
	}

	slogGroup struct {
		name  string
		attrs []slog.Attr
	}
)

// NewSlogLogger adapts the application logger for dependencies that use slog.
func NewSlogLogger(logger *log.Logger) *slog.Logger {
	return slog.New(&slogHandler{logger: logger})
}

func (h *slogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	attrs := slices.Clone(h.attrs)
	groupedAttrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		groupedAttrs = append(groupedAttrs, attr)
		return true
	})

	for i := len(h.groups) - 1; i >= 0; i-- {
		groupAttrs := append(slices.Clone(h.groups[i].attrs), groupedAttrs...)
		groupedAttrs = []slog.Attr{
			slog.Group(h.groups[i].name, attrsToAny(groupAttrs)...),
		}
	}

	h.logger.Log(ctx, record.Level, record.Message, append(attrs, groupedAttrs...)...)

	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	child := h.clone()
	if len(child.groups) == 0 {
		child.attrs = append(child.attrs, attrs...)
	} else {
		i := len(child.groups) - 1
		child.groups[i].attrs = append(child.groups[i].attrs, attrs...)
	}

	return child
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	child := h.clone()
	child.groups = append(child.groups, slogGroup{name: name})

	return child
}

func (h *slogHandler) clone() *slogHandler {
	child := &slogHandler{
		logger: h.logger,
		attrs:  slices.Clone(h.attrs),
		groups: slices.Clone(h.groups),
	}
	for i := range child.groups {
		child.groups[i].attrs = slices.Clone(child.groups[i].attrs)
	}

	return child
}

func attrsToAny(attrs []slog.Attr) []any {
	values := make([]any, len(attrs))
	for i, attr := range attrs {
		values[i] = attr
	}

	return values
}
