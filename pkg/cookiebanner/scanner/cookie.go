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

package scanner

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"go.probo.inc/probo/pkg/coredata"
)

func mapSetCookie(line string, setByURL string) (Cookie, bool) {
	parsed, err := http.ParseSetCookie(line)
	if err != nil || parsed.Name == "" || strings.EqualFold(parsed.Name, consentCookieName) {
		return Cookie{}, false
	}

	cookie := Cookie{
		Name:      parsed.Name,
		Source:    coredata.CookieSourceHTTP,
		ValueSize: len(parsed.Value),
		HTTPOnly:  new(parsed.HttpOnly),
		Secure:    new(parsed.Secure),
	}
	if setByURL != "" {
		cookie.SetByURL = new(setByURL)
	}

	if parsed.MaxAge > 0 {
		cookie.MaxAgeSeconds = new(parsed.MaxAge)
	} else if parsed.MaxAge < 0 {
		cookie.MaxAgeSeconds = new(0)
	} else if !parsed.Expires.IsZero() {
		seconds := max(int(time.Until(parsed.Expires).Seconds()), 0)
		cookie.MaxAgeSeconds = new(seconds)
	}

	if parsed.Domain == "" {
		cookie.HostOnly = new(true)
	} else {
		domain := strings.TrimPrefix(parsed.Domain, ".")
		cookie.Domain = new(domain)
		cookie.HostOnly = new(false)
	}

	switch parsed.SameSite {
	case http.SameSiteStrictMode:
		cookie.SameSite = new(coredata.CookieSameSiteStrict)
	case http.SameSiteLaxMode:
		cookie.SameSite = new(coredata.CookieSameSiteLax)
	case http.SameSiteNoneMode:
		cookie.SameSite = new(coredata.CookieSameSiteNone)
	}

	if parsed.Partitioned {
		cookie.Partitioned = new(true)
	}

	return cookie, true
}

func mapResourceType(kind network.ResourceType, frameID, mainFrame networkFrame) (coredata.TrackerResourceType, bool) {
	switch kind {
	case network.ResourceTypeScript:
		return coredata.TrackerResourceTypeScript, true
	case network.ResourceTypeStylesheet:
		return coredata.TrackerResourceTypeStylesheet, true
	case network.ResourceTypeImage:
		return coredata.TrackerResourceTypeImage, true
	case network.ResourceTypeMedia, network.ResourceTypeTextTrack:
		return coredata.TrackerResourceTypeMedia, true
	case network.ResourceTypeFont:
		return coredata.TrackerResourceTypeFont, true
	case network.ResourceTypeXHR, network.ResourceTypeFetch:
		return coredata.TrackerResourceTypeFetch, true
	case network.ResourceTypePing:
		return coredata.TrackerResourceTypeBeacon, true
	case network.ResourceTypeDocument:
		if mainFrame != "" && frameID != "" && frameID != mainFrame {
			return coredata.TrackerResourceTypeIframe, true
		}
	}

	return "", false
}

// networkFrame is the CDP frame id without importing cdp into every helper.
type networkFrame = string

func resourceOf(rawURL, initiator string, kind coredata.TrackerResourceType) (Resource, bool) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Resource{}, false
	}

	path := u.Path
	if path == "" {
		path = "/"
	}

	resource := Resource{
		Type:   kind,
		Origin: u.Scheme + "://" + strings.ToLower(u.Host),
		Path:   path,
	}
	if initiator != "" {
		resource.InitiatorURL = new(initiator)
	}

	return resource, true
}

func blockSDKPost(method, rawURL, bannerID, apiBase string) bool {
	apiOrigin, apiOK := requestOrigin(apiBase)
	requestOriginURL, requestOK := requestOrigin(rawURL)
	if method != http.MethodPost || bannerID == "" || !apiOK || !requestOK || !strings.EqualFold(apiOrigin, requestOriginURL) {
		return false
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	path := strings.TrimSuffix(u.Path, "/")
	base := "/" + bannerID + "/"

	return strings.HasSuffix(path, base+"report") ||
		strings.HasSuffix(path, base+"detected-cookies") ||
		strings.HasSuffix(path, base+"consents")
}
