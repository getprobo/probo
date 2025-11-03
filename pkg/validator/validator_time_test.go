// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
		}
		if err.Code != ErrorCodeOutOfRange {
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
		}
		if err.Code != ErrorCodeOutOfRange {
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

func TestFutureDate(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	t.Run("future date", func(t *testing.T) {
		err := FutureDate()(&future)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("past date", func(t *testing.T) {
		err := FutureDate()(&past)
		if err == nil {
			t.Fatal("expected validation error")
		}
		if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var timeVal *time.Time
		err := FutureDate()(timeVal)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestPastDate(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	t.Run("past date", func(t *testing.T) {
		err := PastDate()(&past)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("future date", func(t *testing.T) {
		err := PastDate()(&future)
		if err == nil {
			t.Fatal("expected validation error")
		}
		if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var timeVal *time.Time
		err := PastDate()(timeVal)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestMinDuration(t *testing.T) {
	minDuration := 10 * time.Minute

	t.Run("duration above minimum", func(t *testing.T) {
		duration := 20 * time.Minute
		err := MinDuration(minDuration)(&duration)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duration equal to minimum", func(t *testing.T) {
		duration := 10 * time.Minute
		err := MinDuration(minDuration)(&duration)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duration below minimum", func(t *testing.T) {
		duration := 5 * time.Minute
		err := MinDuration(minDuration)(&duration)
		if err == nil {
			t.Fatal("expected validation error")
		}
		if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var duration *time.Duration
		err := MinDuration(minDuration)(duration)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestMaxDuration(t *testing.T) {
	maxDuration := 1 * time.Hour

	t.Run("duration below maximum", func(t *testing.T) {
		duration := 30 * time.Minute
		err := MaxDuration(maxDuration)(&duration)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duration equal to maximum", func(t *testing.T) {
		duration := 1 * time.Hour
		err := MaxDuration(maxDuration)(&duration)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duration above maximum", func(t *testing.T) {
		duration := 2 * time.Hour
		err := MaxDuration(maxDuration)(&duration)
		if err == nil {
			t.Fatal("expected validation error")
		}
		if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var duration *time.Duration
		err := MaxDuration(maxDuration)(duration)
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
		}
		if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("duration above maximum", func(t *testing.T) {
		duration := 2 * time.Hour
		err := RangeDuration(minDuration, maxDuration)(&duration)
		if err == nil {
			t.Fatal("expected validation error")
		}
		if err.Code != ErrorCodeOutOfRange {
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
