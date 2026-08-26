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
	"fmt"
	"strconv"
	"testing"

	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

type benchmarkFactory func(automerge.ActorID) (*automerge.Document, error)

func BenchmarkDocumentCreation(b *testing.B) {
	benchmarkEngines(
		b,
		func(b *testing.B, factory benchmarkFactory) {
			b.ReportAllocs()

			for b.Loop() {
				document, err := factory(actor(200))
				if err != nil {
					b.Fatal(err)
				}

				if err := document.Close(); err != nil {
					b.Fatal(err)
				}
			}
		},
	)
}

func BenchmarkMapMutations(b *testing.B) {
	for _, size := range []int{100, 1_000} {
		b.Run(
			fmt.Sprintf("size=%d", size),
			func(b *testing.B) {
				benchmarkEngines(
					b,
					func(b *testing.B, factory benchmarkFactory) {
						b.ReportAllocs()

						for b.Loop() {
							document, err := factory(actor(201))
							if err != nil {
								b.Fatal(err)
							}

							values, err := document.Root().CreateObject(

								"values",
								automerge.ObjectTypeMap,
							)
							if err != nil {
								b.Fatal(err)
							}

							for index := range size {
								if err := values.PutScalar(

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

								"map mutations",
								commitTime,
							); err != nil {
								b.Fatal(err)
							}

							if err := document.Close(); err != nil {
								b.Fatal(err)
							}
						}
					},
				)
			},
		)
	}
}

func BenchmarkTextTyping(b *testing.B) {
	for _, size := range []int{100, 1_000} {
		b.Run(
			fmt.Sprintf("size=%d", size),
			func(b *testing.B) {
				benchmarkEngines(
					b,
					func(b *testing.B, factory benchmarkFactory) {
						b.ReportAllocs()

						for b.Loop() {
							document, err := factory(actor(202))
							if err != nil {
								b.Fatal(err)
							}

							text, err := document.CreateText("body")
							if err != nil {
								b.Fatal(err)
							}

							for index := range size {
								if err := text.Splice(

									uint32(index),
									0,
									"x",
								); err != nil {
									b.Fatal(err)
								}
							}

							if _, err := document.Commit(

								"text typing",
								commitTime,
							); err != nil {
								b.Fatal(err)
							}

							if err := document.Close(); err != nil {
								b.Fatal(err)
							}
						}
					},
				)
			},
		)
	}
}

func BenchmarkLoad(b *testing.B) {
	data := benchmarkDocument(b, 10_000)

	b.Run(
		"native",
		func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				document, err := automerge.Load(data, actor(203))
				if err != nil {
					b.Fatal(err)
				}

				if err := document.Close(); err != nil {
					b.Fatal(err)
				}
			}
		},
	)
	b.Run(
		"reference",
		func(b *testing.B) {
			warmReference(b)
			b.ReportAllocs()

			for b.Loop() {
				document, err := automerge.LoadReference(data, actor(203))
				if err != nil {
					b.Fatal(err)
				}

				if err := document.Close(); err != nil {
					b.Fatal(err)
				}
			}
		},
	)
}

func BenchmarkSave(b *testing.B) {
	data := benchmarkDocument(b, 10_000)

	b.Run(
		"native",
		func(b *testing.B) {
			document, err := automerge.Load(data, actor(204))
			if err != nil {
				b.Fatal(err)
			}

			defer func() { _ = document.Close() }()

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if _, err := document.Save(); err != nil {
					b.Fatal(err)
				}
			}
		},
	)
	b.Run(
		"reference",
		func(b *testing.B) {
			warmReference(b)

			document, err := automerge.LoadReference(data, actor(204))
			if err != nil {
				b.Fatal(err)
			}

			defer func() { _ = document.Close() }()

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if _, err := document.Save(); err != nil {
					b.Fatal(err)
				}
			}
		},
	)
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
		b.Run(
			combination.name,
			func(b *testing.B) {
				warmReference(b)
				b.ReportAllocs()

				for b.Loop() {
					source, err := combination.source(actor(205))
					if err != nil {
						b.Fatal(err)
					}

					text, err := source.CreateText("body")
					if err != nil {
						b.Fatal(err)
					}

					if err := text.Splice(0, 0, benchmarkText(1_000)); err != nil {
						b.Fatal(err)
					}

					if _, err := source.Commit("sync source", commitTime); err != nil {
						b.Fatal(err)
					}

					target, err := combination.target(actor(206))
					if err != nil {
						b.Fatal(err)
					}

					sourceState, err := source.NewSyncState()
					if err != nil {
						b.Fatal(err)
					}

					targetState, err := target.NewSyncState()
					if err != nil {
						b.Fatal(err)
					}

					if err := benchmarkSynchronize(
						sourceState,
						targetState,
					); err != nil {
						b.Fatal(err)
					}

					_ = sourceState.Close()
					_ = targetState.Close()
					_ = source.Close()
					_ = target.Close()
				}
			},
		)
	}
}

func benchmarkEngines(
	b *testing.B,
	benchmark func(*testing.B, benchmarkFactory),
) {
	b.Helper()

	b.Run(
		"native",
		func(b *testing.B) {
			benchmark(b, automerge.New)
		},
	)
	b.Run(
		"reference",
		func(b *testing.B) {
			warmReference(b)
			benchmark(b, automerge.NewReference)
		},
	)
}

func benchmarkDocument(b *testing.B, size int) []byte {
	b.Helper()

	document, err := automerge.New(actor(207))
	if err != nil {
		b.Fatal(err)
	}

	defer func() { _ = document.Close() }()

	text, err := document.CreateText("body")
	if err != nil {
		b.Fatal(err)
	}

	if err := text.Splice(0, 0, benchmarkText(size)); err != nil {
		b.Fatal(err)
	}

	if _, err := document.Commit("benchmark fixture", commitTime); err != nil {
		b.Fatal(err)
	}

	data, err := document.Save()
	if err != nil {
		b.Fatal(err)
	}

	return data
}

func warmReference(b *testing.B) {
	b.Helper()
	b.StopTimer()

	document, err := automerge.NewReference(actor(208))
	if err != nil {
		b.Fatal(err)
	}

	if err := document.Close(); err != nil {
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
	left *automerge.SyncState,
	right *automerge.SyncState,
) error {
	for range 100 {
		progressed := false

		message, ok, err := left.GenerateMessage()
		if err != nil {
			return err
		}

		if ok {
			if err := right.ReceiveMessage(message); err != nil {
				return err
			}

			progressed = true
		}

		message, ok, err = right.GenerateMessage()
		if err != nil {
			return err
		}

		if ok {
			if err := left.ReceiveMessage(message); err != nil {
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
