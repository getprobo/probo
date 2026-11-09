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

package richtext

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestValidateCommentContentJSON(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, coredata.MembershipProfileEntityType).String()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "empty", in: "", wantErr: false},
		{
			name:    "paragraph",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}`,
			wantErr: false,
		},
		{
			name: "mention",
			in: fmt.Sprintf(
				`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
				profileID,
			),
			wantErr: false,
		},
		{
			name: "mention wrong entity type",
			in: fmt.Sprintf(
				`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
				gid.New(gid.TenantID{}, coredata.DocumentEntityType).String(),
			),
			wantErr: true,
		},
		{
			name:    "heading rejected",
			in:      `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"hi"}]}]}`,
			wantErr: true,
		},
		{
			name:    "image rejected",
			in:      `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png"}}]}`,
			wantErr: true,
		},
		{
			name:    "underline rejected",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"underline"}],"text":"hi"}]}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateCommentContentJSON(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
