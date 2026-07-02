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

import "testing"

func TestAccessReviewCampaignStatusIsValid(t *testing.T) {
	t.Parallel()

	for _, value := range AccessReviewCampaignStatuses() {
		if !value.IsValid() {
			t.Fatalf("IsValid() = false for %q", value)
		}
	}

	if AccessReviewCampaignStatus("BOGUS").IsValid() {
		t.Fatal("IsValid() = true for invalid value")
	}
}

func TestAccessReviewCampaignStatusUnmarshalText(t *testing.T) {
	t.Parallel()

	for _, value := range AccessReviewCampaignStatuses() {
		t.Run(string(value), func(t *testing.T) {
			t.Parallel()

			var got AccessReviewCampaignStatus
			if err := got.UnmarshalText([]byte(value)); err != nil {
				t.Fatalf("UnmarshalText(%q) returned error: %v", value, err)
			}

			if got != value {
				t.Fatalf("UnmarshalText(%q) = %q, want %q", value, got, value)
			}
		})
	}

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		var got AccessReviewCampaignStatus
		if err := got.UnmarshalText([]byte("BOGUS")); err == nil {
			t.Fatal("UnmarshalText(BOGUS) expected error")
		}
	})
}

func TestAccessReviewCampaignStatusMarshalText(t *testing.T) {
	t.Parallel()

	for _, value := range AccessReviewCampaignStatuses() {
		t.Run(string(value), func(t *testing.T) {
			t.Parallel()

			got, err := value.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText() returned error: %v", err)
			}

			if string(got) != value.String() {
				t.Fatalf("MarshalText() = %q, want %q", string(got), value.String())
			}
		})
	}
}
