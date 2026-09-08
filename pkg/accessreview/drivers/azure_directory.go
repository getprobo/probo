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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
)

const (
	azureGraphAPIVersion                  = "v1.0"
	azureGraphDirectoryPath               = "directoryObjects"
	azureGraphGetByIDsPath                = "getByIds"
	azureGraphUsersPath                   = "users"
	azureGraphBatchPath                   = "$batch"
	azureGraphReportsPath                 = "reports"
	azureGraphAuthenticationMethodsPath   = "authenticationMethods"
	azureGraphUserRegistrationDetailsPath = "userRegistrationDetails"
	azureGraphGetByIDsBatch               = 1000
	azureGraphLicenceErrorCode            = "Authentication_RequestFromNonPremiumTenantOrB2CTenant"
)

type (
	azureGetByIDsRequest struct {
		IDs   []string `json:"ids"`
		Types []string `json:"types"`
	}

	azureGetByIDsPage struct {
		Value    []azureDirectoryObject `json:"value"`
		NextLink string                 `json:"@odata.nextLink"`
	}

	azureDirectoryObject struct {
		ODataType         string `json:"@odata.type"`
		ID                string `json:"id"`
		DisplayName       string `json:"displayName"`
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
		AccountEnabled    *bool  `json:"accountEnabled"`
	}

	azureGraphErrorBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
)

func resolveAzurePrincipals(
	ctx context.Context,
	session *cloudazure.Session,
	identities []azureIdentity,
) error {
	if len(identities) == 0 {
		return nil
	}

	ids := make([]string, 0, len(identities))

	byID := make(map[string]int, len(identities))
	for i, identity := range identities {
		ids = append(ids, identity.PrincipalID)
		byID[identity.PrincipalID] = i
	}

	for start := 0; start < len(ids); start += azureGraphGetByIDsBatch {
		end := min(start+azureGraphGetByIDsBatch, len(ids))

		objects, err := getAzureDirectoryObjects(ctx, session, ids[start:end])
		if err != nil {
			return err
		}

		for _, object := range objects {
			if !azureDirectoryObjectResolved(object) {
				continue
			}

			i, ok := byID[object.ID]
			if !ok {
				continue
			}

			identities[i].DisplayName = object.DisplayName
			identities[i].Email = azureEmail(
				identities[i].PrincipalType,
				object.Mail,
				object.UserPrincipalName,
			)
			identities[i].AccountEnabled = object.AccountEnabled
		}
	}

	return nil
}

func getAzureDirectoryObjects(
	ctx context.Context,
	session *cloudazure.Session,
	ids []string,
) ([]azureDirectoryObject, error) {
	pageURL, err := azureGetByIDsURL(session.GraphBaseURL())
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(
		azureGetByIDsRequest{
			IDs:   ids,
			Types: []string{"user", "group", "servicePrincipal"},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot encode azure directory getByIds request: %w", err)
	}

	var (
		all     []azureDirectoryObject
		payload = body
	)

	for range maxPaginationPages {
		objects, next, err := fetchAzureGetByIDsPage(ctx, session, pageURL, payload)
		if err != nil {
			return nil, err
		}

		all = append(all, objects...)
		if next == "" {
			return all, nil
		}

		pageURL, err = sameHostNextPageURL("azure", session.GraphBaseURL(), next)
		if err != nil {
			return nil, err
		}

		payload = nil
	}

	return nil, fmt.Errorf("cannot list all azure directory objects: %w", ErrPaginationLimitReached)
}

func fetchAzureGetByIDsPage(
	ctx context.Context,
	session *cloudazure.Session,
	pageURL string,
	body []byte,
) ([]azureDirectoryObject, string, error) {
	method := http.MethodGet

	var reader io.Reader

	if body != nil {
		method = http.MethodPost
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, pageURL, reader)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create azure directory getByIds request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := session.GraphClient().Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cannot execute azure directory getByIds request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", azureGraphResponseError(resp)
	}

	var page azureGetByIDsPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", fmt.Errorf("cannot decode azure directory getByIds response: %w", err)
	}

	return page.Value, page.NextLink, nil
}

func azureGetByIDsURL(graphBaseURL string) (string, error) {
	endpoint, err := url.JoinPath(
		graphBaseURL,
		azureGraphAPIVersion,
		azureGraphDirectoryPath,
		azureGraphGetByIDsPath,
	)
	if err != nil {
		return "", fmt.Errorf("cannot build azure directory getByIds URL: %w", err)
	}

	return endpoint, nil
}

func azureDirectoryObjectResolved(object azureDirectoryObject) bool {
	if object.ID == "" {
		return false
	}

	// getByIds returns id with everything else null when the app lacks
	// permission on that object type. That shape is not a resolved object.
	return object.DisplayName != "" ||
		object.Mail != "" ||
		object.UserPrincipalName != "" ||
		object.AccountEnabled != nil
}

func azureGraphResponseError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return &azcore.ResponseError{StatusCode: resp.StatusCode}
	}

	var payload azureGraphErrorBody

	_ = json.Unmarshal(body, &payload)

	return &azcore.ResponseError{
		StatusCode: resp.StatusCode,
		ErrorCode:  payload.Error.Code,
	}
}

func azureGraphError(statusCode int, body []byte) error {
	var payload azureGraphErrorBody

	_ = json.Unmarshal(body, &payload)

	return &azcore.ResponseError{
		StatusCode: statusCode,
		ErrorCode:  payload.Error.Code,
	}
}

func azureGraphLicenceError(err error) bool {
	apiErr, ok := errors.AsType[*azcore.ResponseError](err)
	if !ok {
		return false
	}

	return apiErr.ErrorCode == azureGraphLicenceErrorCode
}

func azureGraphBadRequest(err error) bool {
	apiErr, ok := errors.AsType[*azcore.ResponseError](err)
	if !ok {
		return false
	}

	return apiErr.StatusCode == http.StatusBadRequest
}

func azureGraphEnrichmentUnavailable(err error) bool {
	return cloudazure.As[cloudazure.ErrPermissionDenied](err) ||
		cloudazure.As[cloudazure.ErrNotFound](err) ||
		azureGraphLicenceError(err)
}

func fetchAzureGraphJSON(
	ctx context.Context,
	session *cloudazure.Session,
	endpoint string,
	dst any,
) error {
	return doAzureGraphJSON(ctx, session, http.MethodGet, endpoint, nil, dst)
}

func postAzureGraphJSON(
	ctx context.Context,
	session *cloudazure.Session,
	endpoint string,
	body []byte,
	dst any,
) error {
	return doAzureGraphJSON(ctx, session, http.MethodPost, endpoint, body, dst)
}

func doAzureGraphJSON(
	ctx context.Context,
	session *cloudazure.Session,
	method string,
	endpoint string,
	body []byte,
	dst any,
) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return fmt.Errorf("cannot create azure graph request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := session.GraphClient().Do(req)
	if err != nil {
		return fmt.Errorf("cannot execute azure graph request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return azureGraphResponseError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("cannot decode azure graph response: %w", err)
	}

	return nil
}
