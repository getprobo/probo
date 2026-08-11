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

package testutil

import (
	"encoding/json"
	"fmt"
	"mime"
	"strings"
)

func collectSSEDataPayloads(body []byte) [][]byte {
	var (
		events [][]byte
		buf    []string
	)

	flush := func() {
		if len(buf) > 0 {
			events = append(events, []byte(strings.Join(buf, "\n")))
			buf = nil
		}
	}

	for line := range strings.Lines(string(body)) {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			flush()

			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		before, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		if strings.TrimSpace(before) != "data" {
			continue
		}

		buf = append(buf, strings.TrimSpace(after))
	}

	flush()

	return events
}

func decodeMCPResponseBody(contentType string, body []byte, requestID int) ([]byte, error) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "text/event-stream" {
		return body, nil
	}

	for _, payload := range collectSSEDataPayloads(body) {
		var envelope struct {
			ID *int `json:"id"`
		}

		if err := json.Unmarshal(payload, &envelope); err != nil {
			continue
		}

		if envelope.ID == nil || *envelope.ID != requestID {
			continue
		}

		return payload, nil
	}

	return nil, fmt.Errorf("SSE response did not contain JSON-RPC id %d", requestID)
}
