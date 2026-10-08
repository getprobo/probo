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

package identityfederation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/crypto/jose"
)

// logAzureAssertion writes one decoded Azure assertion per process. The raw
// token and its signature never appear: a log line of this credential would
// be enough to exchange it at Entra until it expires.
func (i *Issuer) logAzureAssertion(ctx context.Context, token string) {
	if i.logger == nil {
		return
	}

	i.azureLogOnce.Do(
		func() {
			i.logger.InfoCtx(
				ctx,
				"minted azure identity federation assertion",
				azureAssertionLogFields(token, i.JWKS())...,
			)
		},
	)
}

// azureAssertionLogFields decodes a JWT's header and claims and reports
// whether jwks verifies the signature. A decode failure stops at
// decode_error so a malformed segment is never copied into the log.
func azureAssertionLogFields(token string, jwks *jose.JWKS) []log.Attr {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return []log.Attr{log.String("decode_error", "jwt does not have three segments")}
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return []log.Attr{log.String("decode_error", "cannot decode jwt header")}
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return []log.Attr{log.String("decode_error", "cannot decode jwt claims")}
	}

	var header jose.JWTHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return []log.Attr{log.String("decode_error", "cannot parse jwt header")}
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return []log.Attr{log.String("decode_error", "cannot parse jwt claims")}
	}

	fields := []log.Attr{
		log.String("alg", header.Algorithm),
		log.String("typ", header.Type),
		log.String("kid", header.KeyID),
		log.String("iss", claims.Issuer),
		log.String("sub", claims.Subject),
		log.String("aud", claims.Audience),
		log.Int64("iat", claims.IssuedAt),
		log.Int64("nbf", claims.NotBefore),
		log.Int64("exp", claims.ExpiresAt),
		log.String("jti", claims.JTI),
		log.Int("header_length", len(parts[0])),
		log.Int("payload_length", len(parts[1])),
		log.Int("signature_length", len(parts[2])),
	}

	if _, err := jose.VerifyJWTWithJWKS(token, jwks); err != nil {
		fields = append(
			fields,
			log.Bool("signature_verified", false),
			log.String("verify_error", err.Error()),
		)

		return fields
	}

	return append(fields, log.Bool("signature_verified", true))
}
