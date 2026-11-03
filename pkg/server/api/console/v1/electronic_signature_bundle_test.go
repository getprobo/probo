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

package console_v1

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestHandleElectronicSignatureBundle_RejectsInvalidID(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/electronic-signatures/not-a-gid/bundle", nil)
	request = withSignatureID(request, "not-a-gid")

	handleElectronicSignatureBundle(testLogger(), nil, nil, nil)(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandleElectronicSignatureBundle_RequiresIdentity(t *testing.T) {
	t.Parallel()

	signatureID := gid.New(gid.NewTenantID(), coredata.ElectronicSignatureEntityType).String()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/electronic-signatures/"+signatureID+"/bundle", nil)
	request = withSignatureID(request, signatureID)

	handleElectronicSignatureBundle(testLogger(), nil, nil, nil)(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestBundleContentDisposition_KeepsUnicodeName(t *testing.T) {
	t.Parallel()

	header := bundleContentDisposition("Accordé.zip")

	assert.Contains(t, header, `filename="Accord_.zip"`)
	assert.Contains(t, header, "filename*=UTF-8''Accord%C3%A9.zip")
}

func withSignatureID(request *http.Request, signatureID string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("signatureID", signatureID)

	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeCtx))
}

func testLogger() *log.Logger {
	return log.NewLogger(log.WithOutput(io.Discard))
}

func TestBundleContentDisposition_QuotesSpaces(t *testing.T) {
	t.Parallel()

	header := bundleContentDisposition("Non-Disclosure Agreement.zip")

	require.Contains(t, header, `filename="Non-Disclosure Agreement.zip"`)
	assert.Contains(t, header, "filename*=UTF-8''Non-Disclosure%20Agreement.zip")
}
