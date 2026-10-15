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

package itam

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func TestValidateDeviceIDs(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	deviceID := gid.New(tenantID, coredata.DeviceEntityType)

	t.Run("accepts device ids", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validateDeviceIDs([]gid.GID{deviceID}, maxDeviceBatchSize))
	})

	t.Run("reports every failure", func(t *testing.T) {
		t.Parallel()

		err := validateDeviceIDs(deviceIDsOverEveryLimit(maxDeviceBatchSize), maxDeviceBatchSize)

		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		require.NotEmpty(t, validationErrors.ByCode(validator.ErrorCodeOutOfRange))
		require.NotEmpty(t, validationErrors.ByCode(validator.ErrorCodeInvalidGID))
	})
}

func deviceIDsOverEveryLimit(max int) []gid.GID {
	tenantID := gid.NewTenantID()

	ids := make([]gid.GID, max+1)
	for i := range ids {
		ids[i] = gid.New(tenantID, coredata.DeviceEntityType)
	}

	ids[0] = gid.New(tenantID, coredata.FileEntityType)

	return ids
}
