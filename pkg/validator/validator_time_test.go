// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package validator

import (
	"testing"
	"time"
)

func TestAfter(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	t.Run("time after reference", func(t *testing.T) {
		err := After(past)(&future)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("time before reference", func(t *testing.T) {
		err := After(future)(&past)
		if err == nil {
			t.Fatal("expected validation error")
		} else if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("same time", func(t *testing.T) {
		err := After(now)(&now)
		if err == nil {
			t.Error("expected validation error for equal times")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var timeVal *time.Time

		err := After(now)(timeVal)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestBefore(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	t.Run("time before reference", func(t *testing.T) {
		err := Before(future)(&past)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("time after reference", func(t *testing.T) {
		err := Before(past)(&future)
		if err == nil {
			t.Fatal("expected validation error")
		} else if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("same time", func(t *testing.T) {
		err := Before(now)(&now)
		if err == nil {
			t.Error("expected validation error for equal times")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var timeVal *time.Time

		err := Before(now)(timeVal)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestRangeDuration(t *testing.T) {
	minDuration := 10 * time.Minute
	maxDuration := 1 * time.Hour

	t.Run("duration within range", func(t *testing.T) {
		duration := 30 * time.Minute

		err := RangeDuration(minDuration, maxDuration)(&duration)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duration at minimum", func(t *testing.T) {
		duration := 10 * time.Minute

		err := RangeDuration(minDuration, maxDuration)(&duration)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duration at maximum", func(t *testing.T) {
		duration := 1 * time.Hour

		err := RangeDuration(minDuration, maxDuration)(&duration)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duration below minimum", func(t *testing.T) {
		duration := 5 * time.Minute

		err := RangeDuration(minDuration, maxDuration)(&duration)
		if err == nil {
			t.Fatal("expected validation error")
		} else if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("duration above maximum", func(t *testing.T) {
		duration := 2 * time.Hour

		err := RangeDuration(minDuration, maxDuration)(&duration)
		if err == nil {
			t.Fatal("expected validation error")
		} else if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var duration *time.Duration

		err := RangeDuration(minDuration, maxDuration)(duration)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}
