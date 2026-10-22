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

package automerge

import (
	"fmt"

	"go.probo.inc/probo/pkg/automerge/internal/storage"
	"go.probo.inc/probo/pkg/automerge/internal/sync"
)

// ActorsFromSyncMessage returns the distinct Automerge actor ids carried by the
// change chunks inside a sync message. It reads only the inbound message, so the
// actors belong to the peer that sent it rather than to a later heads diff of
// the shared document.
func ActorsFromSyncMessage(message []byte) ([][]byte, error) {
	parsed, err := sync.ParseMessage(message)
	if err != nil {
		return nil, fmt.Errorf("cannot parse sync message: %w", err)
	}

	if len(parsed.Changes) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(parsed.Changes))
	actors := make([][]byte, 0, len(parsed.Changes))

	for i, change := range parsed.Changes {
		document, err := storage.DecodePartial(change)
		if err != nil {
			return nil, fmt.Errorf("cannot decode sync change %d: %w", i, err)
		}

		for _, decoded := range document.Changes {
			key := string(decoded.Actor)
			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}

			actors = append(actors, decoded.Actor.Bytes())
		}
	}

	return actors, nil
}
