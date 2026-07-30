// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

// Package safecsv wraps encoding/csv for exports that may contain user-controlled
// text. Writer sanitizes every cell before encoding to reduce spreadsheet formula
// injection when a file is opened in Excel, LibreOffice Calc, Google Sheets, or
// similar tools.
//
// Use safecsv for all user-facing CSV downloads. Do not write export CSV with
// encoding/csv alone when record fields can contain untrusted or attacker-influenced
// content.
package safecsv

import (
	"encoding/csv"
	"io"
)

type Writer struct {
	inner *csv.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{inner: csv.NewWriter(w)}
}

// Write encodes a CSV record after sanitizing every field. Prefer WriteRow for
// literal column values at call sites.
func (w *Writer) Write(record []string) error {
	return w.inner.Write(SanitizeRecord(record))
}

// WriteRow encodes one CSV record from fields, sanitizing each value before
// encoding (spreadsheet formula injection mitigation).
func (w *Writer) WriteRow(fields ...string) error {
	return w.inner.Write(SanitizeRecord(fields))
}

func (w *Writer) Flush() {
	w.inner.Flush()
}

func (w *Writer) Error() error {
	return w.inner.Error()
}
