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

package duration_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/duration"
)

func TestDurationUnmarshalJSONInt(t *testing.T) {
	var got duration.Duration
	err := json.Unmarshal([]byte(`60`), &got)
	require.NoError(t, err)

	assert.Equal(t, duration.Duration(60*time.Second), got)
}

func TestDurationUnmarshalJSONString(t *testing.T) {
	var got duration.Duration
	err := json.Unmarshal([]byte(`"5m"`), &got)
	require.NoError(t, err)

	assert.Equal(t, duration.Duration(5*time.Minute), got)
}

func TestDurationMarshalJSONRoundTrip(t *testing.T) {
	want := duration.Duration(90 * time.Second)

	data, err := json.Marshal(want)
	require.NoError(t, err)
	assert.Equal(t, `"1m30s"`, string(data))

	var got duration.Duration
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestDurationUnmarshalJSONInvalid(t *testing.T) {
	tests := []string{
		`true`,
		`1.5`,
		`"later"`,
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			var got duration.Duration
			err := json.Unmarshal([]byte(tt), &got)
			require.Error(t, err)
		})
	}
}
