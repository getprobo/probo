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

package deviceagent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestLoadEnrollments_MigratesLegacyKey(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	deviceID := gid.New(gid.NewTenantID(), coredata.DeviceEntityType)

	require.NoError(t, SaveAPIKey(dir, "legacy-key"))
	require.NoError(t, SaveConfig(dir, &Config{
		ServerURL: "https://US.probo.com",
		DeviceID:  deviceID.String(),
	}))

	items, err := LoadEnrollments(dir)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "https://us.probo.com", items[0].ServerURL)
	assert.Equal(t, deviceID.String(), items[0].DeviceID)
	assert.Equal(t, "legacy-key", items[0].APIKey)

	_, err = LoadAPIKey(dir)
	require.ErrorIs(t, err, ErrKeyNotFound)

	cfg, err := LoadConfig(dir)
	require.NoError(t, err)
	assert.Empty(t, cfg.ServerURL)
	assert.Empty(t, cfg.DeviceID)
}

func TestAddEnrollment(t *testing.T) {
	t.Parallel()

	t.Run(
		"stores a second enrollment on another server",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			firstID := gid.New(gid.NewTenantID(), coredata.DeviceEntityType)
			secondID := gid.New(gid.NewTenantID(), coredata.DeviceEntityType)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/agent/v1/heartbeat", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")

				deviceID := firstID.String()
				if r.Header.Get("Authorization") == "Bearer second-key" {
					deviceID = secondID.String()
				}

				_, _ = w.Write([]byte(
					`{"device_id":"` + deviceID + `","heartbeat_interval_seconds":300,"posture_interval_seconds":3600}`,
				))
			}))
			t.Cleanup(srv.Close)

			agent := New(dir, "test", nil)
			_, err := agent.AddEnrollment(context.Background(), srv.URL, "first-key")
			require.NoError(t, err)

			resp, err := agent.AddEnrollment(context.Background(), srv.URL, "second-key")
			require.NoError(t, err)
			assert.Equal(t, secondID.String(), resp.DeviceID)

			items, err := LoadEnrollments(dir)
			require.NoError(t, err)
			require.Len(t, items, 2)
			assert.Equal(t, "first-key", items[0].APIKey)
			assert.Equal(t, firstID.String(), items[0].DeviceID)
			assert.Equal(t, "second-key", items[1].APIKey)
			assert.Equal(t, secondID.String(), items[1].DeviceID)

			enrolled, err := IsEnrolled(EnrollmentRunDir(dir))
			require.NoError(t, err)
			assert.True(t, enrolled)
		},
	)

	t.Run(
		"drops the key when the organization already has this machine",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "device hardware uuid already enrolled", http.StatusConflict)
			}))
			t.Cleanup(srv.Close)

			agent := New(dir, "test", nil)
			_, err := agent.AddEnrollment(context.Background(), srv.URL, "conflict-key")
			require.ErrorIs(t, err, ErrAlreadyEnrolledInOrganization)

			items, err := LoadEnrollments(dir)
			require.NoError(t, err)
			assert.Empty(t, items)

			enrolled, err := IsEnrolled(EnrollmentRunDir(dir))
			require.NoError(t, err)
			assert.False(t, enrolled)
		},
	)

	t.Run(
		"keeps the key when activation cannot reach the server",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
			}))
			t.Cleanup(srv.Close)

			agent := New(dir, "test", nil)
			_, err := agent.AddEnrollment(context.Background(), srv.URL, "retry-key")
			require.Error(t, err)
			assert.ErrorContains(t, err, "cannot activate device")

			items, err := LoadEnrollments(dir)
			require.NoError(t, err)
			require.Len(t, items, 1)
			assert.Equal(t, "retry-key", items[0].APIKey)
			assert.Empty(t, items[0].DeviceID)
		},
	)
}
