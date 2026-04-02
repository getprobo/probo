// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package coredata

import "testing"

func TestAccessReviewCampaignStatusScan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    AccessReviewCampaignStatus
		wantErr bool
	}{
		{name: "draft string", input: "DRAFT", want: AccessReviewCampaignStatusDraft},
		{name: "in_progress string", input: "IN_PROGRESS", want: AccessReviewCampaignStatusInProgress},
		{name: "pending_actions string", input: "PENDING_ACTIONS", want: AccessReviewCampaignStatusPendingActions},
		{name: "failed string", input: "FAILED", want: AccessReviewCampaignStatusFailed},
		{name: "completed string", input: "COMPLETED", want: AccessReviewCampaignStatusCompleted},
		{name: "cancelled bytes", input: []byte("CANCELLED"), want: AccessReviewCampaignStatusCancelled},
		{name: "invalid value", input: "BOGUS", wantErr: true},
		{name: "unsupported type", input: 42, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got AccessReviewCampaignStatus
			err := got.Scan(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Scan(%v) expected error", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("Scan(%v) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("Scan(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAccessReviewCampaignStatusValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status AccessReviewCampaignStatus
		want   string
	}{
		{name: "draft", status: AccessReviewCampaignStatusDraft, want: "DRAFT"},
		{name: "in_progress", status: AccessReviewCampaignStatusInProgress, want: "IN_PROGRESS"},
		{name: "completed", status: AccessReviewCampaignStatusCompleted, want: "COMPLETED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.status.Value()
			if err != nil {
				t.Fatalf("Value() returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Value() = %q, want %q", got, tt.want)
			}
		})
	}
}
