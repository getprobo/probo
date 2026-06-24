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
