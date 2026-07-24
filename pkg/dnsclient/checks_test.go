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

package dnsclient

import (
	"context"
	"testing"
	"time"

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

	t.Run("retries over tcp when udp response is truncated", func(t *testing.T) {
		t.Parallel()

		var networks []string

		client := &Client{
			exchange: func(_ context.Context, msg *dns.Msg, network string) (*dns.Msg, error) {
				networks = append(networks, network)
				if network == "udp" {
					return &dns.Msg{
						MsgHeader: dns.MsgHeader{
							Rcode:     dns.RcodeSuccess,
							Truncated: true,
						},
					}, nil
				}

				cname := &dns.CNAME{Hdr: dns.Header{Name: msg.Question[0].Header().Name}}
				cname.Target = "custom.getprobo.com."

				return &dns.Msg{Answer: []dns.RR{cname}}, nil
			},
		}

		err := client.CheckCNAME(context.Background(), "trust.example.com", "custom.getprobo.com")

		require.NoError(t, err)
		assert.Equal(t, []string{"udp", "tcp"}, networks)
	})

	t.Run("rejects response still truncated after tcp retry", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, _ string) (*dns.Msg, error) {
				cname := &dns.CNAME{Hdr: dns.Header{Name: "trust.example.com."}}
				cname.Target = "custom.getprobo.com."

				return &dns.Msg{
					MsgHeader: dns.MsgHeader{
						Rcode:     dns.RcodeSuccess,
						Truncated: true,
					},
					Answer: []dns.RR{cname},
				}, nil
			},
		}

		err := client.CheckCNAME(context.Background(), "trust.example.com", "custom.getprobo.com")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "truncated")
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

	t.Run("maps nxdomain to ErrTXTNotFound", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, _ string) (*dns.Msg, error) {
				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeNameError},
				}, nil
			},
		}

		err := client.CheckTXT(context.Background(), "mail.example.com", "probo-verification=token")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTXTNotFound)
	})

	t.Run("retries over tcp when udp response is truncated", func(t *testing.T) {
		t.Parallel()

		var networks []string

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, network string) (*dns.Msg, error) {
				networks = append(networks, network)
				if network == "udp" {
					return &dns.Msg{
						MsgHeader: dns.MsgHeader{
							Rcode:     dns.RcodeSuccess,
							Truncated: true,
						},
					}, nil
				}

				txt := &dns.TXT{Hdr: dns.Header{Name: "example.com."}}
				txt.Txt = []string{"probo-verification=token"}

				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess},
					Answer:    []dns.RR{txt},
				}, nil
			},
		}

		err := client.CheckTXT(context.Background(), "example.com", "probo-verification=token")

		require.NoError(t, err)
		assert.Equal(t, []string{"udp", "tcp"}, networks)
	})
}

func TestCaaPermitsIssuer(t *testing.T) {
	t.Parallel()

	t.Run("matching issue permits", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{caaRecord("issue", "letsencrypt.org; accounturi=https://example.com", 0)}

		assert.True(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})

	t.Run("non-matching issue denies", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{caaRecord("issue", "letsencrypt.org", 0)}

		assert.False(t, caaPermitsIssuer(records, "digicert.com"))
	})

	t.Run("empty issue value denies", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{caaRecord("issue", ";", 0)}

		assert.False(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})

	t.Run("issue tag is case insensitive", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{caaRecord("ISSUE", "LetsEncrypt.ORG", 0)}

		assert.True(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})

	t.Run("only issuewild permits non-wildcard", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{caaRecord("issuewild", "letsencrypt.org", 0)}

		assert.True(t, caaPermitsIssuer(records, "digicert.com"))
	})

	t.Run("iodef non-critical ignored with matching issue", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{
			caaRecord("iodef", "mailto:security@example.com", 0),
			caaRecord("issue", "letsencrypt.org", 0),
		}

		assert.True(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})

	t.Run("critical unknown tag denies even with matching issue", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{
			caaRecord("issue", "letsencrypt.org", 0),
			caaRecord("unknown", "value", 1),
		}

		assert.False(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})

	t.Run("critical issuewild is recognized and ignored for non-wildcard", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{
			caaRecord("issuewild", "other.ca", 1),
			caaRecord("issue", "letsencrypt.org", 0),
		}

		assert.True(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})

	t.Run("malformed issue value does not authorize issuer prefix", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{
			caaRecord("issue", "letsencrypt.org; accounturi", 0),
		}

		assert.False(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})

	t.Run("malformed issue value alone forbids issuance", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{caaRecord("issue", "%%%%%", 0)}

		assert.False(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})

	t.Run("valid issue alongside malformed still authorizes", func(t *testing.T) {
		t.Parallel()

		records := []*dns.CAA{
			caaRecord("issue", "%%%%%", 0),
			caaRecord("issue", "letsencrypt.org; accounturi=https://example.com", 0),
		}

		assert.True(t, caaPermitsIssuer(records, "letsencrypt.org"))
	})
}

func TestCheckCAA(t *testing.T) {
	t.Parallel()

	t.Run("returns error on servfail", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, _ string) (*dns.Msg, error) {
				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeServerFailure},
				}, nil
			},
		}

		err := client.CheckCAA(context.Background(), "trust.example.com", "letsencrypt.org")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "SERVFAIL")
		assert.NotErrorIs(t, err, ErrCAADenied)
	})

	t.Run("retries over tcp when udp response is truncated", func(t *testing.T) {
		t.Parallel()

		var networks []string

		client := &Client{
			exchange: func(_ context.Context, msg *dns.Msg, network string) (*dns.Msg, error) {
				networks = append(networks, network)
				if network == "udp" {
					return &dns.Msg{
						MsgHeader: dns.MsgHeader{
							Rcode:     dns.RcodeSuccess,
							Truncated: true,
						},
					}, nil
				}

				name := msg.Question[0].Header().Name
				caa := caaRecord("issue", "letsencrypt.org", 0)
				caa.Hdr.Name = name

				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess},
					Answer:    []dns.RR{caa},
				}, nil
			},
		}

		err := client.CheckCAA(context.Background(), "trust.example.com", "letsencrypt.org")

		require.NoError(t, err)
		assert.Equal(t, []string{"udp", "tcp"}, networks)
	})

	t.Run("returns error when truncated after tcp retry", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, _ string) (*dns.Msg, error) {
				return &dns.Msg{
					MsgHeader: dns.MsgHeader{
						Rcode:     dns.RcodeSuccess,
						Truncated: true,
					},
				}, nil
			},
		}

		err := client.CheckCAA(context.Background(), "trust.example.com", "letsencrypt.org")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "truncated")
	})

	t.Run("returns error on nxdomain", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, _ *dns.Msg, _ string) (*dns.Msg, error) {
				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeNameError},
				}, nil
			},
		}

		err := client.CheckCAA(context.Background(), "trust.example.com", "letsencrypt.org")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "NXDOMAIN")
	})

	t.Run("permits when first non-empty rrset allows issuer", func(t *testing.T) {
		t.Parallel()

		var queried []string

		client := &Client{
			exchange: func(_ context.Context, msg *dns.Msg, _ string) (*dns.Msg, error) {
				name := msg.Question[0].Header().Name
				queried = append(queried, name)

				if name != "example.com." {
					return &dns.Msg{MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess}}, nil
				}

				caa := caaRecord("issue", "letsencrypt.org", 0)
				caa.Hdr.Name = name

				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess},
					Answer:    []dns.RR{caa},
				}, nil
			},
		}

		err := client.CheckCAA(context.Background(), "trust.example.com", "letsencrypt.org")

		require.NoError(t, err)
		assert.Equal(t, []string{"trust.example.com.", "example.com."}, queried)
	})

	t.Run("denies when parent forbids after empty child", func(t *testing.T) {
		t.Parallel()

		var queried []string

		client := &Client{
			exchange: func(_ context.Context, msg *dns.Msg, _ string) (*dns.Msg, error) {
				name := msg.Question[0].Header().Name
				queried = append(queried, name)

				if name == "trust.example.com." {
					return &dns.Msg{MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess}}, nil
				}

				caa := caaRecord("issue", "digicert.com", 0)
				caa.Hdr.Name = name

				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess},
					Answer:    []dns.RR{caa},
				}, nil
			},
		}

		err := client.CheckCAA(context.Background(), "trust.example.com", "letsencrypt.org")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCAADenied)
		assert.Equal(t, []string{"trust.example.com.", "example.com."}, queried)
	})

	t.Run("denies when first non-empty rrset forbids issuer", func(t *testing.T) {
		t.Parallel()

		client := &Client{
			exchange: func(_ context.Context, msg *dns.Msg, _ string) (*dns.Msg, error) {
				name := msg.Question[0].Header().Name
				caa := caaRecord("issue", "digicert.com", 0)
				caa.Hdr.Name = name

				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess},
					Answer:    []dns.RR{caa},
				}, nil
			},
		}

		err := client.CheckCAA(context.Background(), "trust.example.com", "letsencrypt.org")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCAADenied)
	})

	t.Run("applies exchange timeout per label", func(t *testing.T) {
		t.Parallel()

		var deadlines []time.Time

		client := &Client{
			ExchangeTimeout: 2 * time.Second,
			exchange: func(ctx context.Context, msg *dns.Msg, _ string) (*dns.Msg, error) {
				deadline, ok := ctx.Deadline()
				require.True(t, ok)
				deadlines = append(deadlines, deadline)

				name := msg.Question[0].Header().Name
				if name != "example.com." {
					return &dns.Msg{MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess}}, nil
				}

				caa := caaRecord("issue", "letsencrypt.org", 0)
				caa.Hdr.Name = name

				return &dns.Msg{
					MsgHeader: dns.MsgHeader{Rcode: dns.RcodeSuccess},
					Answer:    []dns.RR{caa},
				}, nil
			},
		}

		err := client.CheckCAA(context.Background(), "trust.example.com", "letsencrypt.org")

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(deadlines), 2)
		assert.True(
			t,
			deadlines[1].After(deadlines[0]),
			"expected a fresh per-label deadline, got shared climb deadline",
		)
	})
}

func caaRecord(tag, value string, flag uint8) *dns.CAA {
	caa := &dns.CAA{Hdr: dns.Header{Name: "example.com."}}
	caa.Flag = flag
	caa.Tag = tag
	caa.Value = value

	return caa
}
