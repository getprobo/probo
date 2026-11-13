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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	// iruDevicesPageSize is the largest page Iru's device list serves.
	iruDevicesPageSize = 300

	// iruMaxPages caps the device list at 300,000 devices. Iru documents no
	// total, so a provider that kept answering full pages would otherwise
	// be paged forever.
	iruMaxPages = 1000

	// iruErrorBodyLimit bounds how much of an error response is read before
	// it is discarded. Its text is never kept.
	iruErrorBodyLimit = 64 << 10
)

// iruAPIHostPattern is the shape of the per-tenant API host Iru shows under
// Settings > Access: a tenant label on its US or EU API domain. Iru still
// serves its API from kandji.io after the rename and documents no other host.
// Validating against it keeps a settings value from steering requests to any
// other host.
var iruAPIHostPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.api\.(eu\.)?kandji\.io$`)

// IruAPIHost canonicalizes the API URL a customer pastes from Iru, such as
// "acme.api.kandji.io" or "https://acme.api.kandji.io/", to its bare host.
func IruAPIHost(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))

	if strings.Contains(value, "://") {
		u, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("invalid iru api url")
		}

		if u.Scheme != "https" || u.User != nil || u.Port() != "" || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
			return "", fmt.Errorf("invalid iru api url")
		}

		value = u.Hostname()
	}

	value = strings.TrimSuffix(value, "/")

	if !iruAPIHostPattern.MatchString(value) {
		return "", fmt.Errorf("invalid iru api host")
	}

	return value, nil
}

// IruBaseURL builds the API base URL of an Iru tenant from its host. It is
// the one place the URL is composed, so the driver and the connection check
// cannot drift onto different hosts.
func IruBaseURL(apiHost string) (string, error) {
	host, err := IruAPIHost(apiHost)
	if err != nil {
		return "", err
	}

	base := url.URL{Scheme: "https", Host: host}

	return base.String(), nil
}

// IruDevicesProbeURL is the cheapest request that proves a token can list
// devices: one device from the endpoint the sync reads.
func IruDevicesProbeURL(apiHost string) (string, error) {
	base, err := IruBaseURL(apiHost)
	if err != nil {
		return "", err
	}

	return iruDevicesURL(base, 1, 0)
}

func iruDevicesURL(baseURL string, limit, offset int) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse iru base url: %w", err)
	}

	u = u.JoinPath("api", "v1", "devices")

	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// IruDriver lists the devices of one Iru (formerly Kandji) tenant. The
// tenant is fixed by the host, so the token's tenant is the only one read.
type IruDriver struct {
	httpClient *http.Client
	baseURL    string
}

var _ Driver = (*IruDriver)(nil)

// NewIruDriver builds a driver against baseURL, the tenant's API origin
// (https://<tenant>.api.kandji.io). The HTTP client carries the API token.
func NewIruDriver(httpClient *http.Client, baseURL string) *IruDriver {
	return &IruDriver{httpClient: httpClient, baseURL: baseURL}
}

type (
	iruDevice struct {
		DeviceID     string  `json:"device_id"`
		DeviceName   string  `json:"device_name"`
		Model        string  `json:"model"`
		SerialNumber string  `json:"serial_number"`
		Platform     string  `json:"platform"`
		OSVersion    string  `json:"os_version"`
		LastCheckIn  string  `json:"last_check_in"`
		User         iruUser `json:"user"`
		IsRemoved    bool    `json:"is_removed"`
	}

	// iruUser is the user a device is assigned to. Iru sends an object when
	// there is one and an empty string when there is not.
	iruUser struct {
		Email string `json:"email"`
	}
)

func (u *iruUser) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		*u = iruUser{}
		return nil
	}

	type plain iruUser

	var v plain
	if err := json.Unmarshal(trimmed, &v); err != nil {
		return err
	}

	*u = iruUser(v)

	return nil
}

func (d *IruDriver) ListDevices(ctx context.Context) ([]DeviceRecord, error) {
	var records []DeviceRecord

	// One page past the cap is read so that a fleet of exactly the cap,
	// whose last allowed page is full, is told apart from a larger one.
	for page := range iruMaxPages + 1 {
		devices, err := d.listPage(ctx, page*iruDevicesPageSize)
		if err != nil {
			return nil, err
		}

		if page == iruMaxPages {
			if len(devices) > 0 {
				break
			}

			return records, nil
		}

		for _, device := range devices {
			record, ok := iruDeviceRecord(device)
			if !ok {
				continue
			}

			records = append(records, record)
		}

		// Iru returns a bare array with no total, so a short page is the
		// only sign of the last one.
		if len(devices) < iruDevicesPageSize {
			return records, nil
		}
	}

	return nil, fmt.Errorf("cannot list all iru devices: %w", ErrPaginationLimitReached)
}

func (d *IruDriver) listPage(ctx context.Context, offset int) ([]iruDevice, error) {
	endpoint, err := iruDevicesURL(d.baseURL, iruDevicesPageSize, offset)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create iru devices request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute iru devices request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		// Drained but never kept: Iru's error bodies can name the tenant.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, iruErrorBodyLimit))
		return nil, fmt.Errorf("cannot fetch iru devices: unexpected status %d", resp.StatusCode)
	}

	var devices []iruDevice
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("cannot decode iru devices response: %w", err)
	}

	// A JSON null decodes to a nil slice without error. Read as an empty
	// page it would end the listing and revoke every synced device, so
	// only a literal empty array means there are no more devices.
	if devices == nil {
		return nil, fmt.Errorf("cannot decode iru devices response: device list is null")
	}

	return devices, nil
}

// iruDeviceRecord maps one Iru device. A device Iru reports as removed is no
// longer managed and is left out, so a sync treats it as gone.
func iruDeviceRecord(d iruDevice) (DeviceRecord, bool) {
	id := strings.TrimSpace(d.DeviceID)
	if id == "" || d.IsRemoved {
		return DeviceRecord{}, false
	}

	return DeviceRecord{
		ExternalID:    id,
		Name:          nonEmpty(d.DeviceName),
		SerialNumber:  nonEmpty(d.SerialNumber),
		Platform:      iruPlatform(d.Platform),
		Model:         nonEmpty(d.Model),
		OSVersion:     nonEmpty(d.OSVersion),
		UserEmail:     nonEmpty(d.User.Email),
		LastCheckInAt: parseTime(d.LastCheckIn),
	}, true
}

// iruPlatform maps Iru's platform names. A platform Iru adds later reports no
// platform rather than a wrong one.
func iruPlatform(platform string) *coredata.DevicePlatform {
	var p coredata.DevicePlatform

	switch strings.TrimSpace(platform) {
	case "Mac":
		p = coredata.DevicePlatformDarwin
	case "Windows":
		p = coredata.DevicePlatformWindows
	case "iPhone":
		p = coredata.DevicePlatformIOS
	case "iPad":
		p = coredata.DevicePlatformIPadOS
	case "AppleTV":
		p = coredata.DevicePlatformTVOS
	case "Android":
		p = coredata.DevicePlatformAndroid
	default:
		return nil
	}

	return &p
}

func nonEmpty(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	return &s
}

// parseTime reads an RFC 3339 timestamp and treats an empty or malformed one
// as absent: a check-in time is display data, not worth failing a sync over.
func parseTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil
	}

	return &t
}
