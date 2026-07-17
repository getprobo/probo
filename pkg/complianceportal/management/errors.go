// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package management

import "errors"

var (
	// ErrCustomDomainNotActive is returned when a domain is set as primary
	// while its SSL certificate is not yet active.
	ErrCustomDomainNotActive = errors.New("custom domain SSL certificate is not active")

	// ErrCustomDomainManaged is returned when an operation is attempted on the
	// managed probopage subdomain that is only allowed on customer domains.
	ErrCustomDomainManaged = errors.New("managed custom domain cannot be modified")

	// ErrCustomDomainNotFound is returned when no custom domain exists for the
	// requested resource.
	ErrCustomDomainNotFound = errors.New("custom domain not found")
)
