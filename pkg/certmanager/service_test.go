// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package certmanager_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

func TestServiceRun_ACMEDisabled(t *testing.T) {
	t.Parallel()

	service := certmanager.NewService(
		nil,
		nil,
		cipher.EncryptionKey{},
		certmanager.Config{},
		log.NewLogger(log.WithOutput(io.Discard)),
	)

	err := service.Run(context.Background())

	assert.EqualError(t, err, "cannot run certificate manager service without ACME")
}
