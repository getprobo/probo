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
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	productionautomerge "go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
)

// The register fixture mirrors a generated compliance register: every entry is
// a run of headings and short labelled paragraphs. Each block is its own
// object with nested attributes and parents, so the document is dominated by
// small objects around one long text, the shape that single-text fixtures do
// not exercise.
const (
	registerEntryParagraphs = 30
	registerHistoryChanges  = 2_000
)

var (
	registerFixtures     sync.Map
	registerHistoryOnce  sync.Once
	registerHistoryBytes []byte
)

type registerFixture struct {
	once    sync.Once
	content string
	data    []byte
}

func registerActor(seed byte) productionautomerge.ActorID {
	var actor productionautomerge.ActorID

	for index := range actor {
		actor[index] = seed + byte(index)
	}

	return actor
}

func registerContent(entries int) string {
	type node = map[string]any

	text := func(value string, bold bool) node {
		result := node{"type": "text", "text": value}
		if bold {
			result["marks"] = []node{{"type": "bold"}}
		}

		return result
	}
	heading := func(level int, value string) node {
		return node{
			"type":    "heading",
			"attrs":   node{"level": level},
			"content": []node{text(value, false)},
		}
	}

	blocks := []node{heading(1, "1. Register")}

	for entry := range entries {
		blocks = append(blocks, heading(2, fmt.Sprintf("1.%d Entry %d", entry+1, entry+1)))

		for paragraph := range registerEntryParagraphs {
			if paragraph%8 == 0 {
				blocks = append(
					blocks,
					heading(3, fmt.Sprintf("1.%d.%d Section", entry+1, paragraph/8+1)),
				)
			}

			blocks = append(
				blocks,
				node{
					"type": "paragraph",
					"content": []node{
						text(fmt.Sprintf("Field %d: ", paragraph+1), true),
						text(
							fmt.Sprintf(
								"value recorded for entry %d, field %d, kept short",
								entry+1,
								paragraph+1,
							),
							false,
						),
					},
				},
			)
		}
	}

	encoded, err := json.Marshal(node{"type": "doc", "content": blocks})
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func seedRegister(tb testing.TB, content string) *productionautomerge.Document {
	tb.Helper()

	spans, err := automergeprosemirror.ToSpans(content)
	if err != nil {
		tb.Fatal(err)
	}

	document, err := productionautomerge.New(registerActor(1))
	if err != nil {
		tb.Fatal(err)
	}

	text, err := document.CreateText("body")
	if err != nil {
		tb.Fatal(err)
	}

	if err := text.UpdateSpans(spans, automergeprosemirror.UpdateSpansConfig()); err != nil {
		tb.Fatal(err)
	}

	if _, err := document.Commit("seed register", commitTime); err != nil {
		tb.Fatal(err)
	}

	return document
}

func benchmarkRegister(tb testing.TB, entries int) *registerFixture {
	tb.Helper()

	value, _ := registerFixtures.LoadOrStore(entries, &registerFixture{})
	fixture := value.(*registerFixture)

	fixture.once.Do(func() {
		fixture.content = registerContent(entries)

		document := seedRegister(tb, fixture.content)
		defer func() { _ = document.Close() }()

		data, err := document.Save()
		if err != nil {
			tb.Fatal(err)
		}

		fixture.data = data
	})

	return fixture
}

// benchmarkRegisterHistory is one text edited over many small changes, some of
// them carrying marks, so loading it has a real change graph to rebuild.
func benchmarkRegisterHistory(tb testing.TB) []byte {
	tb.Helper()

	registerHistoryOnce.Do(func() {
		document, err := productionautomerge.New(registerActor(2))
		if err != nil {
			tb.Fatal(err)
		}

		defer func() { _ = document.Close() }()

		text, err := document.CreateText("body")
		if err != nil {
			tb.Fatal(err)
		}

		const word = "edited "

		for change := range registerHistoryChanges {
			position := uint32(change * len(word))
			if err := text.Splice(position, 0, word); err != nil {
				tb.Fatal(err)
			}

			if change%10 == 0 {
				if err := text.Mark(
					position,
					position+uint32(len(word))-1,
					"strong",
					productionautomerge.BoolScalar(true),
					productionautomerge.MarkExpandNone,
				); err != nil {
					tb.Fatal(err)
				}
			}

			if _, err := document.Commit(
				"edit",
				commitTime.Add(time.Duration(change)*time.Second),
			); err != nil {
				tb.Fatal(err)
			}
		}

		data, err := document.Save()
		if err != nil {
			tb.Fatal(err)
		}

		registerHistoryBytes = data
	})

	return registerHistoryBytes
}

func BenchmarkLoadRegister(b *testing.B) {
	for _, entries := range []int{10, 40, registerLoadEntries} {
		b.Run(
			fmt.Sprintf("entries=%d", entries),
			func(b *testing.B) {
				fixture := benchmarkRegister(b, entries)

				b.ReportAllocs()

				for b.Loop() {
					document, err := productionautomerge.Load(fixture.data, registerActor(3))
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
}

func BenchmarkLoadHistory(b *testing.B) {
	data := benchmarkRegisterHistory(b)

	b.ReportAllocs()

	for b.Loop() {
		document, err := productionautomerge.Load(data, registerActor(3))
		if err != nil {
			b.Fatal(err)
		}

		if err := document.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSpansRegister(b *testing.B) {
	for _, entries := range []int{10, 40, registerLoadEntries} {
		b.Run(
			fmt.Sprintf("entries=%d", entries),
			func(b *testing.B) {
				fixture := benchmarkRegister(b, entries)

				document, err := productionautomerge.Load(fixture.data, registerActor(3))
				if err != nil {
					b.Fatal(err)
				}

				defer func() { _ = document.Close() }()

				text, err := document.Text("body")
				if err != nil {
					b.Fatal(err)
				}

				b.ReportAllocs()

				for b.Loop() {
					if _, err := text.Spans(); err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}

func BenchmarkSeedRegister(b *testing.B) {
	for _, entries := range []int{10, 40} {
		b.Run(
			fmt.Sprintf("entries=%d", entries),
			func(b *testing.B) {
				fixture := benchmarkRegister(b, entries)

				b.ReportAllocs()

				for b.Loop() {
					document := seedRegister(b, fixture.content)

					if err := document.Close(); err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}
