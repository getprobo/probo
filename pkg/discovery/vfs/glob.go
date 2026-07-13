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

package vfs

import (
	"path"
	"strings"
)

// MatchGlob reports whether workspacePath matches pattern. Pattern uses / as the
// separator. A * segment matches exactly one path segment; other segments use
// path.Match semantics.
func MatchGlob(pattern, workspacePath string) bool {
	pattern = NormalizePath(pattern)

	workspacePath = NormalizePath(workspacePath)
	if pattern == "" || workspacePath == "" {
		return false
	}

	return matchGlobParts(strings.Split(pattern, "/"), strings.Split(workspacePath, "/"))
}

func matchGlobParts(patternParts, pathParts []string) bool {
	if len(patternParts) == 0 {
		return len(pathParts) == 0
	}

	if len(pathParts) == 0 {
		return false
	}

	if patternParts[0] == "*" {
		if len(patternParts) == 1 {
			return true
		}

		return matchGlobParts(patternParts[1:], pathParts[1:])
	}

	matched, err := path.Match(patternParts[0], pathParts[0])
	if err != nil || !matched {
		return false
	}

	return matchGlobParts(patternParts[1:], pathParts[1:])
}
