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

package changefilter

import (
	"fmt"
	"time"
)

// ChangeEventFilterArgs are the CLI inputs shared by list-change-events and
// list-grant-change-events.
type ChangeEventFilterArgs struct {
	SourceID             string
	ChangeTypes          []string
	ObservedAtOnOrAfter  string
	ObservedAtOnOrBefore string
}

// BuildChangeEventFilter validates observation-time flags and builds the
// GraphQL AccessInventoryChangeEventFilter map. An empty map means no filter.
func BuildChangeEventFilter(args ChangeEventFilterArgs) (map[string]any, error) {
	filter := map[string]any{}

	if args.SourceID != "" {
		filter["accessReviewSourceId"] = args.SourceID
	}

	if len(args.ChangeTypes) > 0 {
		filter["changeTypes"] = args.ChangeTypes
	}

	if args.ObservedAtOnOrAfter != "" {
		if _, err := time.Parse(time.RFC3339, args.ObservedAtOnOrAfter); err != nil {
			return nil, fmt.Errorf("invalid --observed-at-on-or-after %q: use RFC3339", args.ObservedAtOnOrAfter)
		}

		filter["observedAtOnOrAfter"] = args.ObservedAtOnOrAfter
	}

	if args.ObservedAtOnOrBefore != "" {
		if _, err := time.Parse(time.RFC3339, args.ObservedAtOnOrBefore); err != nil {
			return nil, fmt.Errorf("invalid --observed-at-on-or-before %q: use RFC3339", args.ObservedAtOnOrBefore)
		}

		filter["observedAtOnOrBefore"] = args.ObservedAtOnOrBefore
	}

	return filter, nil
}
