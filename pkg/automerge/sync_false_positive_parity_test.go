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

// This file reproduces the upstream Bloom false-positive recovery tests
// (should_handle_false_positive_head and should_handle_chains_of_false_positives
// in rust/automerge/src/sync.rs). A Bloom false positive causes a peer to
// wrongly believe the other already has a change and withhold it; the V2 sync
// protocol must still converge by detecting the missing dependency and
// requesting it. The false positive is located with the reference engine's
// actual Bloom filter (exposed through ReferenceBloomContains) so the scenario
// is deterministic and genuine on the reference engine, while the native engine
// — which uses exact head comparison instead of Bloom filters — must converge in
// the same topology.

package automerge_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// bloomOracle returns a reference document used only to evaluate Bloom filter
// membership. Change hashes are engine-independent, so the same oracle locates
// the false positive for both the native and reference runs.
func bloomOracle(t *testing.T, ctx context.Context) *automerge.Document {
	t.Helper()

	oracle, err := automerge.NewReference(ctx, actor(0xB0))
	require.NoError(t, err)
	closeDocument(t, oracle)

	return oracle
}

func unionSortedHeadHex(left, right []string) []string {
	seen := make(map[string]struct{}, len(left)+len(right))
	union := make([]string, 0, len(left)+len(right))

	for _, value := range append(append([]string{}, left...), right...) {
		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		union = append(union, value)
	}

	sort.Strings(union)

	return union
}

func headHashes(t *testing.T, ctx context.Context, document *automerge.Document) []automerge.Hash {
	t.Helper()

	heads, err := document.Heads(ctx)
	require.NoError(t, err)

	return heads
}

// TestRustSync_ShouldHandleFalsePositiveHead reproduces
// should_handle_false_positive_head: two concurrent changes n1 and n2 are built
// on a shared history where n2 is a false positive in the Bloom filter of {n1}.
// Synchronization must still converge both peers to the union of their heads.
func TestRustSync_ShouldHandleFalsePositiveHead(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				oracle := bloomOracle(t, ctx)

				doc1, err := engine.open(ctx, actor(0xa1))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(ctx, actor(0xd4))
				require.NoError(t, err)
				closeDocument(t, doc2)

				for i := range int64(10) {
					putInt(t, ctx, doc1, "x", i, "x", commitTime.Add(time.Duration(i)*time.Second))
				}

				syncQuiescent(t, ctx, readWriteSyncState(t, ctx, doc1), readWriteSyncState(t, ctx, doc2))

				var n1, n2 *automerge.Document

				for i := 0; ; i++ {
					require.Less(t, i, 10000, "no Bloom false positive found")

					candidate1, err := doc1.Fork(ctx, actor(0x11))
					require.NoError(t, err)

					putRoot(t, ctx, candidate1, "x", fmt.Sprintf("%d @ n1", i), "n1", commitTime.Add(time.Hour))

					candidate2, err := doc1.Fork(ctx, actor(0x22))
					require.NoError(t, err)

					putRoot(t, ctx, candidate2, "x", fmt.Sprintf("%d @ n2", i), "n2", commitTime.Add(time.Hour))

					n1Heads := headHashes(t, ctx, candidate1)
					n2Heads := headHashes(t, ctx, candidate2)

					falsePositive, err := oracle.ReferenceBloomContains(ctx, n1Heads, n2Heads[0])
					require.NoError(t, err)

					if falsePositive {
						n1 = candidate1
						n2 = candidate2

						closeDocument(t, n1)
						closeDocument(t, n2)

						break
					}

					require.NoError(t, candidate1.Close(ctx))
					require.NoError(t, candidate2.Close(ctx))
				}

				allHeads := unionSortedHeadHex(sortedHeadHex(t, ctx, n1), sortedHeadHex(t, ctx, n2))

				syncQuiescent(t, ctx, readWriteSyncState(t, ctx, n1), readWriteSyncState(t, ctx, n2))

				assert.Equal(t, allHeads, sortedHeadHex(t, ctx, n1))
				assert.Equal(t, allHeads, sortedHeadHex(t, ctx, n2))
			},
		)
	}
}

// TestRustSync_ShouldHandleChainsOfFalsePositives reproduces
// should_handle_chains_of_false_positives: two changes chained on one peer are
// both false positives in the other peer's Bloom filter. Synchronization must
// still converge both peers to the union of their heads.
func TestRustSync_ShouldHandleChainsOfFalsePositives(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				oracle := bloomOracle(t, ctx)

				doc1, err := engine.open(ctx, actor(0xa1))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(ctx, actor(0xd4))
				require.NoError(t, err)
				closeDocument(t, doc2)

				for i := range int64(10) {
					putInt(t, ctx, doc1, "x", i, "x", commitTime.Add(time.Duration(i)*time.Second))
				}

				syncQuiescent(t, ctx, readWriteSyncState(t, ctx, doc1), readWriteSyncState(t, ctx, doc2))

				putInt(t, ctx, doc1, "x", 5, "x5", commitTime.Add(time.Hour))
				bloomSeeds := headHashes(t, ctx, doc1)

				findFalsePositive := func(base *automerge.Document, label string) *automerge.Document {
					for i := 0; ; i++ {
						require.Less(t, i, 10000, "no Bloom false positive found for %s", label)

						candidate, err := base.Fork(ctx, actor(0x8c))
						require.NoError(t, err)

						putRoot(t, ctx, candidate, "x", fmt.Sprintf("%d %s", i, label), label, commitTime.Add(2*time.Hour))

						heads := headHashes(t, ctx, candidate)

						falsePositive, err := oracle.ReferenceBloomContains(ctx, bloomSeeds, heads[0])
						require.NoError(t, err)

						if falsePositive {
							return candidate
						}

						require.NoError(t, candidate.Close(ctx))
					}
				}

				chain1 := findFalsePositive(doc2, "at 89abdef")
				closeDocument(t, chain1)

				chain2 := findFalsePositive(chain1, "again")
				closeDocument(t, chain2)

				putRoot(t, ctx, chain2, "x", "final @ 89abcdef", "final", commitTime.Add(3*time.Hour))

				allHeads := unionSortedHeadHex(sortedHeadHex(t, ctx, doc1), sortedHeadHex(t, ctx, chain2))

				syncQuiescent(t, ctx, readWriteSyncState(t, ctx, doc1), readWriteSyncState(t, ctx, chain2))

				assert.Equal(t, allHeads, sortedHeadHex(t, ctx, doc1))
				assert.Equal(t, allHeads, sortedHeadHex(t, ctx, chain2))
			},
		)
	}
}
