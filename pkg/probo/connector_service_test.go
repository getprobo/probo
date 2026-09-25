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

package probo

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func TestEnableAccounts_RejectsUnsafeAccountFields(t *testing.T) {
	t.Parallel()

	longName := strings.Repeat("a", TitleMaxLength+1)

	tests := []struct {
		name    string
		account EnableConnectorAccount
		field   string
	}{
		{
			name: "empty external account id",
			account: EnableConnectorAccount{
				Name: "Acme",
			},
			field: "accounts[0].external_account_id",
		},
		{
			name: "name over max length",
			account: EnableConnectorAccount{
				ExternalAccountID: "123456789012",
				Name:              longName,
			},
			field: "accounts[0].name",
		},
		{
			name: "html in name",
			account: EnableConnectorAccount{
				ExternalAccountID: "123456789012",
				Name:              "<script>alert(1)</script>",
			},
			field: "accounts[0].name",
		},
		{
			name: "newline in external account id",
			account: EnableConnectorAccount{
				ExternalAccountID: "123\n456",
				Name:              "Acme",
			},
			field: "accounts[0].external_account_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &ConnectorService{}
			connectorID := gid.New(gid.NewTenantID(), coredata.ConnectorEntityType)

			_, err := service.EnableAccounts(
				t.Context(),
				nil,
				connectorID,
				[]EnableConnectorAccount{tt.account},
			)
			require.Error(t, err)

			validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
			require.True(t, ok)
			assert.Contains(t, validationErrors.Fields(), tt.field)
		})
	}
}

func TestEnableAccounts_RejectsDuplicateExternalAccountIDs(t *testing.T) {
	t.Parallel()

	service := &ConnectorService{}
	connectorID := gid.New(gid.NewTenantID(), coredata.ConnectorEntityType)

	_, err := service.EnableAccounts(
		t.Context(),
		nil,
		connectorID,
		[]EnableConnectorAccount{
			{ExternalAccountID: "111", Name: "Prod"},
			{ExternalAccountID: "111", Name: "Production"},
		},
	)
	require.Error(t, err)

	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)

	duplicate := validationErrors.ByField("accounts.external_account_id")
	require.Len(t, duplicate, 1)
	assert.Equal(t, validator.ErrorCodeInvalidFormat, duplicate[0].Code)
}
