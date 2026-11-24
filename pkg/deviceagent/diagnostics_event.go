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

package deviceagent

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf16"
)

func wevtutilQueryArgs(limit int) []string {
	count := max(limit/32, 1)

	return []string{
		"qe",
		"Application",
		"/q:*[System[Provider[@Name='" + WindowsEventSource + "']]]",
		"/c:" + strconv.Itoa(count),
		"/rd:true",
		"/f:xml",
	}
}

func decodeCommandOutput(out []byte) string {
	if len(out) < 2 || out[0] != 0xFF || out[1] != 0xFE {
		return string(out)
	}

	units := make([]uint16, 0, (len(out)-2)/2)
	for i := 2; i+1 < len(out); i += 2 {
		units = append(units, uint16(out[i])|uint16(out[i+1])<<8)
	}

	for len(units) > 0 && units[len(units)-1] == 0 {
		units = units[:len(units)-1]
	}

	return string(utf16.Decode(units))
}

func eventMessagesFromXML(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "<?xml") {
		if end := strings.Index(raw, "?>"); end >= 0 {
			raw = strings.TrimSpace(raw[end+2:])
		}
	}

	decoder := xml.NewDecoder(strings.NewReader(raw))
	messages := make([]string, 0)

	for {
		var event struct {
			XMLName xml.Name
			Data    []string `xml:"EventData>Data"`
		}

		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("cannot parse Windows event: %w", err)
		}

		if event.XMLName.Local != "Event" {
			continue
		}

		parts := make([]string, 0, len(event.Data))
		for _, item := range event.Data {
			item = strings.TrimSpace(item)
			if item != "" {
				parts = append(parts, item)
			}
		}

		if len(parts) == 0 {
			continue
		}

		messages = append(messages, strings.Join(parts, "\n"))
	}

	return messages, nil
}

func eventMessageFromXML(raw string) (string, error) {
	var event struct {
		Data []string `xml:"EventData>Data"`
	}
	if err := xml.Unmarshal([]byte(raw), &event); err != nil {
		return "", fmt.Errorf("cannot parse Windows event: %w", err)
	}

	parts := make([]string, 0, len(event.Data))
	for _, item := range event.Data {
		item = strings.TrimSpace(item)
		if item != "" {
			parts = append(parts, item)
		}
	}

	return strings.Join(parts, "\n"), nil
}

// joinNewestEventLines keeps the newest lines, up to limit bytes, and
// returns them oldest first. lines is newest first.
func joinNewestEventLines(lines []string, limit int) string {
	if limit <= 0 {
		return ""
	}

	kept := make([]string, 0, len(lines))
	size := 0

	for _, line := range lines {
		next := len(line)
		if len(kept) > 0 {
			next++
		}

		if len(kept) > 0 && size+next > limit {
			break
		}

		kept = append(kept, line)

		size += next
		if size >= limit {
			break
		}
	}

	for i, j := 0, len(kept)-1; i < j; i, j = i+1, j-1 {
		kept[i], kept[j] = kept[j], kept[i]
	}

	return strings.Join(kept, "\n")
}
