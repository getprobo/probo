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

package coredata

import (
	"time"

	"go.probo.inc/probo/pkg/timespan"
)

type InternalControlOperatingFrequency struct {
	Mode     InternalControlOperatingMode `json:"mode"`
	Interval *timespan.TimeSpan           `json:"interval,omitempty"`
	Event    *string                      `json:"event,omitempty"`
}

// Columns splits an operating frequency into the stored mode, interval, and
// event. A nil frequency clears all three. Only the fields that belong to the
// mode are kept.
func (f *InternalControlOperatingFrequency) Columns() (
	*InternalControlOperatingMode,
	*timespan.TimeSpan,
	*string,
) {
	if f == nil {
		return nil, nil, nil
	}

	mode := f.Mode

	switch mode {
	case InternalControlOperatingModePeriodic:
		return &mode, f.Interval, nil
	case InternalControlOperatingModeEvent:
		return &mode, nil, f.Event
	default:
		return &mode, nil, nil
	}
}

// NextInternalControlDue returns the next evidence or test due instant for a
// cadence measured from from. An empty or non-positive duration has no schedule.
func NextInternalControlDue(from time.Time, cadence *timespan.TimeSpan) *time.Time {
	if cadence == nil || cadence.IsZero() {
		return nil
	}

	next := cadence.AddTo(from)
	if !next.After(from) {
		return nil
	}

	return &next
}
