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

package prosemirror_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.probo.inc/probo/pkg/prosemirror"
)

func BenchmarkParse(b *testing.B) {
	for _, fixture := range benchFixtures(b) {
		b.Run(
			fixture.name,
			func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(fixture.raw)))

				for b.Loop() {
					if _, err := prosemirror.Parse(fixture.raw); err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}

func BenchmarkValidate(b *testing.B) {
	for _, fixture := range benchFixtures(b) {
		b.Run(
			fixture.name,
			func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(fixture.raw)))

				for b.Loop() {
					if err := prosemirror.ValidateDocumentContentJSON(fixture.raw); err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}

func BenchmarkSanitize(b *testing.B) {
	for _, fixture := range benchFixtures(b) {
		b.Run(
			fixture.name,
			func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(fixture.raw)))

				for b.Loop() {
					if _, err := prosemirror.SanitizeDocumentJSON(fixture.raw); err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}

func BenchmarkRenderHTML(b *testing.B) {
	for _, fixture := range benchFixtures(b) {
		node, err := prosemirror.Parse(fixture.raw)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(
			fixture.name,
			func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(fixture.raw)))

				for b.Loop() {
					if _, err := prosemirror.RenderHTML(node); err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}

func BenchmarkRenderMarkdown(b *testing.B) {
	for _, fixture := range benchFixtures(b) {
		node, err := prosemirror.Parse(fixture.raw)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(
			fixture.name,
			func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(fixture.raw)))

				for b.Loop() {
					if _, err := prosemirror.RenderMarkdown(node); err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}

func BenchmarkTextLength(b *testing.B) {
	for _, fixture := range benchFixtures(b) {
		node, err := prosemirror.Parse(fixture.raw)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(
			fixture.name,
			func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(fixture.raw)))

				for b.Loop() {
					_ = node.TextLength()
				}
			},
		)
	}
}

type benchFixture struct {
	name string
	raw  string
}

func benchFixtures(tb testing.TB) []benchFixture {
	tb.Helper()

	names := []string{"bench_small.json", "bench_large.json", "bench_mentions.json"}
	out := make([]benchFixture, 0, len(names))

	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			tb.Fatal(err)
		}

		out = append(
			out,
			benchFixture{
				name: name[:len(name)-len(filepath.Ext(name))],
				raw:  string(raw),
			},
		)
	}

	return out
}
