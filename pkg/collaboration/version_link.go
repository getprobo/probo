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

package collaboration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/richtext"
)

const (
	// versionLinkClockSkew absorbs the gap between the client clock stamped
	// on a change and the server clock stamped on the version.
	versionLinkClockSkew = 10 * time.Minute
	// versionLinkMaxCandidates bounds how many changes are materialized to
	// find the state a version was taken at.
	versionLinkMaxCandidates = 500
)

type (
	// VersionLink is the outcome of tying one legacy version to the history.
	VersionLink struct {
		VersionID gid.GID
		Major     int
		Minor     int
		// Heads are the encoded heads the version was linked to. Nil when no
		// state of the history matches the version content.
		Heads []byte
	}

	renderAt func(heads []automerge.Hash) (string, error)
)

// LinkVersionHeads ties the versions of a document that predate the link to
// the revision of the history whose content they hold. A version is linked to
// the state after the change that reproduces its content, the closest one to
// the time it was created. A version that matches no state is left unlinked
// and keeps standing alone in the history. With dryRun nothing is written.
func (s *Service) LinkVersionHeads(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	dryRun bool,
) ([]VersionLink, error) {
	var versions coredata.DocumentVersions

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return versions.LoadWithoutCollaborationHeadsByDocumentID(ctx, conn, scope, documentID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load unlinked document versions: %w", err)
	}

	if len(versions) == 0 {
		return nil, nil
	}

	opened, err := s.Open(ctx, scope, documentID)
	if err != nil {
		if errors.Is(err, ErrDocumentCollaborationNotFound) {
			return unlinkedVersions(versions), nil
		}

		return nil, fmt.Errorf("cannot open collaboration to link versions: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	changes, err := opened.Document.ChangeHistory()
	if err != nil {
		return nil, fmt.Errorf("cannot read collaboration change history: %w", err)
	}

	render := func(heads []automerge.Hash) (string, error) {
		return materializeAcceptedAtHeads(opened.Document, heads)
	}

	links := make([]VersionLink, len(versions))

	for i, version := range versions {
		links[i] = VersionLink{
			VersionID: version.ID,
			Major:     version.Major,
			Minor:     version.Minor,
		}

		hash, ok, err := matchVersionChange(version.Content, version.CreatedAt, changes, render)
		if err != nil {
			return nil, fmt.Errorf("cannot match version %d.%d to the history: %w", version.Major, version.Minor, err)
		}

		if !ok {
			continue
		}

		links[i].Heads = encodeHeads([]automerge.Hash{hash})

		if dryRun {
			continue
		}

		version.CollaborationHeads = links[i].Heads

		err = s.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return version.UpdateCollaborationHeads(ctx, conn, scope)
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return links, nil
}

func unlinkedVersions(versions coredata.DocumentVersions) []VersionLink {
	links := make([]VersionLink, len(versions))
	for i, version := range versions {
		links[i] = VersionLink{VersionID: version.ID, Major: version.Major, Minor: version.Minor}
	}

	return links
}

// matchVersionChange finds the change after which the document reads as
// content. It starts from the last change made before the version was created
// and walks back, then forward, so the closest state wins.
func matchVersionChange(
	content string,
	createdAt time.Time,
	changes []automerge.ChangeInfo,
	render renderAt,
) (automerge.Hash, bool, error) {
	if len(changes) == 0 {
		return automerge.Hash{}, false, nil
	}

	start := 0

	for i, change := range changes {
		if !change.Time.IsZero() && change.Time.After(createdAt.Add(versionLinkClockSkew)) {
			break
		}

		start = i
	}

	try := func(index int) (bool, error) {
		rendered, err := render([]automerge.Hash{changes[index].Hash})
		if err != nil {
			return false, err
		}

		return rendered == content, nil
	}

	for i, tried := start, 0; i >= 0 && tried < versionLinkMaxCandidates; i, tried = i-1, tried+1 {
		matched, err := try(i)
		if err != nil {
			return automerge.Hash{}, false, err
		}

		if matched {
			return changes[i].Hash, true, nil
		}
	}

	for i, tried := start+1, 0; i < len(changes) && tried < versionLinkMaxCandidates; i, tried = i+1, tried+1 {
		matched, err := try(i)
		if err != nil {
			return automerge.Hash{}, false, err
		}

		if matched {
			return changes[i].Hash, true, nil
		}
	}

	return automerge.Hash{}, false, nil
}

// materializeAcceptedAtHeads renders the document at heads the way a version
// is taken: suggestions that are still open are left out.
func materializeAcceptedAtHeads(
	document *automerge.Document,
	heads []automerge.Hash,
) (string, error) {
	spans, err := spansAtHeads(document, heads)
	if err != nil {
		return "", err
	}

	content, err := automergeprosemirror.Render(ProjectAcceptedSpans(spans))
	if err != nil {
		return "", fmt.Errorf("cannot render collaboration spans at heads: %w", err)
	}

	content, err = richtext.DocumentSchema.Sanitize(content)
	if err != nil {
		return "", fmt.Errorf("cannot sanitize collaboration content at heads: %w", err)
	}

	return content, nil
}
