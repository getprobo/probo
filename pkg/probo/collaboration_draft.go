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
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// ErrDocumentArchived is returned when a draft cannot be edited because its
// owning document is archived.
type ErrDocumentArchived struct{}

// ErrDocumentVersionGenerated is returned when a draft cannot be edited because its
// owning document version is generated rather than hand-written.
type ErrDocumentVersionGenerated struct{}

func (e ErrDocumentArchived) Error() string {
	return "cannot modify an archived document"
}

func (e ErrDocumentVersionGenerated) Error() string {
	return "cannot edit content of a generated document version"
}

// assertDraftEditable checks that documentID can currently be edited,
// and returns its organization id. Editability is the document: it is not
// archived and it is not generated.
func assertDraftEditable(
	ctx context.Context,
	scope coredata.Scoper,
	conn pg.Querier,
	documentID gid.GID,
) (gid.GID, error) {
	document := &coredata.Document{}
	if err := document.LoadByID(ctx, conn, scope, documentID); err != nil {
		return gid.GID{}, fmt.Errorf("cannot load document: %w", err)
	}

	if document.ArchivedAt != nil {
		return gid.GID{}, &ErrDocumentArchived{}
	}

	if document.WriteMode == coredata.DocumentWriteModeGenerated {
		return gid.GID{}, &ErrDocumentVersionGenerated{}
	}

	return document.OrganizationID, nil
}

// loadSeedContent returns the text used to seed a collaboration document
// that has never been opened. The live text lives on the document until the
// CRDT is seeded.
func loadSeedContent(
	ctx context.Context,
	scope coredata.Scoper,
	conn pg.Querier,
	documentID gid.GID,
) (string, error) {
	document := &coredata.Document{}
	if err := document.LoadByID(ctx, conn, scope, documentID); err != nil {
		return "", fmt.Errorf("cannot load document for seed: %w", err)
	}

	return document.Content, nil
}

// writeLiveContent stores the materialized CRDT text on the document. Versions
// stay frozen snapshots; this is the projection readers use until they open
// the CRDT.
func writeLiveContent(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	documentID gid.GID,
	content string,
) error {
	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
		return fmt.Errorf("cannot load document for content: %w", err)
	}

	document.Content = content
	document.UpdatedAt = time.Now()

	if err := document.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update document content: %w", err)
	}

	return nil
}
