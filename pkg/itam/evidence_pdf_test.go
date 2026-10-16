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
)

func TestRenderDeviceEvidenceHTML(t *testing.T) {
	t.Parallel()

	enrolledAt := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	lastSeenAt := time.Date(2026, 3, 4, 15, 30, 0, 0, time.FixedZone("CET", 3600))
	hostname := `laptop <script>`
	serial := "SN-1"

	html, err := renderDeviceEvidenceHTML(deviceEvidenceDocument{
		OrganizationName: `Acme <script>`,
		LogoSrc:          template.URL("data:image/png;base64,aaaa"),
		GeneratedAt:      "16 Oct 2026, 09:19 UTC",
		GeneratedBy:      `Ada <b>Lovelace</b>`,
		GeneratedOn:      "https://eu.probo.com",
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
	assert.Equal(t, 2, strings.Count(body, "M224,64v88H32V64"))
	assert.Equal(t, 2, strings.Count(body, "M216,56v56c0,96-88,120-88,120"))
	assert.Contains(t, body, "Acme &lt;script&gt;")
	assert.Contains(t, body, `src="data:image/png;base64,aaaa"`)
	assert.Contains(t, body, "Device posture evidence")
	assert.Contains(t, body, "Document control")
	assert.Contains(t, body, "Classification")
	assert.Contains(t, body, "Confidential")
	assert.Contains(t, body, "<td>Number of devices</td>")
	assert.Contains(t, body, "<td>2</td>")
	assert.Contains(t, body, "16 Oct 2026, 09:19 UTC")
	assert.Contains(t, body, "Ada &lt;b&gt;Lovelace&lt;/b&gt;")
	assert.Contains(t, body, "Generated on")
	assert.Contains(t, body, "https://eu.probo.com")
	assert.Contains(t, body, "laptop &lt;script&gt;")
	assert.NotContains(t, body, "<script>")
	assert.NotContains(t, body, "<b>")
	assert.Contains(t, body, "Ada Lovelace")
	assert.Contains(t, body, "Disk encryption")
	assert.Contains(t, body, "Firewall enabled")
	assert.Contains(t, body, "2 Mar 2026")
	assert.Contains(t, body, "4 Mar 2026, 14:30 UTC")
	assert.Contains(t, body, "No posture checks recorded")
	assert.Contains(t, body, evidenceUnassignedLabel)
	assert.Contains(t, body, evidenceNeverLabel)
}

func TestDeviceEvidenceFooterTemplate(t *testing.T) {
	t.Parallel()

	assert.Contains(t, deviceEvidenceFooterTemplate, "Powered by <svg")
	assert.Contains(t, deviceEvidenceFooterTemplate, `viewBox="0 0 90 24"`)
	assert.Contains(t, deviceEvidenceFooterTemplate, `<span class="pageNumber"></span>`)
	assert.Contains(t, deviceEvidenceFooterTemplate, `<span class="totalPages"></span>`)
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

func TestDeviceEvidenceImageMIME(t *testing.T) {
	t.Parallel()

	assert.True(t, deviceEvidenceImageMIME("image/png"))
	assert.True(t, deviceEvidenceImageMIME("image/svg+xml"))
	assert.False(t, deviceEvidenceImageMIME("text/html"))
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
