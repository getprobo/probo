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

func TestEqualTo(t *testing.T) {
	t.Run("equal strings", func(t *testing.T) {
		str1 := "password"
		str2 := "password"
		err := EqualTo(&str2)(&str1)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("different strings", func(t *testing.T) {
		str1 := "password"
		str2 := "different"
		err := EqualTo(&str2)(&str1)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("equal integers", func(t *testing.T) {
		num1 := 42
		num2 := 42
		err := EqualTo(&num2)(&num1)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("different integers", func(t *testing.T) {
		num1 := 42
		num2 := 43
		err := EqualTo(&num2)(&num1)
		if err == nil {
			t.Error("expected validation error")
		}
	})
}

func TestNotEqualTo(t *testing.T) {
	t.Run("different strings", func(t *testing.T) {
		str1 := "password"
		str2 := "different"
		err := NotEqualTo(&str2)(&str1)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("equal strings", func(t *testing.T) {
		str1 := "password"
		str2 := "password"
		err := NotEqualTo(&str2)(&str1)
		if err == nil {
			t.Error("expected validation error")
		}
	})
}

func TestEqualTo_TimeComparison(t *testing.T) {
	t.Run("same instant same location", func(t *testing.T) {
		time1 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		time2 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		err := EqualTo(time2)(time1)
		if err != nil {
			t.Errorf("expected no error for same instant, got: %v", err)
		}
	})

	t.Run("same instant different location", func(t *testing.T) {
		// Create the same instant in different time zones
		utcTime := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		est, _ := time.LoadLocation("America/New_York")
		estTime := time.Date(2025, 11, 5, 7, 0, 0, 0, est) // 7am EST = 12pm UTC

		err := EqualTo(utcTime)(estTime)
		if err != nil {
			t.Errorf("expected no error for same instant in different locations, got: %v", err)
		}
	})

	t.Run("different instants same location", func(t *testing.T) {
		time1 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		time2 := time.Date(2025, 11, 5, 13, 0, 0, 0, time.UTC)
		err := EqualTo(time2)(time1)
		if err == nil {
			t.Error("expected validation error for different instants")
		}
	})

	t.Run("pointer to time same instant", func(t *testing.T) {
		time1 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		time2 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		err := EqualTo(&time2)(&time1)
		if err != nil {
			t.Errorf("expected no error for pointer to same instant, got: %v", err)
		}
	})

	t.Run("nil time pointers", func(t *testing.T) {
		var time1 *time.Time
		var time2 *time.Time
		err := EqualTo(time2)(time1)
		if err != nil {
			t.Errorf("expected no error for nil time pointers, got: %v", err)
		}
	})

	t.Run("one nil one non-nil time pointer", func(t *testing.T) {
		var time1 *time.Time
		time2 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		err := EqualTo(&time2)(time1)
		if err == nil {
			t.Error("expected validation error for nil vs non-nil time")
		}
	})

	t.Run("same instant with monotonic clock difference", func(t *testing.T) {
		// Simulate times with different monotonic clock data
		baseTime := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		time1 := baseTime
		time.Sleep(1 * time.Millisecond) // Advances monotonic clock
		time2 := baseTime

		// Even though monotonic clocks differ, the instants are the same
		err := EqualTo(time2)(time1)
		if err != nil {
			t.Errorf("expected no error despite monotonic clock difference, got: %v", err)
		}
	})
}

func TestNotEqualTo_TimeComparison(t *testing.T) {
	t.Run("different instants", func(t *testing.T) {
		time1 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		time2 := time.Date(2025, 11, 5, 13, 0, 0, 0, time.UTC)
		err := NotEqualTo(time2)(time1)
		if err != nil {
			t.Errorf("expected no error for different instants, got: %v", err)
		}
	})

	t.Run("same instant same location", func(t *testing.T) {
		time1 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		time2 := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		err := NotEqualTo(time2)(time1)
		if err == nil {
			t.Error("expected validation error for same instant")
		}
	})

	t.Run("same instant different location", func(t *testing.T) {
		utcTime := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
		est, _ := time.LoadLocation("America/New_York")
		estTime := time.Date(2025, 11, 5, 7, 0, 0, 0, est)

		err := NotEqualTo(utcTime)(estTime)
		if err == nil {
			t.Error("expected validation error for same instant in different locations")
		}
	})
}
