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

package drivers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCrispSubscriptionSettings(t *testing.T) {
	t.Parallel()

	const (
		websiteID = "e8592878-c0d0-4632-b2f7-7d882f288d43"
		pluginID  = "e979a1c3-2c41-4e93-a8ed-410ace27318e"
	)

	t.Run("200 returns the verification code from data.settings", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The plugin subscription-settings endpoint is plugins (plural)
			// with both website_id and plugin_id in the path.
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/plugins/subscription/"+websiteID+"/"+pluginID+"/settings", r.URL.Path)
			assert.Equal(t, "plugin", r.Header.Get("X-Crisp-Tier"))

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"error":false,"reason":"resolved","data":{"plugin_id":"` + pluginID + `","settings":{"probo_verification_code":"ABC234DEF567"}}}`))
		}))
		defer srv.Close()

		client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

		settings, err := GetCrispSubscriptionSettings(context.Background(), client, websiteID, pluginID)
		require.NoError(t, err)
		require.NotNil(t, settings)
		assert.Equal(t, "ABC234DEF567", settings.ProboVerificationCode)
	})

	t.Run("200 without the code returns empty (mismatch handled by caller)", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"error":false,"data":{"settings":{}}}`))
		}))
		defer srv.Close()

		client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

		settings, err := GetCrispSubscriptionSettings(context.Background(), client, websiteID, pluginID)
		require.NoError(t, err)
		require.NotNil(t, settings)
		assert.Empty(t, settings.ProboVerificationCode)
	})

	t.Run("404 reports the plugin is not subscribed", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":true,"reason":"subscription_not_found"}`))
		}))
		defer srv.Close()

		client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

		settings, err := GetCrispSubscriptionSettings(context.Background(), client, websiteID, pluginID)
		require.ErrorIs(t, err, ErrCrispPluginNotSubscribed)
		assert.Nil(t, settings)
	})

	t.Run("non-2xx returns an error", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":true,"reason":"server_error"}`))
		}))
		defer srv.Close()

		client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

		settings, err := GetCrispSubscriptionSettings(context.Background(), client, websiteID, pluginID)
		require.Error(t, err)
		assert.False(t, errors.Is(err, ErrCrispPluginNotSubscribed))
		assert.Nil(t, settings)
	})

	t.Run("2xx with error:true returns an error", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"error":true,"reason":"route_forbidden","data":{}}`))
		}))
		defer srv.Close()

		client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

		settings, err := GetCrispSubscriptionSettings(context.Background(), client, websiteID, pluginID)
		require.Error(t, err)
		assert.Nil(t, settings)
	})
}
