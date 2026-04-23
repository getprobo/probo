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

package domainconnect

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"

	"go.gearno.de/kit/httpclient"
	"golang.org/x/net/publicsuffix"
)

var (
	ErrNotSupported     = errors.New("domain does not support Domain Connect")
	ErrTemplateNotFound = errors.New("DNS provider does not support the requested template")
	ErrPermissionDenied = errors.New("permission denied by DNS provider")
)

// Config holds the Domain Connect service provider configuration.
type Config struct {
	// ProviderID is the Domain Connect provider identifier (e.g. "probo.inc").
	ProviderID string

	// ServiceID is the Domain Connect service identifier (e.g. "customdomain").
	ServiceID string

	// PrivateKey is the signing key used to sign synchronous apply URLs.
	PrivateKey crypto.Signer

	// KeyID identifies the public key published at
	// {KeyID}._domainconnect.{syncPubKeyDomain} for signature verification.
	KeyID string

	// CallbackURL is the URL the DNS provider redirects to after the user
	// approves or denies the template application.
	CallbackURL string
}

// Settings represents the Domain Connect settings returned by a DNS provider.
type Settings struct {
	ProviderName string `json:"providerName"`
	URLSyncUX    string `json:"urlSyncUX"`
	URLAPI       string `json:"urlAPI"`
}

// Discover performs Domain Connect discovery for the given domain.
//
// It queries the _domainconnect TXT record for the registrable domain, then
// fetches the provider settings from the well-known endpoint.
func Discover(ctx context.Context, domain string, resolverAddr string) (*Settings, error) {
	registrable, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return nil, fmt.Errorf("cannot extract registrable domain from %q: %w", domain, err)
	}

	txtHost := "_domainconnect." + registrable

	var resolver *net.Resolver
	if resolverAddr != "" {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", resolverAddr)
			},
		}
	} else {
		resolver = net.DefaultResolver
	}

	records, err := resolver.LookupTXT(ctx, txtHost)
	if err != nil {
		return nil, ErrNotSupported
	}

	if len(records) == 0 {
		return nil, ErrNotSupported
	}

	apiHost := records[0]

	settingsURL := fmt.Sprintf("https://%s/v2/domainTemplates/providers", apiHost)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, settingsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create settings request: %w", err)
	}

	client := httpclient.DefaultClient(httpclient.WithSSRFProtection())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch Domain Connect settings: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// The settings endpoint may not exist on all providers. Fall back to
	// constructing settings from the TXT record value which is the API host.
	settings := &Settings{
		URLSyncUX: "https://" + apiHost,
		URLAPI:    "https://" + apiHost,
	}

	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(settings); err != nil {
			// If we cannot decode, use the defaults from the TXT record.
			settings.URLSyncUX = "https://" + apiHost
			settings.URLAPI = "https://" + apiHost
		}
	}

	return settings, nil
}

// CheckTemplate verifies that the DNS provider supports the given template.
func CheckTemplate(ctx context.Context, apiURL string, providerID string, serviceID string) error {
	u := fmt.Sprintf(
		"%s/v2/domainTemplates/providers/%s/services/%s",
		strings.TrimRight(apiURL, "/"),
		url.PathEscape(providerID),
		url.PathEscape(serviceID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("cannot create template check request: %w", err)
	}

	client := httpclient.DefaultClient(httpclient.WithSSRFProtection())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot check template: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return ErrTemplateNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cannot check template: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// BuildApplyURL constructs the Domain Connect synchronous apply URL.
//
// The caller redirects the user's browser to this URL. The DNS provider shows
// a consent screen; on approval, it creates the DNS records and redirects to
// the callback URL.
func BuildApplyURL(
	cfg Config,
	syncUXURL string,
	domain string,
	host string,
	params map[string]string,
	redirectURI string,
) (string, error) {
	base := fmt.Sprintf(
		"%s/v2/domainTemplates/providers/%s/services/%s/apply",
		strings.TrimRight(syncUXURL, "/"),
		url.PathEscape(cfg.ProviderID),
		url.PathEscape(cfg.ServiceID),
	)

	q := url.Values{}
	q.Set("domain", domain)
	if host != "" {
		q.Set("host", host)
	}
	for k, v := range params {
		q.Set(k, v)
	}
	if redirectURI != "" {
		q.Set("redirect_uri", redirectURI)
	}
	q.Set("providerName", "Probo")

	queryString := q.Encode()

	if cfg.PrivateKey != nil && cfg.KeyID != "" {
		sig, err := signQueryString(queryString, cfg.PrivateKey)
		if err != nil {
			return "", fmt.Errorf("cannot sign apply URL: %w", err)
		}

		queryString += "&sig=" + url.QueryEscape(sig) + "&key=" + url.QueryEscape(cfg.KeyID)
	}

	return base + "?" + queryString, nil
}

// ExtractHostAndDomain splits a fully qualified domain name into the host
// (subdomain) and the registrable domain parts.
//
// For example, "trust.example.com" returns host="trust", domain="example.com".
// For "example.com" it returns host="", domain="example.com".
func ExtractHostAndDomain(fqdn string) (host string, domain string, err error) {
	registrable, err := publicsuffix.EffectiveTLDPlusOne(fqdn)
	if err != nil {
		return "", "", fmt.Errorf("cannot extract registrable domain from %q: %w", fqdn, err)
	}

	if fqdn == registrable {
		return "", registrable, nil
	}

	prefix := strings.TrimSuffix(fqdn, "."+registrable)
	return prefix, registrable, nil
}

func signQueryString(queryString string, key crypto.Signer) (string, error) {
	hash := sha256.Sum256([]byte(queryString))

	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		r, s, err := ecdsa.Sign(rand.Reader, k, hash[:])
		if err != nil {
			return "", fmt.Errorf("cannot sign with ECDSA: %w", err)
		}

		curveBits := k.Curve.Params().BitSize
		keyBytes := (curveBits + 7) / 8

		rBytes := r.Bytes()
		sBytes := s.Bytes()

		sig := make([]byte, 2*keyBytes)
		copy(sig[keyBytes-len(rBytes):keyBytes], rBytes)
		copy(sig[2*keyBytes-len(sBytes):], sBytes)

		return base64.RawURLEncoding.EncodeToString(sig), nil

	case *rsa.PrivateKey:
		sig, err := rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, hash[:])
		if err != nil {
			return "", fmt.Errorf("cannot sign with RSA: %w", err)
		}

		return base64.RawURLEncoding.EncodeToString(sig), nil

	default:
		return "", fmt.Errorf("unsupported key type %T", key)
	}
}

// Enabled returns true when the Domain Connect configuration is complete
// enough to use.
func (c Config) Enabled() bool {
	return c.ProviderID != "" && c.ServiceID != ""
}

// VerifySignature verifies the signature in a Domain Connect signed URL.
// This is used only for testing; DNS providers do their own verification.
func VerifySignature(queryString string, sig string, pub crypto.PublicKey) error {
	hash := sha256.Sum256([]byte(queryString))

	sigBytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return fmt.Errorf("cannot decode signature: %w", err)
	}

	switch k := pub.(type) {
	case *ecdsa.PublicKey:
		curveBits := k.Curve.Params().BitSize
		keyBytes := (curveBits + 7) / 8

		if len(sigBytes) != 2*keyBytes {
			return fmt.Errorf("invalid ECDSA signature length")
		}

		r := new(big.Int).SetBytes(sigBytes[:keyBytes])
		s := new(big.Int).SetBytes(sigBytes[keyBytes:])

		if !ecdsa.Verify(k, hash[:], r, s) {
			return fmt.Errorf("ECDSA signature verification failed")
		}

		return nil

	case *rsa.PublicKey:
		return rsa.VerifyPKCS1v15(k, crypto.SHA256, hash[:], sigBytes)

	default:
		return fmt.Errorf("unsupported key type %T", pub)
	}
}

// NewECDSATestKey generates an ECDSA P-256 key pair for testing.
func NewECDSATestKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}
