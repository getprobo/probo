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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestNewCustomDomain_IncludesDNSRecords(t *testing.T) {
	t.Parallel()

	domain := &coredata.CustomDomain{
		Domain:  "trust.example.com",
		Managed: false,
	}

	got := NewCustomDomain(domain, nil, "custom.test.getprobo.com")

	require.Len(t, got.DNSRecords, 1)
	assert.Equal(t, "CNAME", got.DNSRecords[0].Type)
	assert.Equal(t, "trust.example.com", got.DNSRecords[0].Name)
	assert.Equal(t, "custom.test.getprobo.com", got.DNSRecords[0].Value)
	assert.Equal(t, 300, got.DNSRecords[0].TTL)
	assert.Equal(t, "Point domain to Probo servers", got.DNSRecords[0].Purpose)
	assert.Nil(t, got.Certificate)
	assert.False(t, got.Managed)
}
