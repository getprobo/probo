// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package cmdutil

import (
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cmd/iostreams"
	"go.probo.inc/probo/pkg/proboctl/pgconn"
)

type Factory struct {
	IOStreams *iostreams.IOStreams
	Version   string
	PgDSN     string

	pgClient *pg.Client
}

// PgClient returns a shared pg client, building it on first use. The client
// is memoized because pg.NewClient registers Prometheus collectors, so
// constructing it more than once panics with a duplicate registration.
func (f *Factory) PgClient() (*pg.Client, error) {
	if f.pgClient != nil {
		return f.pgClient, nil
	}

	if f.PgDSN == "" {
		return nil, fmt.Errorf("set --pg-dsn or DATABASE_URL")
	}

	client, err := pgconn.NewPgClientFromDSN(f.PgDSN)
	if err != nil {
		return nil, err
	}

	f.pgClient = client

	return f.pgClient, nil
}
