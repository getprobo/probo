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
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
)

var localeSegment = regexp.MustCompile(`(?i)^[a-z]{2}([-_][a-z]{2,4})?$`)

type (
	// frontier is one queue of templates ranked at each pop: the start
	// page, then a page embedding a host no visited page embedded, then the
	// least visited section, then links before sitemap entries, then
	// shallower pages, then insertion order.
	frontier struct {
		visited  map[string]struct{}
		queued   map[string]*frontierEntry
		entries  []*frontierEntry
		sections map[string]int
		covered  map[string]struct{}
		maxQueue int
	}

	frontierEntry struct {
		item    QueueItem
		key     string
		section string
	}
)

func newFrontier(start Checkpoint, maxQueue int) frontier {
	f := frontier{
		visited:  map[string]struct{}{},
		queued:   map[string]*frontierEntry{},
		sections: map[string]int{},
		covered:  map[string]struct{}{},
		maxQueue: maxQueue,
	}
	for _, key := range start.VisitedTemplates {
		f.visited[key] = struct{}{}
		f.sections[sectionOf(key)]++
	}

	f.cover(start.CoveredHosts)

	for _, item := range start.Queue {
		f.push(item)
	}

	return f
}

// push queues item unless its template was visited. A template keeps one
// entry: a link upgrades a sitemap entry, and embed hosts are merged. At
// the cap, a link displaces the newest sitemap entry with no embed hosts.
func (f *frontier) push(item QueueItem) {
	f.insert(item, true)
}

// requeue puts an interrupted page back. It ignores the queue cap so the
// page is not dropped.
func (f *frontier) requeue(item QueueItem) {
	f.insert(item, false)
}

func (f *frontier) insert(item QueueItem, capped bool) {
	key, ok := templateKey(item.URL)
	if !ok {
		return
	}

	if _, seen := f.visited[key]; seen {
		return
	}

	if entry, waiting := f.queued[key]; waiting {
		entry.merge(item)

		return
	}

	if capped && f.maxQueue > 0 && len(f.entries) >= f.maxQueue && !f.evictForLink(item) {
		return
	}

	item.EmbedHosts = unionHosts(nil, item.EmbedHosts)
	entry := &frontierEntry{item: item, key: key, section: sectionOf(key)}
	f.entries = append(f.entries, entry)
	f.queued[key] = entry
}

func (f *frontier) evictForLink(item QueueItem) bool {
	if item.FromSitemap {
		return false
	}

	for i, entry := range slices.Backward(f.entries) {
		if entry.item.FromSitemap && len(entry.item.EmbedHosts) == 0 {
			f.removeAt(i)

			return true
		}
	}

	return false
}

func (f *frontier) pop() (QueueItem, bool) {
	if len(f.entries) == 0 {
		return QueueItem{}, false
	}

	best := 0
	for i := 1; i < len(f.entries); i++ {
		if f.before(f.entries[i], f.entries[best]) {
			best = i
		}
	}

	item := f.entries[best].item
	f.removeAt(best)

	return item, true
}

func (f *frontier) before(a, b *frontierEntry) bool {
	if as, bs := a.item.isStart(), b.item.isStart(); as != bs {
		return as
	}

	if an, bn := f.uncovered(a), f.uncovered(b); an != bn {
		return an
	}

	if av, bv := f.sections[a.section], f.sections[b.section]; av != bv {
		return av < bv
	}

	if a.item.FromSitemap != b.item.FromSitemap {
		return !a.item.FromSitemap
	}

	return a.item.Depth < b.item.Depth
}

func (f *frontier) uncovered(entry *frontierEntry) bool {
	for _, host := range entry.item.EmbedHosts {
		if _, ok := f.covered[host]; !ok {
			return true
		}
	}

	return false
}

func (f *frontier) markVisited(raw string) {
	key, ok := templateKey(raw)
	if !ok {
		return
	}

	if _, seen := f.visited[key]; !seen {
		f.visited[key] = struct{}{}
		f.sections[sectionOf(key)]++
	}

	if entry, waiting := f.queued[key]; waiting {
		f.removeAt(slices.Index(f.entries, entry))
	}
}

// cover records hosts a visited page embedded or requested, so pages that
// only embed those hosts lose their priority.
func (f *frontier) cover(hosts []string) {
	for _, host := range hosts {
		f.covered[host] = struct{}{}
	}
}

func (f *frontier) len() int {
	return len(f.entries)
}

func (f *frontier) removeAt(i int) {
	delete(f.queued, f.entries[i].key)
	f.entries = slices.Delete(f.entries, i, i+1)
}

func (f *frontier) checkpoint(pages int, elapsed time.Duration, truncated bool) Checkpoint {
	visited := make([]string, 0, len(f.visited))
	for key := range f.visited {
		visited = append(visited, key)
	}

	sort.Strings(visited)

	queue := make([]QueueItem, 0, len(f.entries))
	for _, entry := range f.entries {
		queue = append(queue, entry.item)
	}

	var covered []string
	for host := range f.covered {
		covered = append(covered, host)
	}

	sort.Strings(covered)

	return Checkpoint{
		VisitedTemplates: visited,
		Queue:            queue,
		CoveredHosts:     covered,
		PagesScanned:     pages,
		ElapsedMS:        elapsed.Milliseconds(),
		Truncated:        truncated,
	}
}

func (e *frontierEntry) merge(item QueueItem) {
	if !item.FromSitemap && (e.item.FromSitemap || item.Depth < e.item.Depth) {
		e.item.FromSitemap = false
		e.item.Depth = item.Depth
	}

	e.item.EmbedHosts = unionHosts(e.item.EmbedHosts, item.EmbedHosts)
}

func (i QueueItem) isStart() bool {
	return i.Depth == 0 && !i.FromSitemap
}

// sectionOf is the first path segment of a template key, or the first two
// when the first is a locale, so /en/blog and /en/docs stay apart.
func sectionOf(key string) string {
	u, err := url.Parse(key)
	if err != nil {
		return ""
	}

	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	if len(parts) > 1 && localeSegment.MatchString(parts[0]) {
		return parts[0] + "/" + parts[1]
	}

	return parts[0]
}

func unionHosts(current, added []string) []string {
	if len(added) == 0 {
		return current
	}

	out := slices.Clone(current)
	for _, host := range added {
		if !slices.Contains(out, host) {
			out = append(out, host)
		}
	}

	sort.Strings(out)

	return out
}
