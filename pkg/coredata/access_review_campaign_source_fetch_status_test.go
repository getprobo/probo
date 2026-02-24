package coredata

import "testing"

func TestAccessReviewCampaignSourceFetchStatusIsTerminal(t *testing.T) {
	t.Parallel()

	if AccessReviewCampaignSourceFetchStatusQueued.IsTerminal() {
		t.Fatalf("QUEUED should not be terminal")
	}
	if AccessReviewCampaignSourceFetchStatusFetching.IsTerminal() {
		t.Fatalf("FETCHING should not be terminal")
	}
	if !AccessReviewCampaignSourceFetchStatusSuccess.IsTerminal() {
		t.Fatalf("SUCCESS should be terminal")
	}
	if !AccessReviewCampaignSourceFetchStatusFailed.IsTerminal() {
		t.Fatalf("FAILED should be terminal")
	}
}

func TestAccessReviewCampaignSourceFetchStatusScan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    AccessReviewCampaignSourceFetchStatus
		wantErr bool
	}{
		{
			name:  "queued string",
			input: "QUEUED",
			want:  AccessReviewCampaignSourceFetchStatusQueued,
		},
		{
			name:  "fetching bytes",
			input: []byte("FETCHING"),
			want:  AccessReviewCampaignSourceFetchStatusFetching,
		},
		{
			name:    "invalid value",
			input:   "BOGUS",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got AccessReviewCampaignSourceFetchStatus
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
