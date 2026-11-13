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

package mdm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

const (
	// iruCassetteHost replaces the recording tenant's API host. The host
	// names the tenant, and cassettes are committed to a public repository.
	iruCassetteHost = "acme.api.kandji.io"

	// iruCassetteMaxDevices is how many devices a recording keeps. A few are
	// enough to pin the response shape; the rest of a fleet is not needed.
	iruCassetteMaxDevices = 3
)

// iruScrubbedFields are device fields that name a person, a machine or the
// tenant's own configuration. They are rewritten to synthetic values; every
// other field is kept as recorded, since the shape is what the cassette pins.
var iruScrubbedFields = map[string]func(idx int) any{
	"device_id":      func(idx int) any { return fmt.Sprintf("00000000-0000-4000-8000-%012d", idx+1) },
	"device_name":    func(idx int) any { return fmt.Sprintf("Device %d", idx+1) },
	"serial_number":  func(idx int) any { return fmt.Sprintf("SERIAL%06d", idx+1) },
	"asset_tag":      func(int) any { return "" },
	"blueprint_id":   func(int) any { return "00000000-0000-4000-8000-0000000000bb" },
	"blueprint_name": func(int) any { return "Blueprint" },
	"tags":           func(int) any { return []any{} },
	"mac_address":    func(int) any { return "00:00:5e:00:53:00" },
}

// sanitizeIruDevices rewrites a recorded device list so that it names no real
// tenant, person or machine. It runs on save, so a recording run asserts on
// live data and the replay asserts on what was committed.
func sanitizeIruDevices(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to save iru response with status %d", i.Response.Code)
	}

	u, err := url.Parse(i.Request.URL)
	if err != nil {
		return fmt.Errorf("cannot parse recorded iru url: %w", err)
	}

	u.Host = iruCassetteHost
	i.Request.URL = u.String()
	i.Request.Host = iruCassetteHost

	decoder := json.NewDecoder(strings.NewReader(i.Response.Body))
	decoder.UseNumber()

	var devices []map[string]any
	if err := decoder.Decode(&devices); err != nil {
		return fmt.Errorf("cannot decode recorded iru devices: %w", err)
	}

	if len(devices) > iruCassetteMaxDevices {
		devices = devices[:iruCassetteMaxDevices]
	}

	for idx, device := range devices {
		for field, replacement := range iruScrubbedFields {
			if _, ok := device[field]; ok {
				device[field] = replacement(idx)
			}
		}

		// An assigned user is an object; an unassigned one is "". Both shapes
		// are part of the contract, so only the object is rewritten.
		if user, ok := device["user"].(map[string]any); ok {
			user["email"] = fmt.Sprintf("member%d@example.com", idx+1)
			user["name"] = fmt.Sprintf("Member %d", idx+1)
			user["id"] = fmt.Sprintf("00000000-0000-4000-8000-0000000001%02d", idx+1)
		}
	}

	body, err := json.Marshal(devices)
	if err != nil {
		return fmt.Errorf("cannot encode sanitized iru devices: %w", err)
	}

	replaceCassetteBody(i, string(body))

	return nil
}

func TestIruDriver_Cassette(t *testing.T) {
	t.Parallel()

	token := os.Getenv("IRU_API_TOKEN")

	host := iruCassetteHost
	if token != "" {
		host = os.Getenv("IRU_API_HOST")
		require.NotEmpty(t, host, "IRU_API_HOST must be set to record")
	}

	rec := newRecorder(t, "testdata/iru", "IRU_API_TOKEN", sanitizeIruDevices)

	baseURL, err := IruBaseURL(host)
	require.NoError(t, err)

	records, err := NewIruDriver(newVCRClient(rec, token), baseURL).ListDevices(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	for _, record := range records {
		assert.NotEmpty(t, record.ExternalID)

		if record.UserEmail != nil && token == "" {
			assert.True(t, strings.HasSuffix(*record.UserEmail, "@example.com"))
		}
	}
}

// TestIruCassette_Sanitized guards the committed cassette itself: it must name
// only the synthetic host and synthetic people, and carry no credential.
func TestIruCassette_Sanitized(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/iru.yaml")
	if os.IsNotExist(err) {
		t.Skip("cassette not recorded")
	}

	require.NoError(t, err)

	content := string(data)

	assert.NotContains(t, strings.ToLower(content), "authorization")
	assert.NotContains(t, strings.ToLower(content), "bearer")

	hosts := regexp.MustCompile(`[a-z0-9-]+\.api\.(eu\.)?kandji\.io`).FindAllString(content, -1)
	for _, host := range hosts {
		assert.Equal(t, iruCassetteHost, host)
	}

	emails := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`).FindAllString(content, -1)
	for _, email := range emails {
		assert.True(t, strings.HasSuffix(email, "@example.com"), "unexpected email %q in cassette", email)
	}
}

func TestIruAPIHost(t *testing.T) {
	t.Parallel()

	t.Run(
		"canonicalizes pasted hosts and urls",
		func(t *testing.T) {
			t.Parallel()

			for raw, want := range map[string]string{
				"acme.api.kandji.io":              "acme.api.kandji.io",
				"  ACME.api.kandji.io  ":          "acme.api.kandji.io",
				"acme.api.kandji.io/":             "acme.api.kandji.io",
				"https://acme.api.kandji.io":      "acme.api.kandji.io",
				"https://acme.api.eu.kandji.io/":  "acme.api.eu.kandji.io",
				"https://acme-corp.api.kandji.io": "acme-corp.api.kandji.io",
			} {
				host, err := IruAPIHost(raw)
				require.NoError(t, err, raw)
				assert.Equal(t, want, host, raw)
			}
		},
	)

	t.Run(
		"refuses anything but an iru api host",
		func(t *testing.T) {
			t.Parallel()

			for _, raw := range []string{
				"",
				"acme",
				"acme.kandji.io",
				"api.kandji.io",
				"a.b.api.kandji.io",
				"-acme.api.kandji.io",
				"acme.api.kandji.io.evil.example",
				"evil.example/acme.api.kandji.io",
				"http://acme.api.kandji.io",
				"https://acme.api.kandji.io:8443",
				"https://user@acme.api.kandji.io",
				"https://acme.api.kandji.io/api/v1/devices",
				"https://acme.api.kandji.io/?x=1",
				"acme.api.us.kandji.io",
			} {
				_, err := IruAPIHost(raw)
				assert.Error(t, err, "expected %q to be refused", raw)
			}
		},
	)
}

func TestIruDevicesProbeURL(t *testing.T) {
	t.Parallel()

	probeURL, err := IruDevicesProbeURL("acme.api.eu.kandji.io")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.api.eu.kandji.io/api/v1/devices?limit=1&offset=0", probeURL)
}

func TestIruPlatform(t *testing.T) {
	t.Parallel()

	for platform, want := range map[string]coredata.DevicePlatform{
		"Mac":     coredata.DevicePlatformDarwin,
		"Windows": coredata.DevicePlatformWindows,
		"iPhone":  coredata.DevicePlatformIOS,
		"iPad":    coredata.DevicePlatformIPadOS,
		"AppleTV": coredata.DevicePlatformTVOS,
		"Android": coredata.DevicePlatformAndroid,
	} {
		got := iruPlatform(platform)
		require.NotNil(t, got, platform)
		assert.Equal(t, want, *got, platform)
	}

	// A platform Iru adds later reports none rather than a wrong one.
	assert.Nil(t, iruPlatform("Vision"))
	assert.Nil(t, iruPlatform(""))
}

// iruStub serves pages of devices by offset and records the query of every
// request it receives.
func iruStub(t *testing.T, pages map[string]string) (*httptest.Server, *[]url.Values) {
	t.Helper()

	var (
		mu      sync.Mutex
		queries []url.Values
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()

		queries = append(queries, r.URL.Query())
		mu.Unlock()

		if r.URL.Path != "/api/v1/devices" {
			http.NotFound(w, r)
			return
		}

		body, ok := pages[r.URL.Query().Get("offset")]
		if !ok {
			body = "[]"
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server, &queries
}

func iruDevicesJSON(t *testing.T, from, count int) string {
	t.Helper()

	devices := make([]map[string]any, count)
	for i := range count {
		devices[i] = map[string]any{
			"device_id": fmt.Sprintf("device-%d", from+i),
			"platform":  "Mac",
			"user":      "",
		}
	}

	body, err := json.Marshal(devices)
	require.NoError(t, err)

	return string(body)
}

func TestIruDriver_ListDevices(t *testing.T) {
	t.Parallel()

	t.Run(
		"pages until a short page",
		func(t *testing.T) {
			t.Parallel()

			server, queries := iruStub(
				t,
				map[string]string{
					"0":   iruDevicesJSON(t, 0, iruDevicesPageSize),
					"300": iruDevicesJSON(t, iruDevicesPageSize, 1),
				},
			)

			records, err := NewIruDriver(server.Client(), server.URL).ListDevices(context.Background())
			require.NoError(t, err)
			assert.Len(t, records, iruDevicesPageSize+1)
			assert.Equal(t, "device-300", records[iruDevicesPageSize].ExternalID)

			require.Len(t, *queries, 2)
			assert.Equal(t, "300", (*queries)[0].Get("limit"))
			assert.Equal(t, "0", (*queries)[0].Get("offset"))
			assert.Equal(t, "300", (*queries)[1].Get("offset"))
		},
	)

	t.Run(
		"stops after an exact final page",
		func(t *testing.T) {
			t.Parallel()

			server, queries := iruStub(
				t,
				map[string]string{"0": iruDevicesJSON(t, 0, iruDevicesPageSize)},
			)

			records, err := NewIruDriver(server.Client(), server.URL).ListDevices(context.Background())
			require.NoError(t, err)
			assert.Len(t, records, iruDevicesPageSize)
			assert.Len(t, *queries, 2)
		},
	)

	t.Run(
		"maps fields and skips removed devices",
		func(t *testing.T) {
			t.Parallel()

			server, _ := iruStub(
				t,
				map[string]string{
					"0": `[
						{
							"device_id": "d-1",
							"device_name": "Member's MacBook Air",
							"model": "MacBook Air (M1, 2020)",
							"serial_number": "C02XL0GHJGH5",
							"platform": "Mac",
							"os_version": "14.4.1",
							"last_check_in": "2024-07-23T14:11:37.150080Z",
							"user": {"email": "member@example.com", "name": "Member", "id": "u-1", "is_archived": false},
							"is_removed": false
						},
						{
							"device_id": "d-2",
							"device_name": "",
							"platform": "iPhone",
							"os_version": "17.5",
							"last_check_in": "",
							"user": ""
						},
						{"device_id": "d-3", "platform": "Mac", "user": "", "is_removed": true},
						{"device_id": "", "platform": "Mac", "user": ""}
					]`,
				},
			)

			records, err := NewIruDriver(server.Client(), server.URL).ListDevices(context.Background())
			require.NoError(t, err)
			require.Len(t, records, 2)

			mac := records[0]
			assert.Equal(t, "d-1", mac.ExternalID)
			require.NotNil(t, mac.Name)
			assert.Equal(t, "Member's MacBook Air", *mac.Name)
			require.NotNil(t, mac.Model)
			assert.Equal(t, "MacBook Air (M1, 2020)", *mac.Model)
			require.NotNil(t, mac.SerialNumber)
			assert.Equal(t, "C02XL0GHJGH5", *mac.SerialNumber)
			require.NotNil(t, mac.Platform)
			assert.Equal(t, coredata.DevicePlatformDarwin, *mac.Platform)
			require.NotNil(t, mac.OSVersion)
			assert.Equal(t, "14.4.1", *mac.OSVersion)
			require.NotNil(t, mac.UserEmail)
			assert.Equal(t, "member@example.com", *mac.UserEmail)
			require.NotNil(t, mac.LastCheckInAt)
			assert.Equal(t, time.Date(2024, 7, 23, 14, 11, 37, 150080000, time.UTC), mac.LastCheckInAt.UTC())

			phone := records[1]
			assert.Equal(t, "d-2", phone.ExternalID)
			assert.Nil(t, phone.Name)
			assert.Nil(t, phone.UserEmail)
			assert.Nil(t, phone.LastCheckInAt)
			require.NotNil(t, phone.Platform)
			assert.Equal(t, coredata.DevicePlatformIOS, *phone.Platform)
		},
	)

	t.Run(
		"fails without echoing an error body",
		func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"message":"Invalid authentication for acme","code":"invalid_auth"}`))
			}))
			t.Cleanup(server.Close)

			_, err := NewIruDriver(server.Client(), server.URL).ListDevices(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), "401")
			assert.NotContains(t, err.Error(), "acme")
		},
	)

	t.Run(
		"fails on a body that is not a device list",
		func(t *testing.T) {
			t.Parallel()

			server, _ := iruStub(t, map[string]string{"0": `{"detail": "nope"}`})

			_, err := NewIruDriver(server.Client(), server.URL).ListDevices(context.Background())
			require.Error(t, err)
		},
	)

	t.Run(
		"fails on a null device list",
		func(t *testing.T) {
			t.Parallel()

			server, _ := iruStub(t, map[string]string{"0": `null`})

			_, err := NewIruDriver(server.Client(), server.URL).ListDevices(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), "null")
		},
	)

	t.Run(
		"lists a fleet of exactly the page cap",
		func(t *testing.T) {
			t.Parallel()

			server := iruFullPagesStub(t, iruMaxPages)

			records, err := NewIruDriver(server.Client(), server.URL).ListDevices(context.Background())
			require.NoError(t, err)
			assert.Len(t, records, iruMaxPages*iruDevicesPageSize)
		},
	)

	t.Run(
		"fails on a fleet beyond the page cap",
		func(t *testing.T) {
			t.Parallel()

			server := iruFullPagesStub(t, iruMaxPages+1)

			_, err := NewIruDriver(server.Client(), server.URL).ListDevices(context.Background())
			require.ErrorIs(t, err, ErrPaginationLimitReached)
		},
	)
}

// iruFullPagesStub serves fullPages full pages of devices, then empty ones.
func iruFullPagesStub(t *testing.T, fullPages int) *httptest.Server {
	t.Helper()

	fullPage := iruDevicesJSON(t, 0, iruDevicesPageSize)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := "[]"

		var offset int
		if _, err := fmt.Sscan(r.URL.Query().Get("offset"), &offset); err == nil &&
			offset < fullPages*iruDevicesPageSize {
			body = fullPage
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server
}
