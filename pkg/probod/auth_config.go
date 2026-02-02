// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package probod

import (
	"encoding/base64"
	"fmt"
)

type (
	authConfig struct {
		Cookie                              cookieConfig   `json:"cookie"`
		Password                            passwordConfig `json:"password"`
		DisableSignup                       bool           `json:"disable-signup"`
		InvitationConfirmationTokenValidity int            `json:"invitation-confirmation-token-validity"`
		PasswordResetTokenValidity          int            `json:"password-reset-token-validity"`
		MagicLinkTokenValidity              int            `json:"magic-link-token-validity"`
		SAML                                samlConfig     `json:"saml"`
	}

	cookieConfig struct {
		Domain   string `json:"domain"`
		Secret   string `json:"secret"`
		Duration int    `json:"duration"`
		Name     string `json:"name"`
		Secure   bool   `json:"secure"`
	}

	passwordConfig struct {
		Iterations uint32 `json:"iterations"`
		Pepper     string `json:"pepper"`
	}
)

func (c authConfig) GetPepperBytes() ([]byte, error) {
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

func (c authConfig) GetCookieSecretBytes() ([]byte, error) {
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
