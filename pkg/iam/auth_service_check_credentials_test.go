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

package iam

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/passwdhash"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

func TestCheckCredentials_PasswordlessIdentityMatchesUnknownEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	hp, err := passwdhash.NewProfile(bytes32(t), 600000)
	require.NoError(t, err)

	svc := AuthService{
		Service: &Service{
			pg:        client,
			hp:        hp,
			dummyHash: mustHashDummy(hp),
		},
	}

	passwordlessEmail := uniqueEmail(t, "passwordless")
	insertIdentity(t, ctx, client, passwordlessEmail, nil)

	password := "TestPassword123!"
	_, passwordlessErr := svc.CheckCredentials(ctx, passwordlessEmail, password)
	_, unknownErr := svc.CheckCredentials(ctx, uniqueEmail(t, "unknown"), password)

	assertInvalidCredentials(t, passwordlessErr)
	assertInvalidCredentials(t, unknownErr)
	assert.Equal(t, unknownErr.Error(), passwordlessErr.Error())
}

func TestCheckCredentials_WrongPasswordStaysInvalidCredentials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	hp, err := passwdhash.NewProfile(bytes32(t), 600000)
	require.NoError(t, err)

	svc := AuthService{
		Service: &Service{
			pg:        client,
			hp:        hp,
			dummyHash: mustHashDummy(hp),
		},
	}

	password := "TestPassword123!"
	hashed, err := hp.HashPassword([]byte(password))
	require.NoError(t, err)

	email := uniqueEmail(t, "password")
	insertIdentity(t, ctx, client, email, hashed)

	identity, err := svc.CheckCredentials(ctx, email, password)
	require.NoError(t, err)
	require.NotNil(t, identity)
	assert.Equal(t, email, identity.EmailAddress)

	_, err = svc.CheckCredentials(ctx, email, "OtherPassword123!")
	assertInvalidCredentials(t, err)
}

func assertInvalidCredentials(t *testing.T, err error) {
	t.Helper()

	var creds *ErrInvalidCredentials
	require.ErrorAs(t, err, &creds)
	assert.Equal(t, "invalid email or password", err.Error())
}

func insertIdentity(t *testing.T, ctx context.Context, client *pg.Client, email mail.Addr, hashedPassword []byte) {
	t.Helper()

	now := time.Now()
	identity := &coredata.Identity{
		ID:             gid.New(gid.NilTenant, coredata.IdentityEntityType),
		EmailAddress:   email,
		FullName:       "Credential Test",
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := identity.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert identity: %w", err)
		}

		return nil
	}))
}

func uniqueEmail(t *testing.T, label string) mail.Addr {
	t.Helper()

	var buf [8]byte
	_, err := rand.Read(buf[:])
	require.NoError(t, err)

	email, err := mail.ParseAddr(fmt.Sprintf("%s-%s@example.com", label, hex.EncodeToString(buf[:])))
	require.NoError(t, err)

	return email
}

func bytes32(t *testing.T) []byte {
	t.Helper()

	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	require.NoError(t, err)

	return buf
}
