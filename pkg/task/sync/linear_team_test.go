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

package tasksync

import "testing"

func TestLinearTeamForCreate(t *testing.T) {
	t.Parallel()

	team := "team-1"
	blank := "  "

	tests := []struct {
		name        string
		explicitSet bool
		explicit    *string
		defaultID   string
		wantID      string
		wantPublish bool
	}{
		{
			name:        "omitted uses the default team",
			defaultID:   "default",
			wantID:      "default",
			wantPublish: true,
		},
		{
			name:      "omitted with no default publishes nothing",
			defaultID: "",
		},
		{
			name:        "null is no team",
			explicitSet: true,
			defaultID:   "default",
		},
		{
			name:        "blank is no team",
			explicitSet: true,
			explicit:    &blank,
			defaultID:   "default",
		},
		{
			name:        "explicit team ignores the default",
			explicitSet: true,
			explicit:    &team,
			defaultID:   "default",
			wantID:      "team-1",
			wantPublish: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotID, gotPublish := LinearTeamForCreate(tt.explicitSet, tt.explicit, tt.defaultID)
			if gotID != tt.wantID || gotPublish != tt.wantPublish {
				t.Fatalf("LinearTeamForCreate() = (%q, %v), want (%q, %v)", gotID, gotPublish, tt.wantID, tt.wantPublish)
			}
		})
	}
}
