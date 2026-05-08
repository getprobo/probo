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

package cloudaccount

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/rand"
)

const (
	// AWSExternalIDByteLen is the entropy bytes consumed by
	// crypto/rand.HexString when generating an AWS AssumeRole
	// ExternalId. 32 bytes -> 64 hex chars, matching the precedent
	// in pkg/iam/scim/service.go and pkg/coredata/webhook_subscription.go.
	AWSExternalIDByteLen = 32

	// AWSDefaultRegion is the region the Quick-Create URL points
	// at when the caller doesn't supply one. CloudFormation IAM
	// roles are global so the region only controls where the
	// console lands; us-east-1 is the conventional default.
	AWSDefaultRegion = "us-east-1"
)

type (
	// AWSInstallAssets is the structured payload returned to the
	// frontend when the customer starts an AWS install. The
	// QuickCreateURL is content-addressed (the published S3 object
	// key embeds the template's SHA-256 hash) so the customer's
	// stack stays bound to the exact bytes they reviewed.
	AWSInstallAssets struct {
		QuickCreateURL  string                             `json:"quick_create_url"`
		ExternalID      string                             `json:"external_id"`
		PrincipalARN    string                             `json:"principal_arn"`
		RequiredActions []string                           `json:"required_actions"`
		Modules         []coredata.CloudAccountAuditModule `json:"modules"`
		TemplateSHA256  string                             `json:"template_sha256"`
	}

	// AWSInstallTemplateConfig bundles the deployment-side inputs
	// needed to assemble the Quick-Create URL: the bucket prefix
	// where Probo publishes its CloudFormation templates, the
	// SHA-256 of the active template, and the principal ARN
	// customers paste into the trust policy.
	AWSInstallTemplateConfig struct {
		TemplateURL    string
		TemplateSHA256 string
		PrincipalARN   string
		AssumerARN     string
	}
)

// GenerateAWSExternalID returns a freshly generated AssumeRole
// ExternalId. The value is 64 ASCII hex chars (32 bytes of crypto
// entropy). Callers persist it to the cloud_accounts row at
// install-asset generation time and reuse the same value on the
// subsequent Create call.
func GenerateAWSExternalID() (string, error) {
	v, err := rand.HexString(AWSExternalIDByteLen)
	if err != nil {
		return "", fmt.Errorf("cannot generate aws external id: %w", err)
	}

	return v, nil
}

// BuildAWSCloudFormationQuickCreateURL assembles a CloudFormation
// Quick-Create URL pinned to the configured content-addressed
// template object. The returned URL pre-fills the `ExternalId` and
// `RoleName` parameters (the trust policy is shipped with the
// template, so the customer only confirms and clicks Create).
//
// Returns ErrInstallTemplateUnavailable when the deployment hasn't
// configured a template URL or SHA-256.
func BuildAWSCloudFormationQuickCreateURL(
	cfg AWSInstallTemplateConfig,
	externalID string,
	modules []coredata.CloudAccountAuditModule,
	region string,
) (string, error) {
	if cfg.TemplateURL == "" || cfg.TemplateSHA256 == "" {
		return "", fmt.Errorf("cannot build aws quick-create url: %w", ErrInstallTemplateUnavailable)
	}

	if region == "" {
		region = AWSDefaultRegion
	}

	templateURL, err := buildAWSTemplateObjectURL(cfg.TemplateURL, cfg.TemplateSHA256)
	if err != nil {
		return "", err
	}

	base := fmt.Sprintf(
		"https://%s.console.aws.amazon.com/cloudformation/home",
		url.PathEscape(region),
	)

	q := url.Values{}
	q.Set("region", region)

	stackName := "probo-cloud-account"
	q.Set("stackName", stackName)
	q.Set("templateURL", templateURL)
	q.Set("param_ExternalId", externalID)
	q.Set("param_RoleName", stackName)
	q.Set("param_ProboPrincipalARN", cfg.PrincipalARN)

	return fmt.Sprintf("%s?%s#/stacks/quickcreate", base, q.Encode()), nil
}

// buildAWSTemplateObjectURL appends the SHA-256 fragment to the
// configured template URL prefix so the URL is the integrity pin.
// Customers can paste the URL into a sandbox and confirm the bytes
// match the published SHA before authorising the stack.
func buildAWSTemplateObjectURL(prefix, sha256 string) (string, error) {
	if _, err := url.Parse(prefix); err != nil {
		return "", fmt.Errorf("cannot parse aws template url prefix: %w", err)
	}

	prefix = strings.TrimRight(prefix, "/")

	return fmt.Sprintf("%s/access-review-%s.yml", prefix, sha256), nil
}

// awsPolicyDocument is the IAM-policy JSON shape CloudFormation
// templates embed under the role.
type awsPolicyDocument struct {
	Version   string               `json:"Version"`
	Statement []awsPolicyStatement `json:"Statement"`
}

type awsPolicyStatement struct {
	Sid      string   `json:"Sid,omitempty"`
	Effect   string   `json:"Effect"`
	Action   []string `json:"Action"`
	Resource string   `json:"Resource"`
}

// RenderAWSPolicy produces the JSON IAM policy CloudFormation
// templates embed in the cross-account role. Used by tests and by
// CLI operators who want to audit the bytes Probo publishes before
// the stack is created.
func RenderAWSPolicy(modules []coredata.CloudAccountAuditModule) ([]byte, error) {
	actions := AWSActionsForModules(modules)
	if len(actions) == 0 {
		return nil, fmt.Errorf("cannot render aws policy: no actions for modules %v", modules)
	}

	doc := awsPolicyDocument{
		Version: "2012-10-17",
		Statement: []awsPolicyStatement{
			{
				Sid:      "ProboCloudAccountReadOnly",
				Effect:   "Allow",
				Action:   actions,
				Resource: "*",
			},
		},
	}

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("cannot marshal aws policy: %w", err)
	}

	return out, nil
}
