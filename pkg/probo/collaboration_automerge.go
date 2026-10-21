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

package probo

import (
	"context"
	"encoding/hex"
	"fmt"

	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror"
)

func newAutomergeDocument(ctx context.Context) (*automerge.Document, error) {
	actorID, err := automerge.NewActorID()
	if err != nil {
		return nil, fmt.Errorf("cannot generate collaboration actor ID: %w", err)
	}

	document, err := automerge.New(actorID)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge document: %w", err)
	}

	return document, nil
}

func loadAutomergeDocument(ctx context.Context, snapshot []byte) (*automerge.Document, error) {
	actorID, err := automerge.NewActorID()
	if err != nil {
		return nil, fmt.Errorf("cannot generate collaboration actor ID: %w", err)
	}

	document, err := automerge.Load(snapshot, actorID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Automerge document: %w", err)
	}

	return document, nil
}

func encodeHeads(heads []automerge.Hash) []byte {
	encoded := make([]byte, 0, len(heads)*32)
	for _, head := range heads {
		encoded = append(encoded, head[:]...)
	}

	return encoded
}

func decodeHeads(encoded []byte) ([]automerge.Hash, error) {
	if len(encoded) == 0 {
		return nil, nil
	}

	if len(encoded)%len(automerge.Hash{}) != 0 {
		return nil, fmt.Errorf(
			"collaboration heads length %d is not a multiple of %d",
			len(encoded),
			len(automerge.Hash{}),
		)
	}

	heads := make([]automerge.Hash, 0, len(encoded)/len(automerge.Hash{}))
	for offset := 0; offset < len(encoded); offset += len(automerge.Hash{}) {
		var head automerge.Hash
		copy(head[:], encoded[offset:offset+len(automerge.Hash{})])
		heads = append(heads, head)
	}

	return heads, nil
}

func materialize(ctx context.Context, document *automerge.Document) (string, error) {
	text, err := document.Text("body")
	if err != nil {
		return "", fmt.Errorf("cannot get collaboration body: %w", err)
	}

	spans, err := text.Spans()
	if err != nil {
		return "", fmt.Errorf("cannot get collaboration spans: %w", err)
	}

	content, err := automergeprosemirror.Render(spans)
	if err != nil {
		return "", fmt.Errorf("cannot render collaboration spans: %w", err)
	}

	content, err = prosemirror.SanitizeDocumentJSON(content)
	if err != nil {
		return "", fmt.Errorf("cannot sanitize collaboration content: %w", err)
	}

	return content, nil
}

func materializeAtHeads(
	document *automerge.Document,
	heads []automerge.Hash,
) (string, error) {
	text, err := document.Text("body")
	if err != nil {
		return "", fmt.Errorf("cannot get collaboration body at heads: %w", err)
	}

	spans, err := text.SpansAt(heads)
	if err != nil {
		return "", fmt.Errorf("cannot read collaboration spans at heads: %w", err)
	}

	content, err := automergeprosemirror.Render(spans)
	if err != nil {
		return "", fmt.Errorf("cannot render collaboration spans at heads: %w", err)
	}

	return content, nil
}

func writeSpans(document *automerge.Document, contentJSON string, createBody bool) error {
	spans, err := automergeprosemirror.ToSpans(contentJSON)
	if err != nil {
		return fmt.Errorf("cannot convert content to collaboration spans: %w", err)
	}

	text, err := document.Text("body")
	if err != nil {
		if !createBody {
			return fmt.Errorf("cannot get collaboration body: %w", err)
		}

		text, err = document.CreateText("body")
		if err != nil {
			return fmt.Errorf("cannot create collaboration body: %w", err)
		}
	}

	if err := text.UpdateSpans(spans, automergeprosemirror.UpdateSpansConfig()); err != nil {
		return fmt.Errorf("cannot write collaboration spans: %w", err)
	}

	return nil
}

// ParseAutomergeHashes decodes lowercase hex change hashes.
func ParseAutomergeHashes(values []string) ([]automerge.Hash, error) {
	hashes := make([]automerge.Hash, len(values))
	for i, value := range values {
		raw, err := hex.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("cannot decode Automerge hash %q: %w", value, err)
		}

		if len(raw) != len(automerge.Hash{}) {
			return nil, fmt.Errorf(
				"cannot decode Automerge hash %q: expected %d bytes, got %d",
				value,
				len(automerge.Hash{}),
				len(raw),
			)
		}

		copy(hashes[i][:], raw)
	}

	return hashes, nil
}

// EncodeAutomergeHashes encodes change hashes as lowercase hex strings.
func EncodeAutomergeHashes(hashes []automerge.Hash) []string {
	values := make([]string, len(hashes))
	for i, hash := range hashes {
		values[i] = hash.String()
	}

	return values
}
