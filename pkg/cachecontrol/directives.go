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

package cachecontrol

import "time"

type (
	RequestDirective struct {
		maxAge            *uint64
		maxStale          *uint64
		maxStaleUnbounded bool
		minFresh          *uint64
		noCache           bool
		noStore           bool
		noTransform       bool
		onlyIfCached      bool
		extensions        map[string]string
	}

	ResponseDirective struct {
		maxAge          *uint64
		mustRevalidate  bool
		noCache         []string
		noStore         bool
		noTransform     bool
		public          bool
		private         []string
		proxyRevalidate bool
		sMaxAge         *uint64
		extensions      map[string]string
	}
)

func (d *RequestDirective) MaxAge() (uint64, bool) {
	if v := d.maxAge; v != nil {
		return *v, true
	}

	return 0, false
}

func (d *RequestDirective) MaxStale() (seconds uint64, bounded bool, ok bool) {
	if d.maxStaleUnbounded {
		return 0, false, true
	}

	if v := d.maxStale; v != nil {
		return *v, true, true
	}

	return 0, false, false
}

func (d *RequestDirective) MaxStaleUnbounded() bool {
	return d.maxStaleUnbounded
}

func (d *RequestDirective) MinFresh() (uint64, bool) {
	if v := d.minFresh; v != nil {
		return *v, true
	}

	return 0, false
}

func (d *RequestDirective) NoCache() bool {
	return d.noCache
}

func (d *RequestDirective) NoStore() bool {
	return d.noStore
}

func (d *RequestDirective) NoTransform() bool {
	return d.noTransform
}

func (d *RequestDirective) OnlyIfCached() bool {
	return d.onlyIfCached
}

func (d *RequestDirective) Extensions() map[string]string {
	return d.extensions
}

func (d *RequestDirective) Extension(name string) string {
	return d.extensions[name]
}

func (d *ResponseDirective) MaxAge() (uint64, bool) {
	if v := d.maxAge; v != nil {
		return *v, true
	}

	return 0, false
}

func (d *ResponseDirective) MaxAgeDuration() (time.Duration, bool) {
	seconds, ok := d.MaxAge()
	if !ok {
		return 0, false
	}

	return secondsToDuration(seconds), true
}

func (d *ResponseDirective) MustRevalidate() bool {
	return d.mustRevalidate
}

func (d *ResponseDirective) NoCache() []string {
	return d.noCache
}

func (d *ResponseDirective) NoStore() bool {
	return d.noStore
}

func (d *ResponseDirective) NoTransform() bool {
	return d.noTransform
}

func (d *ResponseDirective) Public() bool {
	return d.public
}

func (d *ResponseDirective) Private() []string {
	return d.private
}

func (d *ResponseDirective) ProxyRevalidate() bool {
	return d.proxyRevalidate
}

func (d *ResponseDirective) SMaxAge() (uint64, bool) {
	if v := d.sMaxAge; v != nil {
		return *v, true
	}

	return 0, false
}

func (d *ResponseDirective) Extensions() map[string]string {
	return d.extensions
}

func (d *ResponseDirective) Extension(name string) string {
	return d.extensions[name]
}
