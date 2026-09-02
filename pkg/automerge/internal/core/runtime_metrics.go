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

package core

import (
	"sync/atomic"

	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

type RuntimeMetrics struct {
	GeneralReconciles     uint64
	DirectColumnBatches   uint64
	GlobalOrderFallbacks  uint64
	SnapshotReplacements  uint64
	FullColumnEncodings   uint64
	QueryIndexRescues     uint64
	SemanticChangeRows    uint64
	SemanticOperationRows uint64
	DirectPlanningRows    uint64
	DirectOperationCopies uint64
}

var runtimeMetrics struct {
	generalReconciles     atomic.Uint64
	directColumnBatches   atomic.Uint64
	globalOrderFallbacks  atomic.Uint64
	snapshotReplacements  atomic.Uint64
	queryIndexRescues     atomic.Uint64
	semanticChangeRows    atomic.Uint64
	semanticOperationRows atomic.Uint64
	directPlanningRows    atomic.Uint64
	directOperationCopies atomic.Uint64
}

func ResetRuntimeMetrics() {
	runtimeMetrics.generalReconciles.Store(0)
	runtimeMetrics.directColumnBatches.Store(0)
	runtimeMetrics.globalOrderFallbacks.Store(0)
	runtimeMetrics.snapshotReplacements.Store(0)
	runtimeMetrics.queryIndexRescues.Store(0)
	runtimeMetrics.semanticChangeRows.Store(0)
	runtimeMetrics.semanticOperationRows.Store(0)
	runtimeMetrics.directPlanningRows.Store(0)
	runtimeMetrics.directOperationCopies.Store(0)
	storage.ResetRuntimeMetrics()
}

func ReadRuntimeMetrics() RuntimeMetrics {
	return RuntimeMetrics{
		GeneralReconciles:     runtimeMetrics.generalReconciles.Load(),
		DirectColumnBatches:   runtimeMetrics.directColumnBatches.Load(),
		GlobalOrderFallbacks:  runtimeMetrics.globalOrderFallbacks.Load(),
		SnapshotReplacements:  runtimeMetrics.snapshotReplacements.Load(),
		FullColumnEncodings:   storage.FullColumnEncodings(),
		QueryIndexRescues:     runtimeMetrics.queryIndexRescues.Load(),
		SemanticChangeRows:    runtimeMetrics.semanticChangeRows.Load(),
		SemanticOperationRows: runtimeMetrics.semanticOperationRows.Load(),
		DirectPlanningRows:    runtimeMetrics.directPlanningRows.Load(),
		DirectOperationCopies: runtimeMetrics.directOperationCopies.Load(),
	}
}
