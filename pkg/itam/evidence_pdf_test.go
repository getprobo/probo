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
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filetemplate"
)

func TestRenderDeviceEvidenceHTML(t *testing.T) {
	t.Parallel()

	enrolledAt := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	lastSeenAt := time.Date(2026, 3, 4, 15, 30, 0, 0, time.FixedZone("CET", 3600))
	hostname := `laptop <script>`
	serial := "SN-1"

	html, err := renderDeviceEvidenceHTML(deviceEvidenceDocument{
		Layout: filetemplate.Layout{
			Title:            "Device posture evidence",
			OrganizationName: `Acme <script>`,
			LogoSrc:          template.URL("data:image/png;base64,aaaa"),
			Classification:   deviceEvidenceClassification,
		},
		Kicker:      "16 Oct 2026 / Security & trust",
		GeneratedAt: "16 Oct 2026, 09:19 UTC",
		GeneratedBy: `Ada <b>Lovelace</b>`,
		GeneratedOn: "https://eu.probo.com",
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
					{Check: "Screen lock", Value: "Unknown", Status: "Unknown"},
					{Check: "Remote login", Value: "None", Status: "Not applicable"},
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
	assert.Equal(t, 3, strings.Count(body, `class="page"`))
	assert.Equal(t, 1, strings.Count(body, "Security &amp; trust"))
	assert.Equal(t, 5, strings.Count(body, "Device posture evidence"))
	assert.Equal(t, 3, strings.Count(body, `class="org-logo"`))
	assert.Contains(t, body, `src="data:image/png;base64,aaaa"`)
	assert.Contains(t, body, "Acme &lt;script&gt;")
	assert.Contains(t, body, "16 Oct 2026 / Security &amp; trust")
	assert.NotContains(t, body, "Device evidence</p>")
	assert.Contains(t, body, ">Number of devices<")
	assert.Contains(t, body, "<dd>2</dd>")
	assert.Contains(t, body, "Classification")
	assert.Contains(t, body, "CONFIDENTIAL")
	assert.Contains(t, body, "16 Oct 2026, 09:19 UTC")
	assert.Contains(t, body, "Ada &lt;b&gt;Lovelace&lt;/b&gt;")
	assert.Contains(t, body, "Generated on")
	assert.Contains(t, body, "https://eu.probo.com")
	assert.Contains(t, body, "laptop &lt;script&gt;")
	assert.NotContains(t, body, "<script>")
	assert.NotContains(t, body, "<b>")
	assert.Contains(t, body, "Ada Lovelace")
	assert.Contains(t, body, "Current postures")
	assert.Contains(t, body, "Disk encryption")
	assert.Contains(t, body, "Firewall enabled")
	assert.NotContains(t, body, "<th>Value</th>")
	assert.NotContains(t, body, "<td>On</td>")
	assert.Contains(t, body, `class="mark mark-pass"`)
	assert.Contains(t, body, `aria-label="Pass"`)
	assert.Contains(t, body, `class="mark mark-fail"`)
	assert.Contains(t, body, `aria-label="Fail"`)
	assert.Contains(t, body, `class="mark mark-unknown"`)
	assert.Contains(t, body, `aria-label="Unknown"`)
	assert.Contains(t, body, "N/A")
	assert.NotContains(t, body, "Not applicable")
	assert.Contains(t, body, "Hardware UUID")
	assert.Contains(t, body, "2 Mar 2026")
	assert.Contains(t, body, "4 Mar 2026, 14:30 UTC")
	assert.Contains(t, body, "No checks")
	assert.Contains(t, body, "No posture checks recorded")
	assert.Contains(t, body, evidenceUnassignedLabel)
	assert.Contains(t, body, evidenceNeverLabel)
	assert.NotContains(t, body, "Powered by")
}

func TestDeviceEvidenceImageMIME(t *testing.T) {
	t.Parallel()

	assert.True(t, deviceEvidenceImageMIME("image/png"))
	assert.True(t, deviceEvidenceImageMIME("image/svg+xml"))
	assert.False(t, deviceEvidenceImageMIME("text/html"))
}

func TestEvidencePlatformURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "https://eu.probo.com", evidencePlatformURL(baseurl.MustParse("https://eu.probo.com")))
	assert.Equal(t, "https://us.probo.com", evidencePlatformURL(baseurl.MustParse("https://us.probo.com/console")))
	assert.Equal(t, "http://localhost:8080", evidencePlatformURL(baseurl.MustParse("http://localhost:8080")))
	assert.Empty(t, evidencePlatformURL(nil))
}

func TestEvidenceAuthor(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ada Lovelace", evidenceAuthor("  Ada Lovelace  ", "ada@example.com"))
	assert.Equal(t, "ada@example.com", evidenceAuthor("  ", "ada@example.com"))
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
