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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func requireGID(values url.Values, param string) (gid.GID, error) {
	v := values.Get(param)
	if v == "" {
		return gid.GID{}, fmt.Errorf("missing %s", param)
	}

	id, err := gid.ParseGID(v)
	if err != nil {
		return gid.GID{}, fmt.Errorf("invalid %s", param)
	}

	return id, nil
}

func parseScopes(s string) (coredata.OAuth2Scopes, error) {
	var scopes coredata.OAuth2Scopes
	if err := scopes.UnmarshalText([]byte(s)); err != nil {
		return nil, err
	}
	return scopes, nil
}

type (
	AuthorizeInput struct {
		ClientID            gid.GID
		RedirectURI         string
		State               string
		ResponseType        string
		Scopes              coredata.OAuth2Scopes
		CodeChallenge       string
		CodeChallengeMethod coredata.OAuth2CodeChallengeMethod
		Nonce               string
	}

	AuthorizeConsentInput struct {
		ConsentID gid.GID
		Action    string
	}

	IntrospectInput struct {
		Token string
	}

	RevokeInput struct {
		Token string
	}

	DeviceAuthInput struct {
		ClientID gid.GID
		Scopes   coredata.OAuth2Scopes
	}

	DeviceVerifyInput struct {
		UserCode string
	}

	DeviceVerifySubmitInput struct {
		UserCode string
	}

	AuthorizationCodeGrantInput struct {
		ClientID     string
		ClientSecret string
		Code         string
		RedirectURI  string
		CodeVerifier string
	}

	RefreshTokenGrantInput struct {
		ClientID     string
		ClientSecret string
		RefreshToken string
	}

	DeviceCodeGrantInput struct {
		ClientID   gid.GID
		DeviceCode string
	}

	RegisterInput struct {
		OrganizationID          string   `json:"organization_id"`
		ClientName              string   `json:"client_name"`
		Visibility              string   `json:"visibility"`
		RedirectURIs            []string `json:"redirect_uris"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		LogoURI                 *string  `json:"logo_uri"`
		ClientURI               *string  `json:"client_uri"`
		Contacts                []string `json:"contacts"`
		Scopes                  []string `json:"scopes"`
	}
)

func (in *AuthorizeInput) DecodeQuery(q url.Values) error {
	var err error

	in.ClientID, err = requireGID(q, "client_id")
	if err != nil {
		return err
	}

	in.RedirectURI = q.Get("redirect_uri")
	in.State = q.Get("state")
	in.ResponseType = q.Get("response_type")
	in.CodeChallenge = q.Get("code_challenge")
	in.CodeChallengeMethod = coredata.OAuth2CodeChallengeMethod(q.Get("code_challenge_method"))
	in.Nonce = q.Get("nonce")

	in.Scopes, err = parseScopes(q.Get("scope"))
	if err != nil {
		return err
	}

	return nil
}

func (in *AuthorizeConsentInput) DecodeForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data")
	}

	var err error

	in.ConsentID, err = requireGID(r.Form, "consent_id")
	if err != nil {
		return err
	}

	in.Action = r.FormValue("action")

	return nil
}

func (in *IntrospectInput) DecodeForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data")
	}

	in.Token = r.FormValue("token")
	if in.Token == "" {
		return fmt.Errorf("missing token parameter")
	}
	return nil
}

func (in *RevokeInput) DecodeForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data")
	}

	in.Token = r.FormValue("token")
	return nil
}

func (in *DeviceAuthInput) DecodeForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data")
	}

	var err error

	in.ClientID, err = requireGID(r.Form, "client_id")
	if err != nil {
		return err
	}

	if scopeStr := r.FormValue("scope"); scopeStr != "" {
		in.Scopes, err = parseScopes(scopeStr)
		if err != nil {
			return fmt.Errorf("invalid scope")
		}
	}

	return nil
}

func (in *DeviceVerifyInput) DecodeQuery(q url.Values) error {
	in.UserCode = q.Get("user_code")
	return nil
}

func (in *DeviceVerifySubmitInput) DecodeForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data")
	}

	userCode := r.FormValue("user_code")
	userCode = strings.ReplaceAll(userCode, "-", "")
	in.UserCode = strings.ToUpper(strings.TrimSpace(userCode))
	return nil
}

func (in *AuthorizationCodeGrantInput) DecodeForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data")
	}

	in.ClientID = r.FormValue("client_id")
	in.ClientSecret = r.FormValue("client_secret")
	in.Code = r.FormValue("code")
	in.RedirectURI = r.FormValue("redirect_uri")
	in.CodeVerifier = r.FormValue("code_verifier")

	if in.Code == "" {
		return fmt.Errorf("missing code")
	}

	return nil
}

func (in *RefreshTokenGrantInput) DecodeForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data")
	}

	in.ClientID = r.FormValue("client_id")
	in.ClientSecret = r.FormValue("client_secret")
	in.RefreshToken = r.FormValue("refresh_token")

	if in.RefreshToken == "" {
		return fmt.Errorf("missing refresh_token")
	}

	return nil
}

func (in *DeviceCodeGrantInput) DecodeForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid form data")
	}

	var err error

	in.ClientID, err = requireGID(r.Form, "client_id")
	if err != nil {
		return err
	}

	in.DeviceCode = r.FormValue("device_code")
	if in.DeviceCode == "" {
		return fmt.Errorf("missing device_code")
	}

	return nil
}

type (
	IntrospectResponse struct {
		Active    bool                  `json:"active"`
		Scope     coredata.OAuth2Scopes `json:"scope,omitempty"`
		ClientID  gid.GID               `json:"client_id,omitempty"`
		Sub       gid.GID               `json:"sub,omitempty"`
		Exp       int64                 `json:"exp,omitempty"`
		Iat       int64                 `json:"iat,omitempty"`
		TokenType string                `json:"token_type,omitempty"`
	}

	DeviceAuthResponse struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}

	RegisterResponse struct {
		ClientID                string   `json:"client_id"`
		ClientSecret            string   `json:"client_secret,omitempty"`
		ClientName              string   `json:"client_name"`
		Visibility              string   `json:"visibility"`
		RedirectURIs            []string `json:"redirect_uris"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		Scopes                  []string `json:"scopes"`
	}
)

func InactiveIntrospectResponse() *IntrospectResponse {
	return &IntrospectResponse{Active: false}
}

func ActiveIntrospectResponse(result *IntrospectionResult) *IntrospectResponse {
	return &IntrospectResponse{
		Active:    true,
		Scope:     result.Scopes,
		ClientID:  result.ClientID,
		Sub:       result.IdentityID,
		Exp:       result.ExpiresAt.Unix(),
		Iat:       result.CreatedAt.Unix(),
		TokenType: "Bearer",
	}
}

type (
	AuthorizeRequest struct {
		IdentityID          gid.GID
		SessionID           gid.GID
		ResponseType        string
		ClientID            gid.GID
		RedirectURI         string
		Scopes              coredata.OAuth2Scopes
		CodeChallenge       string
		CodeChallengeMethod coredata.OAuth2CodeChallengeMethod
		Nonce               string
		State               string
		AuthTime            time.Time
	}

	ConsentApprovalRequest struct {
		ConsentID  gid.GID
		IdentityID gid.GID
		Approved   bool
		AuthTime   time.Time
	}

	RegisterClientRequest struct {
		IdentityID              gid.GID
		OrganizationID          string
		ClientName              string
		Visibility              string
		RedirectURIs            []string
		GrantTypes              []string
		ResponseTypes           []string
		TokenEndpointAuthMethod string
		LogoURI                 *string
		ClientURI               *string
		Contacts                []string
		Scopes                  []string
	}

	TokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
		IDToken      string `json:"id_token,omitempty"`
		Scope        string `json:"scope,omitempty"`
	}

	IntrospectionResult struct {
		Scopes     coredata.OAuth2Scopes
		ClientID   gid.GID
		IdentityID gid.GID
		ExpiresAt  time.Time
		CreatedAt  time.Time
	}

	DeviceCodeResult struct {
		DeviceCode string
		UserCode   coredata.OAuth2UserCode
		ExpiresIn  int
		Interval   int
	}
)
