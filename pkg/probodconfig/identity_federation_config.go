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

package probodconfig

type (
	// IdentityFederationConfig configures the outbound OIDC issuer used to federate into
	// customer cloud accounts.
	IdentityFederationConfig struct {
		Enabled bool `json:"enabled"`
		// IssuerBaseURL is the base of the advertised issuer, for example
		// https://proboidentity.com. It defaults to {base-url}/federation, so
		// a self-hosted deployment needs no second domain.
		//
		// This value becomes immutable the moment a customer registers it with
		// their cloud provider: changing it requires every customer to redeploy
		// infrastructure they own.
		IssuerBaseURL string                               `json:"issuer-base-url,omitempty"`
		SigningKeys   []IdentityFederationSigningKeyConfig `json:"signing-keys,omitempty"`
	}

	// IdentityFederationSigningKeyConfig is one RSA key published in the identity federation
	// JWKS. Only active keys sign new tokens; retired keys stay listed so that
	// tokens a cloud provider already cached keep verifying.
	IdentityFederationSigningKeyConfig struct {
		PrivateKey RSAPrivateKey `json:"private-key"`
		KID        string        `json:"kid"`
		Active     bool          `json:"active"`
	}
)
