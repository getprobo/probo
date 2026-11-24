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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestAccessSourceSyncRunStatus(t *testing.T) {
	t.Parallel()

	assert.True(t, coredata.AccessSourceSyncRunStatusQueued.IsValid())
	assert.True(t, coredata.AccessSourceSyncRunStatusRunning.IsValid())
	assert.True(t, coredata.AccessSourceSyncRunStatusSuccess.IsValid())
	assert.True(t, coredata.AccessSourceSyncRunStatusFailed.IsValid())
	assert.False(t, coredata.AccessSourceSyncRunStatus("BOGUS").IsValid())

	assert.False(t, coredata.AccessSourceSyncRunStatusQueued.IsTerminal())
	assert.False(t, coredata.AccessSourceSyncRunStatusRunning.IsTerminal())
	assert.True(t, coredata.AccessSourceSyncRunStatusSuccess.IsTerminal())
	assert.True(t, coredata.AccessSourceSyncRunStatusFailed.IsTerminal())

	var status coredata.AccessSourceSyncRunStatus

	require.NoError(t, status.UnmarshalText([]byte("RUNNING")))
	assert.Equal(t, coredata.AccessSourceSyncRunStatusRunning, status)

	require.Error(t, status.UnmarshalText([]byte("BOGUS")))
}
