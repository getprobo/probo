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

package sign

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	publishableKeyLivePfx = "pk_live_"
	publishableKeyTestPfx = "pk_test_"
	secretKeyLivePfx      = "sk_live_"
	secretKeyTestPfx      = "sk_test_"
)

func newPublishableKey(mode coredata.DocumentSigningConfigMode) (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	pfx := publishableKeyLivePfx
	if mode == coredata.DocumentSigningConfigModeTest {
		pfx = publishableKeyTestPfx
	}

	return pfx + hex.EncodeToString(b), nil
}

func newSecretKeyPlaintext(mode coredata.DocumentSigningConfigMode) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	pfx := secretKeyLivePfx
	if mode == coredata.DocumentSigningConfigModeTest {
		pfx = secretKeyTestPfx
	}

	return pfx + hex.EncodeToString(b), nil
}

func keyMode(key string) (coredata.DocumentSigningConfigMode, bool) {
	switch {
	case strings.HasPrefix(key, publishableKeyLivePfx),
		strings.HasPrefix(key, secretKeyLivePfx):
		return coredata.DocumentSigningConfigModeLive, true
	case strings.HasPrefix(key, publishableKeyTestPfx),
		strings.HasPrefix(key, secretKeyTestPfx):
		return coredata.DocumentSigningConfigModeTest, true
	}

	return "", false
}

func isPublishableKey(key string) bool {
	return strings.HasPrefix(key, publishableKeyLivePfx) ||
		strings.HasPrefix(key, publishableKeyTestPfx)
}

func isSecretKey(key string) bool {
	return strings.HasPrefix(key, secretKeyLivePfx) ||
		strings.HasPrefix(key, secretKeyTestPfx)
}
