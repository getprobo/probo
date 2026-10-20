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
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

const chaosPeers = 3

type syncChaos struct {
	documents [chaosPeers]*automerge.Document
	states    [chaosPeers][chaosPeers]*automerge.SyncState
	last      [chaosPeers][chaosPeers][]byte
}

// TestSyncState_ModelBasedChaos combines operations that used to be tested only
// in isolation: concurrent edits, lost and duplicated messages, read-only mode,
// repeated generation without a reply, document persistence, and serialized
// sync-state restoration. Every generated batch must quiesce and a final reliable
// full-mesh delivery must converge all peers.
func TestSyncState_ModelBasedChaos(t *testing.T) {
	t.Parallel()

	const (
		scenarios = 20
		steps     = 120
	)

	for scenario := range scenarios {
		t.Run(
			fmt.Sprintf("seed-%d", scenario),
			func(t *testing.T) {
				t.Parallel()

				random := rand.New(rand.NewSource(int64(0x51C00000 + scenario)))
				chaos := newSyncChaos(t)
				t.Cleanup(func() { chaos.close() })

				for step := range steps {
					switch random.Intn(8) {
					case 0:
						chaos.mapEdit(t, random.Intn(chaosPeers), scenario, step)
					case 1:
						chaos.textEdit(t, random, random.Intn(chaosPeers), step)
					case 2:
						chaos.send(t, random, true)
					case 3:
						chaos.send(t, random, false)
					case 4:
						chaos.duplicate(t, random)
					case 5:
						chaos.toggleReadOnly(t, random)
					case 6:
						chaos.reload(t, random.Intn(chaosPeers))
					case 7:
						chaos.assertGenerationQuiesces(t, random)
					}
				}

				chaos.converge(t)

				expected := chaosSignature(t, chaos.documents[0])
				for peer := 1; peer < chaosPeers; peer++ {
					assert.Equalf(
						t,
						expected,
						chaosSignature(t, chaos.documents[peer]),
						"peer %d did not converge",
						peer,
					)
				}
			},
		)
	}
}

func newSyncChaos(t *testing.T) *syncChaos {
	t.Helper()

	chaos := &syncChaos{}

	seed, err := automerge.New(actor(0x40))
	require.NoError(t, err)

	body, err := seed.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, body.Splice(0, 0, "seed"))

	_, err = seed.Commit("seed", commitTime)
	require.NoError(t, err)

	saved, err := seed.Save()
	require.NoError(t, err)
	require.NoError(t, seed.Close())

	for peer := range chaosPeers {
		chaos.documents[peer], err = automerge.Load(saved, actor(byte(0x50+peer)))
		require.NoError(t, err)
	}

	for source := range chaosPeers {
		for target := range chaosPeers {
			if source == target {
				continue
			}

			chaos.states[source][target], err = chaos.documents[source].NewSyncState()
			require.NoError(t, err)
		}
	}

	return chaos
}

func (c *syncChaos) mapEdit(
	t *testing.T,
	peer int,
	scenario int,
	step int,
) {
	t.Helper()

	key := fmt.Sprintf("p%d-s%d-%d", peer, scenario, step)
	require.NoError(
		t,
		c.documents[peer].Root().PutScalar(

			key,
			automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(step)},
		),
	)
	_, err := c.documents[peer].Commit(

		"map edit",
		commitTime.Add(time.Duration(step)*time.Second),
	)
	require.NoError(t, err)
}

func (c *syncChaos) textEdit(
	t *testing.T,
	random *rand.Rand,
	peer int,
	step int,
) {
	t.Helper()

	text, err := c.documents[peer].Text("body")
	require.NoError(t, err)

	value, err := text.String()
	require.NoError(t, err)

	index := random.Intn(len(value) + 1)
	require.NoError(t, text.Splice(uint32(index), 0, string(rune('a'+peer))))

	_, err = c.documents[peer].Commit(

		"text edit",
		commitTime.Add(time.Duration(step)*time.Second),
	)
	require.NoError(t, err)
}

func (c *syncChaos) send(
	t *testing.T,
	random *rand.Rand,
	deliver bool,
) {
	t.Helper()

	source, target := randomPair(random)
	message, ok, err := c.states[source][target].GenerateMessage()
	require.NoError(t, err)

	if !ok {
		return
	}

	c.last[source][target] = append(c.last[source][target][:0], message...)

	if deliver {
		require.NoError(t, c.states[target][source].ReceiveMessage(message))
	}
}

func (c *syncChaos) duplicate(t *testing.T, random *rand.Rand) {
	t.Helper()

	source, target := randomPair(random)

	message := c.last[source][target]
	if len(message) == 0 {
		return
	}

	require.NoError(t, c.states[target][source].ReceiveMessage(message))
	require.NoError(t, c.states[target][source].ReceiveMessage(message))
}

func (c *syncChaos) toggleReadOnly(
	t *testing.T,
	random *rand.Rand,
) {
	t.Helper()

	source, target := randomPair(random)
	require.NoError(t, c.states[source][target].SetReadOnly(random.Intn(2) == 0))
}

func (c *syncChaos) reload(t *testing.T, peer int) {
	t.Helper()

	documentData, err := c.documents[peer].Save()
	require.NoError(t, err)

	var states [chaosPeers][]byte

	for target := range chaosPeers {
		if peer == target {
			continue
		}

		states[target], err = c.states[peer][target].Save()
		require.NoError(t, err)
		require.NoError(t, c.states[peer][target].Close())
	}

	require.NoError(t, c.documents[peer].Close())

	c.documents[peer], err = automerge.Load(documentData, actor(byte(0x50+peer)))
	require.NoError(t, err)

	for target := range chaosPeers {
		if peer == target {
			continue
		}

		c.states[peer][target], err = c.documents[peer].LoadSyncState(states[target])
		require.NoError(t, err)
	}
}

func (c *syncChaos) assertGenerationQuiesces(
	t *testing.T,
	random *rand.Rand,
) {
	t.Helper()

	source, target := randomPair(random)

	for count := range 10 {
		message, ok, err := c.states[source][target].GenerateMessage()
		require.NoError(t, err)

		if !ok {
			return
		}

		c.last[source][target] = append(c.last[source][target][:0], message...)

		if count == 9 {
			encoded, saveErr := c.states[source][target].Save()
			t.Fatalf(
				"peer %d -> %d did not quiesce; state=%s saveErr=%v",
				source,
				target,
				encoded,
				saveErr,
			)
		}
	}
}

func (c *syncChaos) converge(t *testing.T) {
	t.Helper()

	for source := range chaosPeers {
		for target := range chaosPeers {
			if source != target {
				require.NoError(t, c.states[source][target].SetReadOnly(false))
			}
		}
	}

	for round := range 200 {
		sent := false

		for source := range chaosPeers {
			for target := range chaosPeers {
				if source == target {
					continue
				}

				message, ok, err := c.states[source][target].GenerateMessage()
				require.NoError(t, err)

				if !ok {
					continue
				}

				sent = true

				require.NoError(t, c.states[target][source].ReceiveMessage(message))
			}
		}

		if !sent {
			return
		}

		if round == 199 {
			t.Fatal("sync chaos peers did not converge")
		}
	}
}

func (c *syncChaos) close() {
	for source := range chaosPeers {
		for target := range chaosPeers {
			if source != target && c.states[source][target] != nil {
				_ = c.states[source][target].Close()
			}
		}

		if c.documents[source] != nil {
			_ = c.documents[source].Close()
		}
	}
}

func randomPair(random *rand.Rand) (int, int) {
	source := random.Intn(chaosPeers)

	target := random.Intn(chaosPeers - 1)
	if target >= source {
		target++
	}

	return source, target
}

func chaosSignature(t *testing.T, document *automerge.Document) string {
	t.Helper()

	heads := sortedHeadHex(t, document)

	keys, err := document.Root().Keys()
	require.NoError(t, err)
	sort.Strings(keys)

	var builder strings.Builder
	fmt.Fprintf(&builder, "heads=%v\n", heads)

	for _, key := range keys {
		if key == "body" {
			continue
		}

		value, err := document.Root().Scalar(key)
		require.NoError(t, err)
		fmt.Fprintf(&builder, "%s=%s\n", key, canonicalScalar(value))
	}

	text, err := document.Text("body")
	require.NoError(t, err)

	content, err := text.String()
	require.NoError(t, err)
	fmt.Fprintf(&builder, "body=%q", content)

	return builder.String()
}
