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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestSlackbotEvent_InsertDeduplicatesEventID(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	eventID := "E-coredata-" + gid.New(gid.NilTenant, coredata.SlackbotEventEntityType).String()
	event := coredata.NewSlackbotEvent(eventID, []byte(`{"first":true}`))
	duplicate := coredata.NewSlackbotEvent(eventID, []byte(`{"first":false}`))
	event.NextAttemptAt = new(time.Now().Add(time.Hour))

	var inserted bool

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				var err error

				inserted, err = event.Insert(ctx, conn)

				return err
			},
		),
	)
	assert.True(t, inserted)
	t.Cleanup(
		func() {
			_ = pgClient.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					return event.Delete(ctx, conn)
				},
			)
		},
	)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				var err error

				inserted, err = duplicate.Insert(ctx, conn)

				return err
			},
		),
	)
	assert.False(t, inserted)

	var persisted coredata.SlackbotEvent

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return persisted.LoadByEventID(ctx, conn, eventID)
			},
		),
	)
	assert.JSONEq(t, `{"first":true}`, string(persisted.Envelope))
}
