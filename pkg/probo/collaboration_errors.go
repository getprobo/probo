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
	"errors"
	"fmt"
)

type ErrCollaborationStateTooLarge struct {
	Size int
}

var (
	ErrDocumentCollaborationNotSeeded = errors.New("collaboration document is not initialized")
	ErrSeedInProgress                 = errors.New("collaboration document seed is in progress")
	ErrDocumentCollaborationStale     = errors.New("collaboration document changed; read it again before editing")

	ErrDocumentCollaborationBaseHeadsUnknown = errors.New(
		"collaboration base heads are unknown; read the document again",
	)
	ErrDocumentCollaborationContentOnSeeded = errors.New(
		"seeded draft rejects full content; use fragment edits",
	)
	ErrDocumentCollaborationEditsNotSeeded = errors.New(
		"collaboration document is not initialized; send full content first",
	)
)

func (e ErrCollaborationStateTooLarge) Error() string {
	return fmt.Sprintf(
		"collaboration state is too large: %d bytes exceeds %d bytes",
		e.Size,
		snapshotMaxBytes,
	)
}
