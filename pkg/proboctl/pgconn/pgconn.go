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

package pgconn

import (
	"fmt"
	"net"
	"net/url"

	"go.gearno.de/kit/pg"
)

func NewPgClientFromDSN(dsn string) (*pg.Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN: %w", err)
	}

	var opts []pg.Option

	switch u.Query().Get("sslmode") {
	case "", "disable":
	case "require":
		opts = append(opts, pg.WithUnsecureTLS())
	case "prefer":
		return nil, fmt.Errorf(
			"unsupported sslmode %q (prefer fallback semantics are not supported)",
			u.Query().Get("sslmode"),
		)
	default:
		return nil, fmt.Errorf("unsupported sslmode %q", u.Query().Get("sslmode"))
	}

	if u.Host != "" {
		host := u.Host
		if u.Port() == "" {
			host = net.JoinHostPort(u.Hostname(), "5432")
		}

		opts = append(opts, pg.WithAddr(host))
	}

	if u.User != nil {
		opts = append(opts, pg.WithUser(u.User.Username()))
		if password, ok := u.User.Password(); ok {
			opts = append(opts, pg.WithPassword(password))
		}
	}

	if len(u.Path) > 1 {
		opts = append(opts, pg.WithDatabase(u.Path[1:]))
	}

	return pg.NewClient(opts...)
}
