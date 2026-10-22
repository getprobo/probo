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
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
)

const (
	changeBatchSize   = 1000
	compactionChanges = 500
	seedLease         = 30 * time.Second
	snapshotMaxBytes  = 8 * 1024 * 1024
	contentMaxText    = 200_000
)

type (
	CollaborationService struct {
		pg     *pg.Client
		logger *log.Logger
	}

	DocumentCollaboration struct {
		Document      *automerge.Document
		Revision      int64
		SeedContent   string
		NeedsSeed     bool
		SeedClaimedAt *time.Time
	}
)

func NewCollaborationService(pgClient *pg.Client, logger *log.Logger) *CollaborationService {
	return &CollaborationService{
		pg:     pgClient,
		logger: logger,
	}
}
