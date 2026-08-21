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

package automerge_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"go.probo.inc/probo/pkg/automerge"
)

type benchmarkFactory func(
	context.Context,
	automerge.ActorID,
) (*automerge.Document, error)

func BenchmarkDocumentCreation(b *testing.B) {
	benchmarkEngines(b, func(b *testing.B, factory benchmarkFactory) {
		ctx := context.Background()

		b.ReportAllocs()

		for b.Loop() {
			document, err := factory(ctx, actor(200))
			if err != nil {
				b.Fatal(err)
			}

			if err := document.Close(ctx); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkMapMutations(b *testing.B) {
	for _, size := range []int{100, 1_000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			benchmarkEngines(b, func(b *testing.B, factory benchmarkFactory) {
				ctx := context.Background()

				b.ReportAllocs()

				for b.Loop() {
					document, err := factory(ctx, actor(201))
					if err != nil {
						b.Fatal(err)
					}

					values, err := document.Root().CreateObject(
						ctx,
						"values",
						automerge.ObjectTypeMap,
					)
					if err != nil {
						b.Fatal(err)
					}

					for index := range size {
						if err := values.PutScalar(
							ctx,
							strconv.Itoa(index),
							automerge.Scalar{
								Type: automerge.ScalarTypeInt,
								Int:  int64(index),
							},
						); err != nil {
							b.Fatal(err)
						}
					}

					if _, err := document.Commit(
						ctx,
						"map mutations",
						commitTime,
					); err != nil {
						b.Fatal(err)
					}

					if err := document.Close(ctx); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkTextTyping(b *testing.B) {
	for _, size := range []int{100, 1_000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			benchmarkEngines(b, func(b *testing.B, factory benchmarkFactory) {
				ctx := context.Background()

				b.ReportAllocs()

				for b.Loop() {
					document, err := factory(ctx, actor(202))
					if err != nil {
						b.Fatal(err)
					}

					text, err := document.CreateText(ctx, "body")
					if err != nil {
						b.Fatal(err)
					}

					for index := range size {
						if err := text.Splice(
							ctx,
							uint32(index),
							0,
							"x",
						); err != nil {
							b.Fatal(err)
						}
					}

					if _, err := document.Commit(
						ctx,
						"text typing",
						commitTime,
					); err != nil {
						b.Fatal(err)
					}

					if err := document.Close(ctx); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkLoad(b *testing.B) {
	ctx := context.Background()
	data := benchmarkDocument(b, 10_000)

	b.Run("native", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			document, err := automerge.Load(ctx, data, actor(203))
			if err != nil {
				b.Fatal(err)
			}

			if err := document.Close(ctx); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("reference", func(b *testing.B) {
		warmReference(b)
		b.ReportAllocs()

		for b.Loop() {
			document, err := automerge.LoadReference(ctx, data, actor(203))
			if err != nil {
				b.Fatal(err)
			}

			if err := document.Close(ctx); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkSave(b *testing.B) {
	ctx := context.Background()
	data := benchmarkDocument(b, 10_000)

	b.Run("native", func(b *testing.B) {
		document, err := automerge.Load(ctx, data, actor(204))
		if err != nil {
			b.Fatal(err)
		}

		defer func() { _ = document.Close(ctx) }()

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			if _, err := document.Save(ctx); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("reference", func(b *testing.B) {
		warmReference(b)

		document, err := automerge.LoadReference(ctx, data, actor(204))
		if err != nil {
			b.Fatal(err)
		}

		defer func() { _ = document.Close(ctx) }()

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			if _, err := document.Save(ctx); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkInitialSync(b *testing.B) {
	combinations := []struct {
		name   string
		source benchmarkFactory
		target benchmarkFactory
	}{
		{
			name:   "native-to-native",
			source: automerge.New,
			target: automerge.New,
		},
		{
			name:   "native-to-reference",
			source: automerge.New,
			target: automerge.NewReference,
		},
		{
			name:   "reference-to-native",
			source: automerge.NewReference,
			target: automerge.New,
		},
	}

	for _, combination := range combinations {
		b.Run(combination.name, func(b *testing.B) {
			ctx := context.Background()

			warmReference(b)
			b.ReportAllocs()

			for b.Loop() {
				source, err := combination.source(ctx, actor(205))
				if err != nil {
					b.Fatal(err)
				}

				text, err := source.CreateText(ctx, "body")
				if err != nil {
					b.Fatal(err)
				}

				if err := text.Splice(ctx, 0, 0, benchmarkText(1_000)); err != nil {
					b.Fatal(err)
				}

				if _, err := source.Commit(ctx, "sync source", commitTime); err != nil {
					b.Fatal(err)
				}

				target, err := combination.target(ctx, actor(206))
				if err != nil {
					b.Fatal(err)
				}

				sourceState, err := source.NewSyncState(ctx)
				if err != nil {
					b.Fatal(err)
				}

				targetState, err := target.NewSyncState(ctx)
				if err != nil {
					b.Fatal(err)
				}

				if err := benchmarkSynchronize(
					ctx,
					sourceState,
					targetState,
				); err != nil {
					b.Fatal(err)
				}

				_ = sourceState.Close(ctx)
				_ = targetState.Close(ctx)
				_ = source.Close(ctx)
				_ = target.Close(ctx)
			}
		})
	}
}

func benchmarkEngines(
	b *testing.B,
	benchmark func(*testing.B, benchmarkFactory),
) {
	b.Helper()

	b.Run("native", func(b *testing.B) {
		benchmark(b, automerge.New)
	})
	b.Run("reference", func(b *testing.B) {
		warmReference(b)
		benchmark(b, automerge.NewReference)
	})
}

func benchmarkDocument(b *testing.B, size int) []byte {
	b.Helper()

	ctx := context.Background()

	document, err := automerge.New(ctx, actor(207))
	if err != nil {
		b.Fatal(err)
	}

	defer func() { _ = document.Close(ctx) }()

	text, err := document.CreateText(ctx, "body")
	if err != nil {
		b.Fatal(err)
	}

	if err := text.Splice(ctx, 0, 0, benchmarkText(size)); err != nil {
		b.Fatal(err)
	}

	if _, err := document.Commit(ctx, "benchmark fixture", commitTime); err != nil {
		b.Fatal(err)
	}

	data, err := document.Save(ctx)
	if err != nil {
		b.Fatal(err)
	}

	return data
}

func warmReference(b *testing.B) {
	b.Helper()
	b.StopTimer()

	ctx := context.Background()

	document, err := automerge.NewReference(ctx, actor(208))
	if err != nil {
		b.Fatal(err)
	}

	if err := document.Close(ctx); err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
}

func benchmarkText(size int) string {
	value := make([]byte, size)
	for index := range value {
		value[index] = byte('a' + index%26)
	}

	return string(value)
}

func benchmarkSynchronize(
	ctx context.Context,
	left *automerge.SyncState,
	right *automerge.SyncState,
) error {
	for range 100 {
		progressed := false

		message, ok, err := left.GenerateMessage(ctx)
		if err != nil {
			return err
		}

		if ok {
			if err := right.ReceiveMessage(ctx, message); err != nil {
				return err
			}

			progressed = true
		}

		message, ok, err = right.GenerateMessage(ctx)
		if err != nil {
			return err
		}

		if ok {
			if err := left.ReceiveMessage(ctx, message); err != nil {
				return err
			}

			progressed = true
		}

		if !progressed {
			return nil
		}
	}

	return fmt.Errorf("sync did not quiesce")
}
