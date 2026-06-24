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

import (
	"encoding/base64"
	"fmt"
)

type AuthConfig struct {
	Cookie                              CookieConfig       `json:"cookie"`
	Password                            PasswordConfig     `json:"password"`
	DisableSignup                       bool               `json:"disable-signup"`
	InvitationConfirmationTokenValidity int                `json:"invitation-confirmation-token-validity"`
	PasswordResetTokenValidity          int                `json:"password-reset-token-validity"`
	MagicLinkTokenValidity              int                `json:"magic-link-token-validity"`
	SAML                                SAMLConfig         `json:"saml"`
	Google                              OIDCProviderConfig `json:"google,omitzero"`
	Microsoft                           OIDCProviderConfig `json:"microsoft,omitzero"`
	OAuth2Server                        OAuth2ServerConfig `json:"oauth2-server"`
}

type OAuth2ServerConfig struct {
	SigningKeys               []OAuth2SigningKeyConfig `json:"signing-keys"`
	AccessTokenDuration       int                      `json:"access-token-duration"`
	RefreshTokenDuration      int                      `json:"refresh-token-duration"`
	AuthorizationCodeDuration int                      `json:"authorization-code-duration"`
	DeviceCodeDuration        int                      `json:"device-code-duration"`
	CIMDAllowedClientIDs      []string                 `json:"cimd-allowed-client-ids,omitempty"`
}

type OAuth2SigningKeyConfig struct {
	PrivateKey string `json:"private-key"`
	KID        string `json:"kid"`
	Active     bool   `json:"active"`
}

type CookieConfig struct {
	Domain   string `json:"domain,omitempty"`
	Secret   string `json:"secret"`
	Duration int    `json:"duration"`
	Name     string `json:"name,omitempty"`
	Secure   bool   `json:"secure"`
}

type PasswordConfig struct {
	Iterations int    `json:"iterations"`
	Pepper     string `json:"pepper"`
}

func (c AuthConfig) GetPepperBytes() ([]byte, error) {
	if c.Password.Pepper == "" {
		return nil, fmt.Errorf("pepper cannot be empty")
	}

	if decoded, err := base64.StdEncoding.DecodeString(c.Password.Pepper); err == nil {
		if len(decoded) < 32 {
			return nil, fmt.Errorf("decoded pepper must be at least 32 bytes long")
		}

		return decoded, nil
	}

	if len(c.Password.Pepper) < 32 {
		return nil, fmt.Errorf("pepper must be at least 32 bytes long")
	}

	return []byte(c.Password.Pepper), nil
}

func (c AuthConfig) GetCookieSecretBytes() ([]byte, error) {
	if c.Cookie.Secret == "" {
		return nil, fmt.Errorf("cookie secret cannot be empty")
	}

	if decoded, err := base64.StdEncoding.DecodeString(c.Cookie.Secret); err == nil {
		if len(decoded) < 32 {
			return nil, fmt.Errorf("decoded cookie secret must be at least 32 bytes long")
		}

		return decoded, nil
	}

	if len(c.Cookie.Secret) < 32 {
		return nil, fmt.Errorf("cookie secret must be at least 32 bytes long")
	}

	return []byte(c.Cookie.Secret), nil
}
