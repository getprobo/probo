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

package mailman

import (
	"fmt"
	"time"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/statelesstoken"
)

const (
	TokenTypeUnsubscribe = "mailing_list_unsubscribe"
	TokenTypeConfirm     = "mailing_list_confirm_subscription"

	unsubscribeTokenExpiry = 365 * 24 * time.Hour
	confirmTokenExpiry     = 15 * 24 * time.Hour
)

type UnsubscribeTokenData struct {
	MailingListID string `json:"m"`
	Email         string `json:"e"`
}

func newUnsubscribeToken(secret string, mailingListID gid.GID, recipientEmail mail.Addr) (string, error) {
	return statelesstoken.NewToken(
		secret,
		TokenTypeUnsubscribe,
		unsubscribeTokenExpiry,
		UnsubscribeTokenData{
			MailingListID: mailingListID.String(),
			Email:         recipientEmail.String(),
		},
	)
}

func ValidateUnsubscribeToken(secret, tokenString string) (*UnsubscribeTokenData, error) {
	payload, err := statelesstoken.ValidateToken[UnsubscribeTokenData](secret, TokenTypeUnsubscribe, tokenString)
	if err != nil {
		return nil, fmt.Errorf("cannot validate unsubscribe token: %w", err)
	}

	return &payload.Data, nil
}

type ConfirmTokenData struct {
	MailingListID string `json:"m"`
	Email         string `json:"e"`
}

func newConfirmToken(secret string, mailingListID gid.GID, recipientEmail mail.Addr) (string, error) {
	return statelesstoken.NewToken(
		secret,
		TokenTypeConfirm,
		confirmTokenExpiry,
		ConfirmTokenData{
			MailingListID: mailingListID.String(),
			Email:         recipientEmail.String(),
		},
	)
}

// ValidateConfirmToken validates a subscription confirmation token and returns
// the embedded payload. Exported so the HTTP handler can use it.
func ValidateConfirmToken(secret, tokenString string) (*ConfirmTokenData, error) {
	payload, err := statelesstoken.ValidateToken[ConfirmTokenData](secret, TokenTypeConfirm, tokenString)
	if err != nil {
		return nil, fmt.Errorf("cannot validate confirm token: %w", err)
	}

	return &payload.Data, nil
}
