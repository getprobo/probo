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

package filemanager

import (
	"go.probo.inc/probo/pkg/coredata"
)

func apiPath(file *coredata.File) string {
	if file.Visibility == coredata.FileVisibilityPublic {
		return "/api/files/v1/public/" + file.ID.String()
	}

	return "/api/files/v1/" + file.ID.String()
}

// GenerateFileURL returns the stable app URL routing through the files API.
func (s *Service) GenerateFileURL(file *coredata.File) string {
	return s.baseURL.WithPath(apiPath(file)).MustString()
}
