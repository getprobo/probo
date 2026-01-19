// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package authn

import (
	"net/http"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/securecookie"
)

type Cookie struct {
	config *securecookie.Config
}

func (c *Cookie) Set(w http.ResponseWriter, session *coredata.Session) {
	_ = securecookie.Set(
		w,
		c.sessionCookieConfig(time.Until(session.ExpiredAt)),
		session.ID.String(),
	)
}

func (c *Cookie) Clear(w http.ResponseWriter) {
	securecookie.Clear(w, c.sessionCookieConfig(-1*time.Second))
}

func NewCookie(config *securecookie.Config) *Cookie {
	return &Cookie{config}
}

func (c *Cookie) sessionCookieConfig(maxAge time.Duration) securecookie.Config {
	return securecookie.Config{
		Name:     c.config.Name,
		Secret:   c.config.Secret,
		Secure:   c.config.Secure,
		HTTPOnly: c.config.HTTPOnly,
		SameSite: c.config.SameSite,
		Path:     c.config.Path,
		MaxAge:   int(maxAge.Seconds()),
	}
}
