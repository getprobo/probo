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

package aws

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/awsx/arn"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const (
	DefaultCloudFormationTemplateURL = "https://probo-cloudformation-template.s3.us-east-2.amazonaws.com/aws-audit-role.yaml"
	DefaultTerraformModuleSource     = "getprobo/audit-role/aws"

	cloudFormationConsoleHost   = "us-east-1.console.aws.amazon.com"
	cloudFormationConsolePath   = "/cloudformation/home"
	cloudFormationConsoleRegion = "us-east-1"
	cloudFormationStackName     = "probo-audit"
)

type (
	// ConnectorSetup is the server-derived values the connect dialog shows so
	// the customer never retypes an issuer, subject or audience.
	ConnectorSetup struct {
		Issuer                       string
		Audience                     string
		Subject                      string
		SuggestedRoleName            string
		TerraformSnippet             string
		CloudFormationQuickCreateURL string
	}

	// ConnectorInstallConfig is the deployment-supplied install artifacts used
	// to fill snippets and one-click links. Empty fields omit that install path.
	ConnectorInstallConfig struct {
		CloudFormationTemplateURL string
		TerraformModuleSource     string
	}

	// ConnectorSetupInput is what BuildConnectorSetup needs. Every Probo-derived
	// string is supplied already resolved so this function never talks to config
	// or an issuer.
	ConnectorSetupInput struct {
		IssuerURL                 string
		Subject                   string
		CloudFormationTemplateURL string
		TerraformModuleSource     string
	}
)

// BuildConnectorSetup fills issuer, subject and snippets from already-resolved
// values. It does not mint a token and does not read global config.
func BuildConnectorSetup(in ConnectorSetupInput) (ConnectorSetup, error) {
	if in.IssuerURL == "" {
		return ConnectorSetup{}, fmt.Errorf("cannot build aws connector setup: issuer is required")
	}

	if in.Subject == "" {
		return ConnectorSetup{}, fmt.Errorf("cannot build aws connector setup: subject is required")
	}

	quickCreateURL, err := cloudFormationQuickCreateURL(
		in.CloudFormationTemplateURL,
		in.IssuerURL,
		in.Subject,
	)
	if err != nil {
		return ConnectorSetup{}, fmt.Errorf("cannot build aws connector setup: %w", err)
	}

	return ConnectorSetup{
		Issuer:                       in.IssuerURL,
		Audience:                     identityfederation.AudienceAWS,
		Subject:                      in.Subject,
		SuggestedRoleName:            coredata.DefaultAWSRoleName,
		TerraformSnippet:             terraformSnippet(in.TerraformModuleSource, in.IssuerURL, in.Subject),
		CloudFormationQuickCreateURL: quickCreateURL,
	}, nil
}

// ConnectorSetupFor resolves issuer and subject from the live issuer, then
// fills snippets. It is what GraphQL and MCP call so those surfaces cannot
// drift from token minting.
func ConnectorSetupFor(
	issuer *identityfederation.Issuer,
	organizationID gid.GID,
	install ConnectorInstallConfig,
) (ConnectorSetup, error) {
	if issuer == nil {
		return ConnectorSetup{}, fmt.Errorf("cannot build aws connector setup: identity federation is not configured")
	}

	issuerURL, err := issuer.IssuerURL(organizationID)
	if err != nil {
		return ConnectorSetup{}, fmt.Errorf("cannot build aws connector setup: %w", err)
	}

	return BuildConnectorSetup(
		ConnectorSetupInput{
			IssuerURL:                 issuerURL,
			Subject:                   organizationID.String(),
			CloudFormationTemplateURL: install.CloudFormationTemplateURL,
			TerraformModuleSource:     install.TerraformModuleSource,
		},
	)
}

// NewConnectorSettings validates the customer-supplied role ARN. The ARN is
// stored as given (trimmed) so partition, account, path and role name stay
// exact. Returned errors are safe to show a client: they never echo the ARN.
func NewConnectorSettings(roleARN string) (coredata.AWSConnectorSettings, error) {
	roleARN = strings.TrimSpace(roleARN)

	if _, err := arn.ParseRole(roleARN); err != nil {
		if errors.Is(err, arn.ErrUnsupportedPartition) {
			return coredata.AWSConnectorSettings{}, fmt.Errorf("cannot create aws connector: awsRoleArn is not a supported AWS partition")
		}

		return coredata.AWSConnectorSettings{}, fmt.Errorf("cannot create aws connector: awsRoleArn is not an IAM role ARN")
	}

	return coredata.AWSConnectorSettings{
		RoleARN: roleARN,
	}, nil
}

func terraformSnippet(moduleSource, issuerURL, subject string) string {
	if moduleSource == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("module \"probo_audit\" {\n")
	b.WriteString("  source = ")
	b.WriteString(strconv.Quote(moduleSource))
	b.WriteString("\n\n")
	b.WriteString("  probo_issuer_url = ")
	b.WriteString(strconv.Quote(issuerURL))
	b.WriteString("\n")
	b.WriteString("  probo_subject    = ")
	b.WriteString(strconv.Quote(subject))
	b.WriteString("\n")
	b.WriteString("  role_name        = ")
	b.WriteString(strconv.Quote(coredata.DefaultAWSRoleName))
	b.WriteString("\n")
	b.WriteString("}\n")

	return b.String()
}

func cloudFormationQuickCreateURL(templateURL, issuerURL, subject string) (string, error) {
	if templateURL == "" {
		return "", nil
	}

	consoleURL := &url.URL{
		Scheme: "https",
		Host:   cloudFormationConsoleHost,
		Path:   cloudFormationConsolePath,
	}

	query := url.Values{}
	query.Set("region", cloudFormationConsoleRegion)
	consoleURL.RawQuery = query.Encode()

	fragmentQuery := url.Values{}
	fragmentQuery.Set("templateURL", templateURL)
	fragmentQuery.Set("stackName", cloudFormationStackName)
	fragmentQuery.Set("param_ProboIssuerURL", issuerURL)
	fragmentQuery.Set("param_ProboSubject", subject)
	fragmentQuery.Set("param_RoleName", coredata.DefaultAWSRoleName)

	withFragment, err := url.Parse(consoleURL.String() + "#" + "/stacks/quickcreate?" + fragmentQuery.Encode())
	if err != nil {
		return "", fmt.Errorf("cannot parse cloudformation quick-create URL: %w", err)
	}

	return withFragment.String(), nil
}
