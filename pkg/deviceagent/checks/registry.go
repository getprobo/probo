// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package checks

import (
	"context"
	"sort"
	"sync"
)

var (
	registryMu sync.Mutex
	registry   []Check
)

// Register adds a check implementation to the process registry.
func Register(key string, run func(context.Context) Result) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(
		registry,
		funcCheck{
			key: key,
			run: run,
		},
	)
}

// All returns a stable snapshot of registered checks.
func All() []Check {
	registryMu.Lock()
	defer registryMu.Unlock()

	out := make([]Check, len(registry))
	copy(out, registry)
	sort.SliceStable(
		out,
		func(i, j int) bool {
			return out[i].Key() < out[j].Key()
		},
	)
	return out
}
