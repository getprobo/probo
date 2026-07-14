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

package types

import (
	"encoding/json"
	"errors"
	"time"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	heartbeatIntervalSeconds    = 300
	postureIntervalSeconds      = 3600
	maxPostureResultsPerRequest = 100
)

type (
	HeartbeatRequest struct {
		HardwareUUID string                  `json:"hardware_uuid"`
		SerialNumber *string                 `json:"serial_number,omitempty"`
		Hostname     string                  `json:"hostname"`
		Platform     coredata.DevicePlatform `json:"platform"`
		OSVersion    string                  `json:"os_version"`
		AgentVersion string                  `json:"agent_version"`
	}

	HeartbeatResponse struct {
		DeviceID         string `json:"device_id"`
		HeartbeatSeconds int    `json:"heartbeat_interval_seconds"`
		PostureSeconds   int    `json:"posture_interval_seconds"`
		ServerTime       string `json:"server_time"`
	}

	PostureResultPayload struct {
		CheckKey   string                       `json:"check_key"`
		Status     coredata.DevicePostureStatus `json:"status"`
		Evidence   json.RawMessage              `json:"evidence,omitempty"`
		ObservedAt time.Time                    `json:"observed_at"`
	}

	PostureRequest struct {
		Results []PostureResultPayload `json:"results"`
	}

	EnrollRequest struct {
		Token string `json:"token"`
	}

	EnrollResponse struct {
		APIKey string `json:"api_key"`
	}
)

func (r HeartbeatRequest) Validate() error {
	if r.HardwareUUID == "" {
		return errors.New("hardware_uuid is required")
	}

	if r.Hostname == "" {
		return errors.New("hostname is required")
	}

	if !r.Platform.IsValid() {
		return errors.New("platform is invalid")
	}

	if r.OSVersion == "" {
		return errors.New("os_version is required")
	}

	if r.AgentVersion == "" {
		return errors.New("agent_version is required")
	}

	return nil
}

func NewHeartbeatResponse(device *coredata.Device) *HeartbeatResponse {
	return &HeartbeatResponse{
		DeviceID:         device.ID.String(),
		HeartbeatSeconds: heartbeatIntervalSeconds,
		PostureSeconds:   postureIntervalSeconds,
		ServerTime:       time.Now().UTC().Format(time.RFC3339),
	}
}

func (r PostureRequest) Validate() error {
	if len(r.Results) > maxPostureResultsPerRequest {
		return errors.New("too many results")
	}

	for _, result := range r.Results {
		if result.CheckKey == "" {
			return errors.New("check_key is required")
		}

		if !result.Status.IsValid() {
			return errors.New("status is invalid")
		}
	}

	return nil
}

func (r EnrollRequest) Validate() error {
	if r.Token == "" {
		return errors.New("token is required")
	}

	return nil
}
