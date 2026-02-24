package probo

import "testing"

func TestNormalizeAccountKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		email      string
		externalID string
		want       string
	}{
		{
			name:       "email only",
			email:      "  Jane@Example.com ",
			externalID: "",
			want:       "jane@example.com",
		},
		{
			name:       "email and external id",
			email:      "Jane@Example.com",
			externalID: " 123 ",
			want:       "jane@example.com|123",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeAccountKey(tt.email, tt.externalID)
			if got != tt.want {
				t.Fatalf("normalizeAccountKey(%q, %q) = %q, want %q", tt.email, tt.externalID, got, tt.want)
			}
		})
	}
}
