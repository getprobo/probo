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

package llm

type (
	Part interface {
		part()
	}

	TextPart struct {
		Text string `json:"text"`
	}

	ImagePart struct {
		URL string `json:"url"`
	}

	FilePart struct {
		Data     string `json:"data"`      // base64-encoded content
		MimeType string `json:"mime_type"` // e.g. "application/pdf", "text/csv", "image/png"
		Filename string `json:"filename"`
	}

	ThinkingPart struct {
		Text      string
		Signature string // Anthropic thinking signature for multi-turn continuity
	}
)

func (TextPart) part()     {}
func (ImagePart) part()    {}
func (FilePart) part()     {}
func (ThinkingPart) part() {}
