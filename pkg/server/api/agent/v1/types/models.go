// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package types

import (
	"encoding/json"
	"errors"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/itam"
)

const (
	heartbeatIntervalSeconds = 300
	postureIntervalSeconds   = 3600
)

type (
	EnrollRequest struct {
		EnrollmentToken string                  `json:"enrollment_token"`
		HardwareUUID    string                  `json:"hardware_uuid"`
		SerialNumber    *string                 `json:"serial_number,omitempty"`
		Hostname        string                  `json:"hostname"`
		Platform        coredata.DevicePlatform `json:"platform"`
		OSVersion       string                  `json:"os_version"`
		AgentVersion    string                  `json:"agent_version"`
	}

	EnrollResponse struct {
		DeviceID         string `json:"device_id"`
		APIKey           string `json:"api_key"`
		HeartbeatSeconds int    `json:"heartbeat_interval_seconds"`
		PostureSeconds   int    `json:"posture_interval_seconds"`
		ServerTime       string `json:"server_time"`
	}

	HeartbeatRequest struct {
		AgentVersion string `json:"agent_version,omitempty"`
		Hostname     string `json:"hostname,omitempty"`
		OSVersion    string `json:"os_version,omitempty"`
		UptimeSec    int64  `json:"uptime_seconds,omitempty"`
	}

	HeartbeatResponse struct {
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
)

func (r EnrollRequest) Validate() error {
	if r.EnrollmentToken == "" {
		return errors.New("enrollment_token is required")
	}
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

func NewEnrollResponse(result *itam.EnrollDeviceResult) *EnrollResponse {
	return &EnrollResponse{
		DeviceID:         result.Device.ID.String(),
		APIKey:           result.APIKey,
		HeartbeatSeconds: heartbeatIntervalSeconds,
		PostureSeconds:   postureIntervalSeconds,
		ServerTime:       time.Now().UTC().Format(time.RFC3339),
	}
}

func (r HeartbeatRequest) Validate() error {
	return nil
}

func NewHeartbeatResponse() *HeartbeatResponse {
	return &HeartbeatResponse{
		HeartbeatSeconds: heartbeatIntervalSeconds,
		PostureSeconds:   postureIntervalSeconds,
		ServerTime:       time.Now().UTC().Format(time.RFC3339),
	}
}

func (r PostureRequest) Validate() error {
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
