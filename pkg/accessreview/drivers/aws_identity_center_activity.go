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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cttypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
)

const (
	// identityCenterLoginLookback is CloudTrail Event History's retention.
	// Older Identity Center sign-ins are invisible, not absent.
	identityCenterLoginLookback = 90 * 24 * time.Hour

	identityCenterUserAuthentication = "UserAuthentication"
	identityStoreMFATarget           = "AWSIdentityStore.ListMfaDevicesForUser"
)

type (
	identityCenterLogin struct {
		at      time.Time
		usedMFA bool
	}

	identityCenterCloudTrailEvent struct {
		UserIdentity struct {
			OnBehalfOf struct {
				UserID string `json:"userId"`
			} `json:"onBehalfOf"`
			AdditionalEventData identityCenterCredentialData `json:"additionalEventData"`
		} `json:"userIdentity"`
		AdditionalEventData identityCenterCredentialData `json:"additionalEventData"`
	}

	identityCenterCredentialData struct {
		CredentialType string `json:"CredentialType"`
	}

	listMfaDevicesOutput struct {
		MfaDevices []json.RawMessage `json:"MfaDevices"`
		MFADevices []json.RawMessage `json:"MFADevices"`
		Devices    []json.RawMessage `json:"Devices"`
	}

	awsJSONError struct {
		code    string
		message string
	}

	setListMfaDevicesTarget struct{}

	listMfaDevicesDeserializer struct {
		out *listMfaDevicesOutput
	}
)

var _ smithy.APIError = awsJSONError{}

func (e awsJSONError) Error() string {
	if e.message == "" {
		return e.code
	}

	return e.code + ": " + e.message
}

func (e awsJSONError) ErrorCode() string {
	return e.code
}

func (e awsJSONError) ErrorMessage() string {
	return e.message
}

func (e awsJSONError) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

func (o listMfaDevicesOutput) hasDevice() bool {
	return len(o.MfaDevices) > 0 || len(o.MFADevices) > 0 || len(o.Devices) > 0
}

func enrichIdentityCenterActivity(
	ctx context.Context,
	session *cloudaws.Session,
	instance identityCenterInstance,
	users []identityCenterUser,
) error {
	cfg := session.Config()
	cfg.Region = instance.region

	logins, loginErr := lookupIdentityCenterLogins(ctx, cloudtrail.NewFromConfig(cfg), users)
	if errors.Is(loginErr, context.Canceled) {
		return loginErr
	}

	devices, deviceErr := listIdentityCenterMFADevices(
		ctx,
		identitystore.NewFromConfig(cfg),
		instance.storeID,
		identityCenterAssignedUsers(users),
	)
	if errors.Is(deviceErr, context.Canceled) {
		return deviceErr
	}

	applyIdentityCenterActivity(users, logins, devices)

	if loginErr != nil {
		return loginErr
	}

	return deviceErr
}

func lookupIdentityCenterLogins(
	ctx context.Context,
	client *cloudtrail.Client,
	users []identityCenterUser,
) (map[string]identityCenterLogin, error) {
	wanted := make(map[string]struct{}, len(users))
	for _, user := range users {
		if user.ID != "" {
			wanted[user.ID] = struct{}{}
		}
	}

	if len(wanted) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	paginator := cloudtrail.NewLookupEventsPaginator(
		client,
		&cloudtrail.LookupEventsInput{
			StartTime: aws.Time(now.Add(-identityCenterLoginLookback)),
			EndTime:   aws.Time(now),
			LookupAttributes: []cttypes.LookupAttribute{
				{
					AttributeKey:   cttypes.LookupAttributeKeyEventName,
					AttributeValue: aws.String(identityCenterUserAuthentication),
				},
			},
		},
	)

	found := make(map[string]identityCenterLogin, len(wanted))

	for range maxPaginationPages {
		if !paginator.HasMorePages() {
			return found, nil
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return found, fmt.Errorf("cannot look up identity center sign-in events: %w", err)
		}

		for _, event := range page.Events {
			userID, usedMFA, ok := parseIdentityCenterLoginEvent(aws.ToString(event.CloudTrailEvent))
			if !ok {
				continue
			}

			if _, want := wanted[userID]; !want {
				continue
			}

			if _, seen := found[userID]; seen {
				continue
			}

			if event.EventTime == nil {
				continue
			}

			found[userID] = identityCenterLogin{at: event.EventTime.UTC(), usedMFA: usedMFA}
			if len(found) == len(wanted) {
				return found, nil
			}
		}
	}

	return found, fmt.Errorf(
		"cannot look up all identity center sign-in events: %w",
		ErrPaginationLimitReached,
	)
}

func parseIdentityCenterLoginEvent(raw string) (string, bool, bool) {
	if raw == "" {
		return "", false, false
	}

	var event identityCenterCloudTrailEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return "", false, false
	}

	userID := event.UserIdentity.OnBehalfOf.UserID
	if userID == "" {
		return "", false, false
	}

	credentialType := event.AdditionalEventData.CredentialType
	if credentialType == "" {
		credentialType = event.UserIdentity.AdditionalEventData.CredentialType
	}

	return userID, identityCenterCredentialUsedMFA(credentialType), true
}

func identityCenterCredentialUsedMFA(credentialType string) bool {
	for part := range strings.SplitSeq(credentialType, ",") {
		switch strings.ToUpper(strings.TrimSpace(part)) {
		case "TOTP", "WEBAUTHN":
			return true
		}
	}

	return false
}

func listIdentityCenterMFADevices(
	ctx context.Context,
	client *identitystore.Client,
	storeID string,
	users []identityCenterUser,
) (map[string]bool, error) {
	hasDevice := make(map[string]bool)

	for _, user := range users {
		if user.ID == "" {
			continue
		}

		found, err := listMfaDevicesForUser(ctx, client, storeID, user.ID)
		if err != nil {
			return hasDevice, err
		}

		if found {
			hasDevice[user.ID] = true
		}
	}

	return hasDevice, nil
}

func applyIdentityCenterActivity(
	users []identityCenterUser,
	logins map[string]identityCenterLogin,
	hasDevice map[string]bool,
) {
	for i := range users {
		if login, ok := logins[users[i].ID]; ok {
			users[i].LastLogin = new(login.at)
			if login.usedMFA {
				users[i].MFAEnabled = new(true)
			}
		}

		if hasDevice[users[i].ID] {
			users[i].MFAEnabled = new(true)
		}
	}
}

// listMfaDevicesForUser retargets DescribeUser: same IdentityStoreId and
// UserId body, no public MFA API, so the identitystore client still signs.
func listMfaDevicesForUser(
	ctx context.Context,
	client *identitystore.Client,
	storeID string,
	userID string,
) (bool, error) {
	var devices listMfaDevicesOutput

	_, err := client.DescribeUser(
		ctx,
		&identitystore.DescribeUserInput{
			IdentityStoreId: aws.String(storeID),
			UserId:          aws.String(userID),
		},
		withListMfaDevicesForUser(&devices),
	)
	if err != nil {
		return false, fmt.Errorf("cannot list identity center mfa devices of a user: %w", err)
	}

	return devices.hasDevice(), nil
}

func withListMfaDevicesForUser(out *listMfaDevicesOutput) func(*identitystore.Options) {
	return func(o *identitystore.Options) {
		o.APIOptions = append(
			o.APIOptions,
			func(stack *middleware.Stack) error {
				if err := stack.Serialize.Add(&setListMfaDevicesTarget{}, middleware.After); err != nil {
					return err
				}

				if _, err := stack.Deserialize.Remove("OperationDeserializer"); err != nil {
					return err
				}

				return stack.Deserialize.Add(
					&listMfaDevicesDeserializer{out: out},
					middleware.After,
				)
			},
		)
	}
}

func (*setListMfaDevicesTarget) ID() string {
	return "ListMfaDevicesForUserTarget"
}

func (*setListMfaDevicesTarget) HandleSerialize(
	ctx context.Context,
	in middleware.SerializeInput,
	next middleware.SerializeHandler,
) (middleware.SerializeOutput, middleware.Metadata, error) {
	if req, ok := in.Request.(*smithyhttp.Request); ok {
		req.Header.Set("X-Amz-Target", identityStoreMFATarget)
	}

	return next.HandleSerialize(ctx, in)
}

func (*listMfaDevicesDeserializer) ID() string {
	return "OperationDeserializer"
}

func (m *listMfaDevicesDeserializer) HandleDeserialize(
	ctx context.Context,
	in middleware.DeserializeInput,
	next middleware.DeserializeHandler,
) (middleware.DeserializeOutput, middleware.Metadata, error) {
	out, metadata, err := next.HandleDeserialize(ctx, in)
	if err != nil {
		return out, metadata, err
	}

	response, ok := out.RawResponse.(*smithyhttp.Response)
	if !ok || response == nil {
		return out, metadata, fmt.Errorf("cannot decode identity center mfa devices of a user: unknown response")
	}

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return out, metadata, fmt.Errorf("cannot read identity center mfa devices of a user: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return out, metadata, parseAWSJSONError(payload)
	}

	if err := json.Unmarshal(payload, m.out); err != nil {
		return out, metadata, fmt.Errorf("cannot decode identity center mfa devices of a user: %w", err)
	}

	out.Result = &identitystore.DescribeUserOutput{}

	return out, metadata, nil
}

func parseAWSJSONError(payload []byte) error {
	var body struct {
		Type    string `json:"__type"`
		Message string `json:"message"`
		Alt     string `json:"Message"`
	}

	if err := json.Unmarshal(payload, &body); err != nil {
		return awsJSONError{code: "UnknownError"}
	}

	code := body.Type
	if i := strings.LastIndex(code, "#"); i >= 0 {
		code = code[i+1:]
	}

	if code == "" {
		code = "UnknownError"
	}

	message := body.Message
	if message == "" {
		message = body.Alt
	}

	return awsJSONError{code: code, message: message}
}
