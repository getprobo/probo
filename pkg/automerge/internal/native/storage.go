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

package native

import (
	internalencoding "go.probo.inc/probo/pkg/automerge/internal/encoding"
	internalstorage "go.probo.inc/probo/pkg/automerge/internal/storage"
)

var (
	Decode            = internalstorage.Decode
	DecodePartial     = internalstorage.DecodePartial
	DecodeIncremental = internalstorage.DecodeIncremental
	EncodeChange      = internalstorage.EncodeChange
)

func deflate(data []byte) ([]byte, error) { return internalstorage.Deflate(data) }

// Small compatibility wrapper while higher-level engine code migrates to the
// shared encoding package.
type reader struct{ inner *internalencoding.Reader }

func newReader(data []byte) *reader { return &reader{inner: internalencoding.NewReader(data)} }
func newReaderAt(data []byte, offset int) *reader {
	return &reader{inner: internalencoding.NewReaderAt(data, offset)}
}
func (r *reader) remaining() int                      { return r.inner.Remaining() }
func (r *reader) offset() int                         { return r.inner.Offset() }
func (r *reader) byte() (byte, error)                 { return r.inner.Byte() }
func (r *reader) bytes(length uint64) ([]byte, error) { return r.inner.Bytes(length) }
func (r *reader) uleb() (uint64, error)               { return r.inner.ULEB() }
func appendULEB(data []byte, value uint64) []byte     { return internalencoding.AppendULEB(data, value) }
func appendLengthPrefixedNative(data, value []byte) []byte {
	return internalencoding.AppendLengthPrefixed(data, value)
}
func decodeLengthPrefixed(r *reader) ([]byte, error) {
	return internalencoding.DecodeLengthPrefixed(r.inner)
}
