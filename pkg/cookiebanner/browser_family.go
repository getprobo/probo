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

package cookiebanner

import (
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// BrowserFamily classifies a User-Agent into a coarse family. Edge is
// checked before Chrome because Chromium Edge contains both tokens.
// Brave and other Chromium shells are indistinguishable from Chrome.
func BrowserFamily(userAgent string) coredata.DiscoveryBrowserFamily {
	switch {
	case strings.Contains(userAgent, "Edg/") || strings.Contains(userAgent, "EdgiOS") || strings.Contains(userAgent, "EdgA"):
		return coredata.DiscoveryBrowserFamilyEdge
	case strings.Contains(userAgent, "Chrome/") || strings.Contains(userAgent, "CriOS"):
		return coredata.DiscoveryBrowserFamilyChrome
	case strings.Contains(userAgent, "Firefox/") || strings.Contains(userAgent, "FxiOS"):
		return coredata.DiscoveryBrowserFamilyFirefox
	case strings.Contains(userAgent, "Safari/"):
		return coredata.DiscoveryBrowserFamilySafari
	default:
		return coredata.DiscoveryBrowserFamilyOther
	}
}
