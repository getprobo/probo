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

// NOTE: Cloud account enum unions are duplicated here only at the *type* level
// to keep this helper package free of a build-time dependency on apps/console's
// Relay-generated artifacts (which are downstream of the GraphQL schema and
// regenerated per build). The shapes are pinned to the schema enums in
// `pkg/server/api/console/v1/graphql/cloud_account.graphql`; if a new value is
// added to the schema, TypeScript will surface the missing case here via the
// exhaustiveness check at the bottom of each switch.
//
// Consumers (e.g. `CloudAccountStatusBadge`) pass through values typed by Relay
// generated unions; TypeScript structurally narrows those literals to these
// helper-side unions.
export type CloudAccountProvider = "AWS" | "GCP" | "AZURE";

export type CloudAccountStatus =
  | "PENDING_VERIFICATION"
  | "VERIFIED"
  | "ERRORED"
  | "DISCONNECTED";

export type CloudAccountScopeKind =
  | "AWS_ACCOUNT"
  | "GCP_PROJECT"
  | "GCP_ORGANIZATION"
  | "AZURE_SUBSCRIPTION"
  | "AZURE_MANAGEMENT_GROUP"
  | "AZURE_TENANT";

export type CloudAccountStatusVariant =
  | "success"
  | "warning"
  | "danger"
  | "neutral";

type Translator = (s: string) => string;

export function getCloudAccountProviderLabel(
  __: Translator,
  provider: CloudAccountProvider,
): string {
  switch (provider) {
    case "AWS":
      return __("Amazon Web Services");
    case "GCP":
      return __("Google Cloud");
    case "AZURE":
      return __("Microsoft Azure");
    default: {
      const exhaustive: never = provider;
      return exhaustive;
    }
  }
}

export function getCloudAccountStatusLabel(
  __: Translator,
  status: CloudAccountStatus,
): string {
  switch (status) {
    case "PENDING_VERIFICATION":
      return __("Pending verification");
    case "VERIFIED":
      return __("Verified");
    case "ERRORED":
      return __("Errored");
    case "DISCONNECTED":
      return __("Disconnected");
    default: {
      const exhaustive: never = status;
      return exhaustive;
    }
  }
}

export function getCloudAccountStatusVariant(
  status: CloudAccountStatus,
): CloudAccountStatusVariant {
  switch (status) {
    case "VERIFIED":
      return "success";
    case "PENDING_VERIFICATION":
      return "warning";
    case "ERRORED":
      return "danger";
    case "DISCONNECTED":
      return "neutral";
    default: {
      const exhaustive: never = status;
      return exhaustive;
    }
  }
}

export function getCloudAccountScopeKindLabel(
  __: Translator,
  kind: CloudAccountScopeKind,
): string {
  switch (kind) {
    case "AWS_ACCOUNT":
      return __("AWS account");
    case "GCP_PROJECT":
      return __("GCP project");
    case "GCP_ORGANIZATION":
      return __("GCP organization");
    case "AZURE_SUBSCRIPTION":
      return __("Azure subscription");
    case "AZURE_MANAGEMENT_GROUP":
      return __("Azure management group");
    case "AZURE_TENANT":
      return __("Azure tenant");
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}
