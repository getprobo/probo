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

package asana

import (
	"fmt"

	"go.probo.inc/probo/pkg/cmd/iostreams"
)

// nameSearchCap is the most Asana typeahead returns in one name search.
const nameSearchCap = 50

// WriteSearchLimitHint writes a stderr note when a name search fills the page
// it can return. Name search does not page. A --limit above the typeahead cap
// still stops at that cap, so the note must not wait for shown to equal --limit.
func WriteSearchLimitHint(ios *iostreams.IOStreams, query string, shown, limit int, noun string) {
	if query == "" || shown == 0 {
		return
	}

	ceiling := limit
	if ceiling <= 0 || ceiling > nameSearchCap {
		ceiling = nameSearchCap
	}

	if shown != ceiling {
		return
	}

	if limit > nameSearchCap {
		_, _ = fmt.Fprintf(
			ios.ErrOut,
			"\nName search returned %d %s, the most it can return. Narrow the query; name search does not page.\n",
			shown,
			noun,
		)

		return
	}

	_, _ = fmt.Fprintf(
		ios.ErrOut,
		"\nName search returned %d %s, the --limit. Raise --limit or narrow the query; name search does not page.\n",
		shown,
		noun,
	)
}
