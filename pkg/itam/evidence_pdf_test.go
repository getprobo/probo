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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRenderDeviceEvidenceHTML(t *testing.T) {
	t.Parallel()

	enrolledAt := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	lastSeenAt := time.Date(2026, 3, 4, 15, 30, 0, 0, time.FixedZone("CET", 3600))
	hostname := `laptop <script>`
	serial := "SN-1"

	html, err := renderDeviceEvidenceHTML(deviceEvidenceDocument{
		Devices: []deviceEvidencePage{
			{
				Hostname:     hostname,
				State:        "ACTIVE",
				Owner:        "Ada Lovelace",
				HardwareUUID: "hw-1",
				SerialNumber: serial,
				Platform:     "macOS",
				OSVersion:    "15.3",
				AgentVersion: "1.2.0",
				EnrolledAt:   evidenceDate(&enrolledAt),
				LastSeenAt:   evidenceDateTime(&lastSeenAt),
				Postures: []deviceEvidencePosture{
					{Check: "Disk encryption", Value: "On", Status: "Pass"},
					{Check: "Firewall enabled", Value: "Off", Status: "Fail"},
				},
			},
			{
				Hostname:     evidencePendingLabel,
				State:        "PENDING",
				Owner:        evidenceUnassignedLabel,
				HardwareUUID: evidencePendingLabel,
				SerialNumber: evidencePendingLabel,
				Platform:     evidencePendingLabel,
				OSVersion:    evidencePendingLabel,
				AgentVersion: evidencePendingLabel,
				EnrolledAt:   evidencePendingLabel,
				LastSeenAt:   evidenceNeverLabel,
			},
		},
	})
	require.NoError(t, err)

	body := string(html)
	assert.Equal(t, 2, strings.Count(body, `class="device"`))
	assert.Contains(t, body, "laptop &lt;script&gt;")
	assert.NotContains(t, body, "<script>")
	assert.Contains(t, body, "Ada Lovelace")
	assert.Contains(t, body, "Disk encryption")
	assert.Contains(t, body, "Firewall enabled")
	assert.Contains(t, body, "2 Mar 2026")
	assert.Contains(t, body, "4 Mar 2026, 14:30 UTC")
	assert.Contains(t, body, "No posture checks recorded")
	assert.Contains(t, body, evidenceUnassignedLabel)
	assert.Contains(t, body, evidenceNeverLabel)
}

func TestEvidencePostureValue(t *testing.T) {
	t.Parallel()

	seconds := 15
	length := 8

	assert.Equal(t, "15s", evidencePostureValue(coredata.DevicePostureValue{
		Kind:   coredata.DevicePostureValueKindSeconds,
		Number: &seconds,
	}))
	assert.Equal(t, "Min length 8", evidencePostureValue(coredata.DevicePostureValue{
		Kind:   coredata.DevicePostureValueKindMinPasswordLength,
		Number: &length,
	}))
	assert.Equal(t, "Unknown", evidencePostureValue(coredata.DevicePostureValue{
		Kind: coredata.DevicePostureValueKindText,
	}))
}
