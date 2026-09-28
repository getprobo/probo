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

package drivers

import (
	"fmt"
	"mime"
	"net/http"
	"strings"
)

// SettingRejectedError reports that the provider accepted the credential and
// refused a connector setting. Code and Message are Probo's own text, safe to
// log and to show. Setting is the ExtraSetting key the connection check blames,
// empty when the credential cannot reach what the setting names.
type SettingRejectedError struct {
	Code       string
	Setting    string
	StatusCode int
	Message    string
}

func (e *SettingRejectedError) Error() string {
	return fmt.Sprintf("setting rejected: %s (status %d): %s", e.Code, e.StatusCode, e.Message)
}

// respondsWithJSON reports whether the API itself answered, rather than an
// edge or proxy page in front of it.
func respondsWithJSON(resp *http.Response) bool {
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return false
	}

	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}
