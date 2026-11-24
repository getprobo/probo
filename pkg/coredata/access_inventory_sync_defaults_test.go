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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/timespan"
)

func TestAccessInventoryProviderKeyForSource(t *testing.T) {
	t.Parallel()

	github := coredata.ConnectorProviderGitHub

	assert.Equal(
		t,
		coredata.AccessInventoryProviderKeyFromConnector(github),
		coredata.AccessInventoryProviderKeyForSource(&github, true),
	)
	assert.Equal(
		t,
		coredata.AccessInventoryProviderKeyCSV,
		coredata.AccessInventoryProviderKeyForSource(nil, true),
	)
	assert.Equal(
		t,
		coredata.AccessInventoryProviderKeyProboMemberships,
		coredata.AccessInventoryProviderKeyForSource(nil, false),
	)
}

func TestAccessInventoryProviderKey(t *testing.T) {
	t.Parallel()

	assert.False(t, coredata.AccessInventoryProviderKey("BOGUS").IsValid())
	assert.True(t, coredata.AccessInventoryProviderKeyCSV.IsValid())
	assert.True(t, coredata.AccessInventoryProviderKeyProboMemberships.IsValid())
	assert.True(
		t,
		coredata.AccessInventoryProviderKeyFromConnector(coredata.ConnectorProviderGitHub).IsValid(),
	)

	var key coredata.AccessInventoryProviderKey

	require.NoError(t, key.UnmarshalText([]byte("CSV")))
	assert.Equal(t, coredata.AccessInventoryProviderKeyCSV, key)
	require.Error(t, key.UnmarshalText([]byte("BOGUS")))
}

func TestDefaultSyncCadence(t *testing.T) {
	t.Parallel()

	hourly, ok := coredata.DefaultSyncCadence(
		coredata.AccessInventoryProviderKeyFromConnector(coredata.ConnectorProviderGitHub),
	).ClockDuration()
	require.True(t, ok)
	assert.Equal(t, time.Hour, hourly)

	identity, ok := coredata.DefaultSyncCadence(
		coredata.AccessInventoryProviderKeyFromConnector(coredata.ConnectorProviderOkta),
	).ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 30*time.Minute, identity)

	authentik, ok := coredata.DefaultSyncCadence(
		coredata.AccessInventoryProviderKeyFromConnector(coredata.ConnectorProviderAuthentik),
	).ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 30*time.Minute, authentik)

	csv, ok := coredata.DefaultSyncCadence(coredata.AccessInventoryProviderKeyCSV).ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 24*time.Hour, csv)

	memberships, ok := coredata.DefaultSyncCadence(
		coredata.AccessInventoryProviderKeyProboMemberships,
	).ClockDuration()
	require.True(t, ok)
	assert.Equal(t, time.Hour, memberships)
}

func TestResolveSyncCadence(t *testing.T) {
	t.Parallel()

	override := timespan.FromDuration(15 * time.Minute)
	resolved := coredata.ResolveSyncCadence(
		coredata.AccessInventoryProviderKeyFromConnector(coredata.ConnectorProviderGitHub),
		&override,
	)
	duration, ok := resolved.ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 15*time.Minute, duration)

	zero := timespan.TimeSpan{}
	fallback := coredata.ResolveSyncCadence(
		coredata.AccessInventoryProviderKeyCSV,
		&zero,
	)
	duration, ok = fallback.ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 24*time.Hour, duration)

	negative := timespan.FromDuration(-time.Hour)
	ignoredNegative := coredata.ResolveSyncCadence(
		coredata.AccessInventoryProviderKeyCSV,
		&negative,
	)
	duration, ok = ignoredNegative.ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 24*time.Hour, duration)

	defaulted := coredata.ResolveSyncCadence(
		coredata.AccessInventoryProviderKeyCSV,
		nil,
	)
	duration, ok = defaulted.ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 24*time.Hour, duration)
}

func TestResolveHideAfterMissingSyncs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.DefaultHideAfterMissingSyncs, coredata.ResolveHideAfterMissingSyncs(nil))
	assert.Equal(
		t,
		5,
		coredata.ResolveHideAfterMissingSyncs(
			&coredata.AccessInventorySettings{HideAfterMissingSyncs: 5},
		),
	)
}

func TestClampHideAfterMissingSyncs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.MinHideAfterMissingSyncs, coredata.ClampHideAfterMissingSyncs(0))
	assert.Equal(t, coredata.MaxHideAfterMissingSyncs, coredata.ClampHideAfterMissingSyncs(100))
	assert.Equal(t, 7, coredata.ClampHideAfterMissingSyncs(7))
}
