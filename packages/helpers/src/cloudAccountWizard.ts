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

import type { CloudAccountScopeKind } from "./cloudAccounts";

// Reducers in this module are deliberately pure: they accept the current
// state plus a discriminated-union action and return the next state. They
// own step-transition validation, mutation-result handling, and the step
// counter so each wizard React component is a thin shell over
// `useReducer(...)` and the GraphQL mutation invocations.
//
// Crucially, NO secret material (GCP service-account key body, Azure
// client_secret) is ever a state field on these reducers. Secrets travel
// through `useRef` in the React layer and are POSTed via the dedicated
// `/api/console/v1/cloud-accounts/credentials/upload` multipart endpoint.

// ---------------------------------------------------------------------------
// AWS wizard
// ---------------------------------------------------------------------------

export type AwsWizardScopeKind = Extract<CloudAccountScopeKind, "AWS_ACCOUNT">;

export type AwsWizardState = {
  step: 1 | 2 | 3;
  label: string;
  region: string;
  scopeKind: AwsWizardScopeKind;
  scopeIdentifier: string;
  // Populated by the install-assets mutation response (step 2). Non-secret.
  quickCreateURL: string | null;
  externalId: string | null;
  principalArn: string | null;
  requiredActions: readonly string[];
  // Populated in step 3.
  roleArn: string;
  submitting: boolean;
  errorMessage: string | null;
};

export type AwsWizardAction =
  | { type: "set-label"; value: string }
  | { type: "set-region"; value: string }
  | { type: "set-scope-identifier"; value: string }
  | { type: "set-role-arn"; value: string }
  | {
      type: "install-assets-success";
      quickCreateURL: string;
      externalId: string;
      principalArn: string;
      requiredActions: readonly string[];
    }
  | { type: "install-assets-error"; message: string }
  | { type: "next-step" }
  | { type: "previous-step" }
  | { type: "submitting" }
  | { type: "submit-error"; message: string }
  | { type: "reset" };

export const awsWizardInitialState: AwsWizardState = {
  step: 1,
  label: "",
  region: "us-east-1",
  scopeKind: "AWS_ACCOUNT",
  scopeIdentifier: "",
  quickCreateURL: null,
  externalId: null,
  principalArn: null,
  requiredActions: [],
  roleArn: "",
  submitting: false,
  errorMessage: null,
};

export function canAdvanceAwsStep(state: AwsWizardState): boolean {
  switch (state.step) {
    case 1:
      return (
        state.label.trim().length > 0 &&
        state.region.trim().length > 0 &&
        state.scopeIdentifier.trim().length > 0
      );
    case 2:
      return state.quickCreateURL !== null && state.externalId !== null;
    case 3:
      return state.roleArn.trim().length > 0 && !state.submitting;
    default:
      return false;
  }
}

export function reduceAwsWizard(
  state: AwsWizardState,
  action: AwsWizardAction,
): AwsWizardState {
  switch (action.type) {
    case "set-label":
      return { ...state, label: action.value, errorMessage: null };
    case "set-region":
      return { ...state, region: action.value, errorMessage: null };
    case "set-scope-identifier":
      return { ...state, scopeIdentifier: action.value, errorMessage: null };
    case "set-role-arn":
      return { ...state, roleArn: action.value, errorMessage: null };
    case "install-assets-success":
      return {
        ...state,
        quickCreateURL: action.quickCreateURL,
        externalId: action.externalId,
        principalArn: action.principalArn,
        requiredActions: action.requiredActions,
        errorMessage: null,
      };
    case "install-assets-error":
      return { ...state, errorMessage: action.message };
    case "next-step":
      if (!canAdvanceAwsStep(state)) return state;
      if (state.step === 3) return state;
      return { ...state, step: (state.step + 1) as AwsWizardState["step"] };
    case "previous-step":
      if (state.step === 1) return state;
      return { ...state, step: (state.step - 1) as AwsWizardState["step"] };
    case "submitting":
      return { ...state, submitting: true, errorMessage: null };
    case "submit-error":
      return { ...state, submitting: false, errorMessage: action.message };
    case "reset":
      return awsWizardInitialState;
    default: {
      const exhaustive: never = action;
      return exhaustive;
    }
  }
}

// ---------------------------------------------------------------------------
// GCP wizard
// ---------------------------------------------------------------------------

export type GcpWizardScopeKind = Extract<
  CloudAccountScopeKind,
  "GCP_PROJECT" | "GCP_ORGANIZATION"
>;

export type GcpWizardState = {
  step: 1 | 2 | 3;
  label: string;
  scopeKind: GcpWizardScopeKind;
  scopeIdentifier: string;
  // Step 2 install-assets response. Non-secret.
  setupScript: string | null;
  requiredRoles: readonly string[];
  requiredApis: readonly string[];
  // Tracks whether step-3 paste field is non-empty WITHOUT storing the
  // payload itself (which lives in a ref outside this state machine).
  hasCredentialPayload: boolean;
  submitting: boolean;
  errorMessage: string | null;
};

export type GcpWizardAction =
  | { type: "set-label"; value: string }
  | { type: "set-scope-kind"; value: GcpWizardScopeKind }
  | { type: "set-scope-identifier"; value: string }
  | { type: "set-has-credential-payload"; value: boolean }
  | {
      type: "install-assets-success";
      setupScript: string;
      requiredRoles: readonly string[];
      requiredApis: readonly string[];
    }
  | { type: "install-assets-error"; message: string }
  | { type: "next-step" }
  | { type: "previous-step" }
  | { type: "submitting" }
  | { type: "submit-error"; message: string }
  | { type: "reset" };

export const gcpWizardInitialState: GcpWizardState = {
  step: 1,
  label: "",
  scopeKind: "GCP_PROJECT",
  scopeIdentifier: "",
  setupScript: null,
  requiredRoles: [],
  requiredApis: [],
  hasCredentialPayload: false,
  submitting: false,
  errorMessage: null,
};

export function canAdvanceGcpStep(state: GcpWizardState): boolean {
  switch (state.step) {
    case 1:
      return (
        state.label.trim().length > 0 && state.scopeIdentifier.trim().length > 0
      );
    case 2:
      return state.setupScript !== null;
    case 3:
      return state.hasCredentialPayload && !state.submitting;
    default:
      return false;
  }
}

export function reduceGcpWizard(
  state: GcpWizardState,
  action: GcpWizardAction,
): GcpWizardState {
  switch (action.type) {
    case "set-label":
      return { ...state, label: action.value, errorMessage: null };
    case "set-scope-kind":
      return { ...state, scopeKind: action.value, errorMessage: null };
    case "set-scope-identifier":
      return { ...state, scopeIdentifier: action.value, errorMessage: null };
    case "set-has-credential-payload":
      return { ...state, hasCredentialPayload: action.value };
    case "install-assets-success":
      return {
        ...state,
        setupScript: action.setupScript,
        requiredRoles: action.requiredRoles,
        requiredApis: action.requiredApis,
        errorMessage: null,
      };
    case "install-assets-error":
      return { ...state, errorMessage: action.message };
    case "next-step":
      if (!canAdvanceGcpStep(state)) return state;
      if (state.step === 3) return state;
      return { ...state, step: (state.step + 1) as GcpWizardState["step"] };
    case "previous-step":
      if (state.step === 1) return state;
      return { ...state, step: (state.step - 1) as GcpWizardState["step"] };
    case "submitting":
      return { ...state, submitting: true, errorMessage: null };
    case "submit-error":
      return { ...state, submitting: false, errorMessage: action.message };
    case "reset":
      return gcpWizardInitialState;
    default: {
      const exhaustive: never = action;
      return exhaustive;
    }
  }
}

// ---------------------------------------------------------------------------
// Azure wizard
// ---------------------------------------------------------------------------

export type AzureWizardScopeKind = Extract<
  CloudAccountScopeKind,
  "AZURE_SUBSCRIPTION" | "AZURE_MANAGEMENT_GROUP" | "AZURE_TENANT"
>;

export type AzureInstallStep = {
  title: string;
  body: string;
  code: string | null;
};

export type AzureWizardState = {
  step: 1 | 2 | 3 | 4;
  label: string;
  scopeKind: AzureWizardScopeKind;
  scopeIdentifier: string;
  tenantId: string;
  clientId: string;
  // Step-2 install-assets response. Non-secret.
  installSteps: readonly AzureInstallStep[];
  requiredRbacRoles: readonly string[];
  requiredGraphPermissions: readonly string[];
  // See GcpWizardState.hasCredentialPayload — secret lives in a ref.
  hasCredentialPayload: boolean;
  submitting: boolean;
  errorMessage: string | null;
};

export type AzureWizardAction =
  | { type: "set-label"; value: string }
  | { type: "set-scope-kind"; value: AzureWizardScopeKind }
  | { type: "set-scope-identifier"; value: string }
  | { type: "set-tenant-id"; value: string }
  | { type: "set-client-id"; value: string }
  | { type: "set-has-credential-payload"; value: boolean }
  | {
      type: "install-assets-success";
      steps: readonly AzureInstallStep[];
      requiredRbacRoles: readonly string[];
      requiredGraphPermissions: readonly string[];
    }
  | { type: "install-assets-error"; message: string }
  | { type: "next-step" }
  | { type: "previous-step" }
  | { type: "submitting" }
  | { type: "submit-error"; message: string }
  | { type: "reset" };

export const azureWizardInitialState: AzureWizardState = {
  step: 1,
  label: "",
  scopeKind: "AZURE_SUBSCRIPTION",
  scopeIdentifier: "",
  tenantId: "",
  clientId: "",
  installSteps: [],
  requiredRbacRoles: [],
  requiredGraphPermissions: [],
  hasCredentialPayload: false,
  submitting: false,
  errorMessage: null,
};

export function canAdvanceAzureStep(state: AzureWizardState): boolean {
  switch (state.step) {
    case 1:
      return (
        state.label.trim().length > 0 && state.scopeIdentifier.trim().length > 0
      );
    case 2:
      return state.installSteps.length > 0;
    case 3:
      return (
        state.tenantId.trim().length > 0 && state.clientId.trim().length > 0
      );
    case 4:
      return state.hasCredentialPayload && !state.submitting;
    default:
      return false;
  }
}

export function reduceAzureWizard(
  state: AzureWizardState,
  action: AzureWizardAction,
): AzureWizardState {
  switch (action.type) {
    case "set-label":
      return { ...state, label: action.value, errorMessage: null };
    case "set-scope-kind":
      return { ...state, scopeKind: action.value, errorMessage: null };
    case "set-scope-identifier":
      return { ...state, scopeIdentifier: action.value, errorMessage: null };
    case "set-tenant-id":
      return { ...state, tenantId: action.value, errorMessage: null };
    case "set-client-id":
      return { ...state, clientId: action.value, errorMessage: null };
    case "set-has-credential-payload":
      return { ...state, hasCredentialPayload: action.value };
    case "install-assets-success":
      return {
        ...state,
        installSteps: action.steps,
        requiredRbacRoles: action.requiredRbacRoles,
        requiredGraphPermissions: action.requiredGraphPermissions,
        errorMessage: null,
      };
    case "install-assets-error":
      return { ...state, errorMessage: action.message };
    case "next-step":
      if (!canAdvanceAzureStep(state)) return state;
      if (state.step === 4) return state;
      return { ...state, step: (state.step + 1) as AzureWizardState["step"] };
    case "previous-step":
      if (state.step === 1) return state;
      return { ...state, step: (state.step - 1) as AzureWizardState["step"] };
    case "submitting":
      return { ...state, submitting: true, errorMessage: null };
    case "submit-error":
      return { ...state, submitting: false, errorMessage: action.message };
    case "reset":
      return azureWizardInitialState;
    default: {
      const exhaustive: never = action;
      return exhaustive;
    }
  }
}
