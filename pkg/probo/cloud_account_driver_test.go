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

package probo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCloudAccountRegistryRef_NilCloudAccounts asserts the helper
// returns nil when the TenantService has no CloudAccounts sub-service
// wired. The driver factory uses this to fail loud rather than panic.
func TestCloudAccountRegistryRef_NilCloudAccounts(t *testing.T) {
	t.Parallel()

	svc := &TenantService{}
	assert.Nil(t, svc.cloudAccountRegistryRef())
}

// TestCloudAccountRegistryRef_NilRegistry asserts the helper returns
// nil when the CloudAccounts sub-service is wired but its registry
// pointer is unset (a malformed config wiring).
func TestCloudAccountRegistryRef_NilRegistry(t *testing.T) {
	t.Parallel()

	svc := &TenantService{
		CloudAccounts: &CloudAccountService{},
	}
	assert.Nil(t, svc.cloudAccountRegistryRef())
}
