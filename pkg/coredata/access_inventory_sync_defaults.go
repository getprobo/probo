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

package coredata

import (
	"time"

	"go.probo.inc/probo/pkg/timespan"
)

const (
	// DefaultHideAfterMissingSyncs is the org default for UI hide-after-X.
	// Disappeared grants stay visible until consecutive_missing_syncs reaches
	// this threshold (decision #4). Three successful misses absorbs brief
	// provider flakiness without lingering for days at hourly cadence.
	DefaultHideAfterMissingSyncs = 3

	MinHideAfterMissingSyncs = 1
	MaxHideAfterMissingSyncs = 30
)

var (
	defaultSyncCadenceFallback = timespan.FromDuration(time.Hour)
	defaultSyncCadenceCSV      = timespan.FromDuration(24 * time.Hour)
	defaultSyncCadenceIdentity = timespan.FromDuration(30 * time.Minute)
)

// DefaultSyncCadence returns the built-in sync interval for a provider key.
// Identity-central providers sync every 30m; CSV polls daily; everything else
// (including Probo memberships) defaults to hourly. Org overrides live in
// access_inventory_provider_sync_configs and win via ResolveSyncCadence.
func DefaultSyncCadence(providerKey AccessInventoryProviderKey) timespan.TimeSpan {
	switch providerKey {
	case AccessInventoryProviderKeyCSV:
		return defaultSyncCadenceCSV
	case AccessInventoryProviderKeyProboMemberships:
		return defaultSyncCadenceFallback
	case
		AccessInventoryProviderKeyFromConnector(ConnectorProviderOkta),
		AccessInventoryProviderKeyFromConnector(ConnectorProviderGoogleWorkspace),
		AccessInventoryProviderKeyFromConnector(ConnectorProviderMicrosoft365),
		AccessInventoryProviderKeyFromConnector(ConnectorProviderAWS),
		AccessInventoryProviderKeyFromConnector(ConnectorProviderGCP),
		AccessInventoryProviderKeyFromConnector(ConnectorProviderAzure):
		return defaultSyncCadenceIdentity
	default:
		return defaultSyncCadenceFallback
	}
}

// ResolveSyncCadence prefers a non-zero org override, else the provider default.
func ResolveSyncCadence(
	providerKey AccessInventoryProviderKey,
	orgOverride *timespan.TimeSpan,
) timespan.TimeSpan {
	if orgOverride != nil && !orgOverride.IsZero() {
		return *orgOverride
	}

	return DefaultSyncCadence(providerKey)
}

// ResolveHideAfterMissingSyncs returns the configured X, or the product
// default when no settings row exists yet.
func ResolveHideAfterMissingSyncs(settings *AccessInventorySettings) int {
	if settings == nil {
		return DefaultHideAfterMissingSyncs
	}

	return settings.HideAfterMissingSyncs
}

// ClampHideAfterMissingSyncs bounds X to the supported range.
func ClampHideAfterMissingSyncs(value int) int {
	if value < MinHideAfterMissingSyncs {
		return MinHideAfterMissingSyncs
	}

	if value > MaxHideAfterMissingSyncs {
		return MaxHideAfterMissingSyncs
	}

	return value
}
