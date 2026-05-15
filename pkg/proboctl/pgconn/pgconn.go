// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
