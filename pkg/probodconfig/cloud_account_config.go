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

package probodconfig

type (
	// CloudAccountConfig bundles the deployment-side inputs the
	// cloud-account subsystem needs: the AWS install template
	// metadata used to assemble Quick-Create URLs, the per-provider
	// enable flags, and the periodic-probe worker tuning.
	CloudAccountConfig struct {
		// AWSAssumerARN is the principal ARN customers paste into
		// the trust policy so Probo can assume their role.
		AWSAssumerARN string `json:"aws-assumer-arn"`
		// AWSTemplateURL is the bucket prefix where Probo publishes
		// its CloudFormation templates.
		AWSTemplateURL string `json:"aws-template-url"`
		// AWSTemplateSHA256 pins the active template by its
		// content hash.
		AWSTemplateSHA256 string `json:"aws-template-sha256"`
		// AWSEnabled gates AWS install assets and verify flows.
		AWSEnabled bool `json:"aws-enabled"`
		// GCPEnabled gates GCP install assets and verify flows.
		GCPEnabled bool `json:"gcp-enabled"`
		// AzureEnabled gates Azure install assets and verify flows.
		AzureEnabled bool `json:"azure-enabled"`
		// ProbeInterval is the periodic-probe poll interval in
		// seconds. Zero falls back to the worker default
		// (15 minutes).
		ProbeInterval int `json:"probe-interval"`
		// ProbeMaxConcurrency caps the number of concurrent probes
		// the worker runs. Zero falls back to the worker default
		// (4).
		ProbeMaxConcurrency int `json:"probe-max-concurrency"`
	}
)
