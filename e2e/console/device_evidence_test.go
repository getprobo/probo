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

package console_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const downloadDeviceEvidenceMutation = `
	mutation DownloadDeviceEvidence($input: DownloadDeviceEvidenceInput!) {
		downloadDeviceEvidence(input: $input) {
			downloadUrl
		}
	}`

func TestDeviceEvidenceDownload(t *testing.T) {
	t.Parallel()

	sharedOwner, _, sharedEmployee, sharedViewer, orgID, _ := setupDeviceEnrollmentClients(t)
	enrolled := enrollDevice(t, sharedOwner, orgID)
	deviceID := enrolled.EnrollDevice.Device.ID

	t.Run("owner downloads a pdf", func(t *testing.T) {
		t.Parallel()

		owner := sharedOwner.ForTest(t)
		url := downloadDeviceEvidence(t, owner, []string{deviceID})
		assertPDF(t, url)
	})

	t.Run("viewer downloads a pdf", func(t *testing.T) {
		t.Parallel()

		viewer := sharedViewer.ForTest(t)
		url := downloadDeviceEvidence(t, viewer, []string{deviceID})
		assertPDF(t, url)
	})

	t.Run("employee is forbidden", func(t *testing.T) {
		t.Parallel()

		employee := sharedEmployee.ForTest(t)
		_, err := employee.Do(downloadDeviceEvidenceMutation, map[string]any{
			"input": map[string]any{"deviceIds": []string{deviceID}},
		})
		testutil.RequireForbiddenError(t, err, "employee should not export device evidence")
	})

	t.Run("empty selection is invalid", func(t *testing.T) {
		t.Parallel()

		owner := sharedOwner.ForTest(t)
		_, err := owner.Do(downloadDeviceEvidenceMutation, map[string]any{
			"input": map[string]any{"deviceIds": []string{}},
		})
		testutil.RequireErrorCode(t, err, "INVALID")
	})

	t.Run("a non-device id is invalid", func(t *testing.T) {
		t.Parallel()

		owner := sharedOwner.ForTest(t)
		_, err := owner.Do(downloadDeviceEvidenceMutation, map[string]any{
			"input": map[string]any{"deviceIds": []string{orgID}},
		})
		testutil.RequireErrorCode(t, err, "INVALID")
	})
}

func downloadDeviceEvidence(t *testing.T, client *testutil.Client, deviceIDs []string) string {
	t.Helper()

	var result struct {
		DownloadDeviceEvidence struct {
			DownloadURL string `json:"downloadUrl"`
		} `json:"downloadDeviceEvidence"`
	}
	client.MustExecute(
		downloadDeviceEvidenceMutation,
		map[string]any{
			"input": map[string]any{"deviceIds": deviceIDs},
		},
		&result,
	)
	require.NotEmpty(t, result.DownloadDeviceEvidence.DownloadURL)

	return result.DownloadDeviceEvidence.DownloadURL
}

func assertPDF(t *testing.T, downloadURL string) {
	t.Helper()

	resp, err := http.Get(downloadURL)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(body), "%PDF"))
}
