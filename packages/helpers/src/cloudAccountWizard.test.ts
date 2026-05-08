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

import { describe, expect, it } from "vitest";

import {
  awsWizardInitialState,
  azureWizardInitialState,
  canAdvanceAwsStep,
  canAdvanceAzureStep,
  canAdvanceGcpStep,
  gcpWizardInitialState,
  reduceAwsWizard,
  reduceAzureWizard,
  reduceGcpWizard,
} from "./cloudAccountWizard";

describe("reduceAwsWizard", () => {
  it("starts at step 1 with empty fields", () => {
    expect(awsWizardInitialState.step).toBe(1);
    expect(canAdvanceAwsStep(awsWizardInitialState)).toBe(false);
  });

  it("blocks advance from step 1 until label, region, and scope identifier are present", () => {
    let s = awsWizardInitialState;
    s = reduceAwsWizard(s, { type: "next-step" });
    expect(s.step).toBe(1);

    s = reduceAwsWizard(s, { type: "set-label", value: "Prod AWS" });
    s = reduceAwsWizard(s, { type: "next-step" });
    expect(s.step).toBe(1);

    s = reduceAwsWizard(s, {
      type: "set-scope-identifier",
      value: "111122223333",
    });
    s = reduceAwsWizard(s, { type: "next-step" });
    expect(s.step).toBe(2);
  });

  it("blocks advance from step 2 until install-assets-success populates quickCreateURL and externalId", () => {
    let s = awsWizardInitialState;
    s = reduceAwsWizard(s, { type: "set-label", value: "Prod" });
    s = reduceAwsWizard(s, {
      type: "set-scope-identifier",
      value: "111122223333",
    });
    s = reduceAwsWizard(s, { type: "next-step" });
    expect(s.step).toBe(2);

    s = reduceAwsWizard(s, { type: "next-step" });
    expect(s.step).toBe(2);

    s = reduceAwsWizard(s, {
      type: "install-assets-success",
      quickCreateURL: "https://aws.example/cfn",
      externalId: "ext-abc",
      principalArn: "arn:aws:iam::000:role/probo",
      requiredActions: ["sts:AssumeRole"],
    });
    s = reduceAwsWizard(s, { type: "next-step" });
    expect(s.step).toBe(3);
    expect(s.externalId).toBe("ext-abc");
  });

  it("step 3 cannot advance until role ARN is set and not while submitting", () => {
    let s = awsWizardInitialState;
    s = reduceAwsWizard(s, { type: "set-label", value: "Prod" });
    s = reduceAwsWizard(s, {
      type: "set-scope-identifier",
      value: "111122223333",
    });
    s = reduceAwsWizard(s, { type: "next-step" });
    s = reduceAwsWizard(s, {
      type: "install-assets-success",
      quickCreateURL: "https://aws.example/cfn",
      externalId: "ext-abc",
      principalArn: "arn:aws:iam::000:role/probo",
      requiredActions: [],
    });
    s = reduceAwsWizard(s, { type: "next-step" });
    expect(s.step).toBe(3);
    expect(canAdvanceAwsStep(s)).toBe(false);

    s = reduceAwsWizard(s, {
      type: "set-role-arn",
      value: "arn:aws:iam::111122223333:role/probo",
    });
    expect(canAdvanceAwsStep(s)).toBe(true);

    s = reduceAwsWizard(s, { type: "submitting" });
    expect(canAdvanceAwsStep(s)).toBe(false);
    expect(s.submitting).toBe(true);
  });

  it("install-assets-error and submit-error surface the error and stop submitting", () => {
    let s = awsWizardInitialState;
    s = reduceAwsWizard(s, { type: "submitting" });
    s = reduceAwsWizard(s, { type: "submit-error", message: "boom" });
    expect(s.submitting).toBe(false);
    expect(s.errorMessage).toBe("boom");

    s = reduceAwsWizard(s, { type: "install-assets-error", message: "no" });
    expect(s.errorMessage).toBe("no");
  });

  it("previous-step decrements but never below 1; reset returns to initial", () => {
    let s = reduceAwsWizard(awsWizardInitialState, { type: "previous-step" });
    expect(s.step).toBe(1);

    s = reduceAwsWizard(s, { type: "set-label", value: "Prod" });
    s = reduceAwsWizard(s, { type: "reset" });
    expect(s).toEqual(awsWizardInitialState);
  });
});

describe("reduceGcpWizard", () => {
  it("blocks step 1 advance until label and scope identifier are present", () => {
    let s = gcpWizardInitialState;
    s = reduceGcpWizard(s, { type: "set-label", value: "Prod GCP" });
    s = reduceGcpWizard(s, { type: "next-step" });
    expect(s.step).toBe(1);

    s = reduceGcpWizard(s, {
      type: "set-scope-identifier",
      value: "my-project-id",
    });
    s = reduceGcpWizard(s, { type: "next-step" });
    expect(s.step).toBe(2);
  });

  it("blocks step 2 advance until install-assets-success populates setupScript", () => {
    let s = gcpWizardInitialState;
    s = reduceGcpWizard(s, { type: "set-label", value: "Prod" });
    s = reduceGcpWizard(s, { type: "set-scope-identifier", value: "p" });
    s = reduceGcpWizard(s, { type: "next-step" });
    s = reduceGcpWizard(s, { type: "next-step" });
    expect(s.step).toBe(2);

    s = reduceGcpWizard(s, {
      type: "install-assets-success",
      setupScript: "#!/bin/bash\n",
      requiredRoles: ["roles/iam.securityReviewer"],
      requiredApis: ["iam.googleapis.com"],
    });
    s = reduceGcpWizard(s, { type: "next-step" });
    expect(s.step).toBe(3);
  });

  it("step 3 advance gated on hasCredentialPayload + not submitting (secret stays out of state)", () => {
    let s = gcpWizardInitialState;
    s = reduceGcpWizard(s, { type: "set-label", value: "Prod" });
    s = reduceGcpWizard(s, { type: "set-scope-identifier", value: "p" });
    s = reduceGcpWizard(s, { type: "next-step" });
    s = reduceGcpWizard(s, {
      type: "install-assets-success",
      setupScript: "x",
      requiredRoles: [],
      requiredApis: [],
    });
    s = reduceGcpWizard(s, { type: "next-step" });
    expect(s.step).toBe(3);
    expect(canAdvanceGcpStep(s)).toBe(false);

    s = reduceGcpWizard(s, { type: "set-has-credential-payload", value: true });
    expect(canAdvanceGcpStep(s)).toBe(true);

    s = reduceGcpWizard(s, { type: "submitting" });
    expect(canAdvanceGcpStep(s)).toBe(false);
  });

  it("scope-kind switching keeps state otherwise consistent", () => {
    let s = gcpWizardInitialState;
    s = reduceGcpWizard(s, {
      type: "set-scope-kind",
      value: "GCP_ORGANIZATION",
    });
    expect(s.scopeKind).toBe("GCP_ORGANIZATION");
  });
});

describe("reduceAzureWizard", () => {
  it("step 1 requires label + scope identifier", () => {
    let s = azureWizardInitialState;
    s = reduceAzureWizard(s, { type: "set-label", value: "Prod Azure" });
    s = reduceAzureWizard(s, { type: "next-step" });
    expect(s.step).toBe(1);

    s = reduceAzureWizard(s, {
      type: "set-scope-identifier",
      value: "00000000-0000-0000-0000-000000000000",
    });
    s = reduceAzureWizard(s, { type: "next-step" });
    expect(s.step).toBe(2);
  });

  it("step 2 requires non-empty installSteps", () => {
    let s = azureWizardInitialState;
    s = reduceAzureWizard(s, { type: "set-label", value: "Prod" });
    s = reduceAzureWizard(s, { type: "set-scope-identifier", value: "sub" });
    s = reduceAzureWizard(s, { type: "next-step" });
    expect(canAdvanceAzureStep(s)).toBe(false);

    s = reduceAzureWizard(s, {
      type: "install-assets-success",
      steps: [{ title: "Grant admin consent", body: "...", code: null }],
      requiredRbacRoles: ["Reader"],
      requiredGraphPermissions: ["Directory.Read.All"],
    });
    expect(canAdvanceAzureStep(s)).toBe(true);
    s = reduceAzureWizard(s, { type: "next-step" });
    expect(s.step).toBe(3);
  });

  it("step 3 requires tenantId + clientId; step 4 requires hasCredentialPayload", () => {
    let s = azureWizardInitialState;
    s = reduceAzureWizard(s, { type: "set-label", value: "Prod" });
    s = reduceAzureWizard(s, { type: "set-scope-identifier", value: "sub" });
    s = reduceAzureWizard(s, { type: "next-step" });
    s = reduceAzureWizard(s, {
      type: "install-assets-success",
      steps: [{ title: "x", body: "y", code: null }],
      requiredRbacRoles: [],
      requiredGraphPermissions: [],
    });
    s = reduceAzureWizard(s, { type: "next-step" });
    expect(s.step).toBe(3);

    expect(canAdvanceAzureStep(s)).toBe(false);
    s = reduceAzureWizard(s, { type: "set-tenant-id", value: "tid" });
    expect(canAdvanceAzureStep(s)).toBe(false);
    s = reduceAzureWizard(s, { type: "set-client-id", value: "cid" });
    expect(canAdvanceAzureStep(s)).toBe(true);

    s = reduceAzureWizard(s, { type: "next-step" });
    expect(s.step).toBe(4);
    expect(canAdvanceAzureStep(s)).toBe(false);
    s = reduceAzureWizard(s, {
      type: "set-has-credential-payload",
      value: true,
    });
    expect(canAdvanceAzureStep(s)).toBe(true);
  });

  it("reset returns to initial state", () => {
    let s = azureWizardInitialState;
    s = reduceAzureWizard(s, { type: "set-label", value: "Prod" });
    s = reduceAzureWizard(s, { type: "set-tenant-id", value: "tid" });
    s = reduceAzureWizard(s, { type: "reset" });
    expect(s).toEqual(azureWizardInitialState);
  });
});
