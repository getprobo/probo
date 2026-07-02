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
