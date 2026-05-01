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
  type CloudAccountProvider,
  type CloudAccountScopeKind,
  type CloudAccountStatus,
  getCloudAccountProviderLabel,
  getCloudAccountScopeKindLabel,
  getCloudAccountStatusLabel,
  getCloudAccountStatusVariant,
} from "./cloudAccounts";

const fakeTranslator = (s: string) => s;

describe("getCloudAccountProviderLabel", () => {
  it.each<[CloudAccountProvider, string]>([
    ["AWS", "Amazon Web Services"],
    ["GCP", "Google Cloud"],
    ["AZURE", "Microsoft Azure"],
  ])("labels %s as %s", (provider, expected) => {
    expect(getCloudAccountProviderLabel(fakeTranslator, provider)).toBe(
      expected,
    );
  });
});

describe("getCloudAccountStatusLabel", () => {
  it.each<[CloudAccountStatus, string]>([
    ["PENDING_VERIFICATION", "Pending verification"],
    ["VERIFIED", "Verified"],
    ["ERRORED", "Errored"],
    ["DISCONNECTED", "Disconnected"],
  ])("labels %s as %s", (status, expected) => {
    expect(getCloudAccountStatusLabel(fakeTranslator, status)).toBe(expected);
  });
});

describe("getCloudAccountStatusVariant", () => {
  it("maps every status to a badge variant", () => {
    const mapping: Record<CloudAccountStatus, string> = {
      VERIFIED: getCloudAccountStatusVariant("VERIFIED"),
      PENDING_VERIFICATION: getCloudAccountStatusVariant(
        "PENDING_VERIFICATION",
      ),
      ERRORED: getCloudAccountStatusVariant("ERRORED"),
      DISCONNECTED: getCloudAccountStatusVariant("DISCONNECTED"),
    };
    expect(mapping).toMatchInlineSnapshot(`
      {
        "DISCONNECTED": "neutral",
        "ERRORED": "danger",
        "PENDING_VERIFICATION": "warning",
        "VERIFIED": "success",
      }
    `);
  });
});

describe("getCloudAccountScopeKindLabel", () => {
  it.each<[CloudAccountScopeKind, string]>([
    ["AWS_ACCOUNT", "AWS account"],
    ["GCP_PROJECT", "GCP project"],
    ["GCP_ORGANIZATION", "GCP organization"],
    ["AZURE_SUBSCRIPTION", "Azure subscription"],
    ["AZURE_MANAGEMENT_GROUP", "Azure management group"],
    ["AZURE_TENANT", "Azure tenant"],
  ])("labels %s as %s", (kind, expected) => {
    expect(getCloudAccountScopeKindLabel(fakeTranslator, kind)).toBe(expected);
  });
});
