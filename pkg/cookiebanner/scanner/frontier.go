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
	"sort"
	"time"
)

type frontier struct {
	visited  map[string]struct{}
	queued   map[string]struct{}
	next     []QueueItem
	later    []QueueItem
	maxQueue int
}

func newFrontier(start Checkpoint, maxQueue int) frontier {
	f := frontier{
		visited:  map[string]struct{}{},
		queued:   map[string]struct{}{},
		maxQueue: maxQueue,
	}
	for _, key := range start.VisitedTemplates {
		f.visited[key] = struct{}{}
	}

	for _, item := range start.Queue {
		f.push(item)
	}

	return f
}

func (f *frontier) push(item QueueItem) {
	if f.maxQueue > 0 && len(f.next)+len(f.later) >= f.maxQueue {
		return
	}

	f.insert(item)
}

// requeue puts an interrupted page back. It ignores the queue cap so the
// page is not dropped.
func (f *frontier) requeue(item QueueItem) {
	f.insert(item)
}

func (f *frontier) insert(item QueueItem) {
	key, ok := templateKey(item.URL)
	if !ok {
		return
	}

	if _, seen := f.visited[key]; seen {
		return
	}

	if _, waiting := f.queued[key]; waiting {
		return
	}

	f.queued[key] = struct{}{}
	if item.FromSitemap {
		f.later = append(f.later, item)
	} else {
		f.next = append(f.next, item)
	}
}

func (f *frontier) pop() (QueueItem, bool) {
	for {
		var (
			item QueueItem
			ok   bool
		)
		if len(f.next) > 0 {
			item, f.next = f.next[0], f.next[1:]
			ok = true
		} else if len(f.later) > 0 {
			item, f.later = f.later[0], f.later[1:]
			ok = true
		}

		if !ok {
			return QueueItem{}, false
		}

		key, valid := templateKey(item.URL)
		if !valid {
			continue
		}

		delete(f.queued, key)

		if _, seen := f.visited[key]; seen {
			continue
		}

		return item, true
	}
}

func (f *frontier) markVisited(raw string) {
	key, ok := templateKey(raw)
	if !ok {
		return
	}

	f.visited[key] = struct{}{}
	delete(f.queued, key)
}

func (f *frontier) checkpoint(pages int, elapsed time.Duration, truncated bool) Checkpoint {
	visited := make([]string, 0, len(f.visited))
	for key := range f.visited {
		visited = append(visited, key)
	}

	sort.Strings(visited)

	queue := make([]QueueItem, 0, len(f.next)+len(f.later))
	queue = append(queue, f.next...)
	queue = append(queue, f.later...)

	return Checkpoint{
		VisitedTemplates: visited,
		Queue:            queue,
		PagesScanned:     pages,
		ElapsedMS:        elapsed.Milliseconds(),
		Truncated:        truncated,
	}
}
