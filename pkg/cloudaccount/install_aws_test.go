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

package cloudaccount_test

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
)

func TestBuildAWSCloudFormationQuickCreateURL(t *testing.T) {
	t.Parallel()

	cfg := cloudaccount.AWSInstallTemplateConfig{
		TemplateURL:    "https://probo-cloud-templates.example/cloud-account",
		TemplateSHA256: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		PrincipalARN:   "arn:aws:iam::000000000000:role/probo-assumer",
	}

	got, err := cloudaccount.BuildAWSCloudFormationQuickCreateURL(
		cfg,
		"abc123",
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		"us-east-1",
	)
	require.NoError(t, err)

	parsed, err := url.Parse(got)
	require.NoError(t, err)

	assert.Contains(t, parsed.Host, "us-east-1.console.aws.amazon.com")
	q := parsed.Query()
	assert.Equal(t, "abc123", q.Get("param_ExternalId"))
	assert.Equal(t, "probo-cloud-account", q.Get("param_RoleName"))
	assert.Equal(t, cfg.PrincipalARN, q.Get("param_ProboPrincipalARN"))
	templateURL := q.Get("templateURL")
	assert.Contains(t, templateURL, cfg.TemplateSHA256, "template object key should embed SHA-256")
	assert.True(t, strings.HasSuffix(parsed.Fragment, "stacks/quickcreate"), "fragment should target the quick-create page")
}

func TestBuildAWSCloudFormationQuickCreateURL_MissingTemplate(t *testing.T) {
	t.Parallel()

	_, err := cloudaccount.BuildAWSCloudFormationQuickCreateURL(
		cloudaccount.AWSInstallTemplateConfig{},
		"abc",
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		"us-east-1",
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, cloudaccount.ErrInstallTemplateUnavailable)
}

func TestRenderAWSPolicy(t *testing.T) {
	t.Parallel()

	out, err := cloudaccount.RenderAWSPolicy([]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview})
	require.NoError(t, err)

	var doc struct {
		Version   string `json:"Version"`
		Statement []struct {
			Sid    string   `json:"Sid"`
			Effect string   `json:"Effect"`
			Action []string `json:"Action"`
		} `json:"Statement"`
	}
	require.NoError(t, json.Unmarshal(out, &doc))
	assert.Equal(t, "2012-10-17", doc.Version)
	require.Len(t, doc.Statement, 1)
	assert.Equal(t, "Allow", doc.Statement[0].Effect)
	assert.Contains(t, doc.Statement[0].Action, "iam:ListUsers")
	assert.Contains(t, doc.Statement[0].Action, "sts:GetCallerIdentity")
}

func TestGenerateAWSExternalID(t *testing.T) {
	t.Parallel()

	v1, err := cloudaccount.GenerateAWSExternalID()
	require.NoError(t, err)

	v2, err := cloudaccount.GenerateAWSExternalID()
	require.NoError(t, err)

	assert.Len(t, v1, 64, "external id is 32 bytes hex-encoded -> 64 ASCII chars")
	assert.Len(t, v2, 64)
	assert.NotEqual(t, v1, v2, "generated external ids must be unique")
}
