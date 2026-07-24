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

package domaindns

import (
	"context"
	"testing"

	"codeberg.org/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCNAME(t *testing.T) {
	t.Parallel()

	t.Run("accepts matching owner and target", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, msg *dns.Msg, _ string) (*dns.Msg, error) {
				cname := &dns.CNAME{Hdr: dns.Header{Name: msg.Question[0].Header().Name}}
				cname.Target = "custom.getprobo.com."

				return &dns.Msg{Answer: []dns.RR{cname}}, nil
			},
		}

		err := client.CheckCNAME(context.Background(), "trust.example.com", "custom.getprobo.com")

		require.NoError(t, err)
	})

	t.Run("rejects apex owned record for subdomain query", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, _ string) (*dns.Msg, error) {
				cname := &dns.CNAME{Hdr: dns.Header{Name: "example.com."}}
				cname.Target = "custom.getprobo.com."

				return &dns.Msg{Answer: []dns.RR{cname}}, nil
			},
		}

		err := client.CheckCNAME(context.Background(), "trust.example.com", "custom.getprobo.com")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cname owner mismatch")
	})
}

func TestCheckTXT(t *testing.T) {
	t.Parallel()

	t.Run("ignores parent apex txt", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, _ string) (*dns.Msg, error) {
				txt := &dns.TXT{Hdr: dns.Header{Name: "example.com."}}
				txt.Txt = []string{"probo-verification=token"}

				return &dns.Msg{Answer: []dns.RR{txt}}, nil
			},
		}

		err := client.CheckTXT(context.Background(), "mail.example.com", "probo-verification=token")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTXTMismatch)
	})

	t.Run("accepts txt on exact domain", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, _ string) (*dns.Msg, error) {
				txt := &dns.TXT{Hdr: dns.Header{Name: "example.com."}}
				txt.Txt = []string{"probo-verification=token"}

				return &dns.Msg{Answer: []dns.RR{txt}}, nil
			},
		}

		err := client.CheckTXT(context.Background(), "example.com", "probo-verification=token")

		require.NoError(t, err)
	})
}

func TestCaaPermitsIssuer(t *testing.T) {
	t.Parallel()

	records := []*dns.CAA{{
		Hdr: dns.Header{Name: "example.com."},
	}}
	records[0].Tag = "issue"
	records[0].Value = "letsencrypt.org; accounturi=https://example.com"

	assert.True(t, caaPermitsIssuer(records, "letsencrypt.org"))
	assert.False(t, caaPermitsIssuer(records, "digicert.com"))
}
