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

import {
  azureWizardInitialState,
  type AzureWizardScopeKind,
  canAdvanceAzureStep,
  formatError,
  type GraphQLError,
  reduceAzureWizard,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Field,
  IconCircleInfo,
  IconShield,
  Markdown,
  Option,
  Select,
  useToast,
} from "@probo/ui";
import { useReducer, useRef } from "react";
import { graphql, useMutation } from "react-relay";

import type { AzureConnectWizardCreateMutation } from "#/__generated__/core/AzureConnectWizardCreateMutation.graphql";
import type { AzureConnectWizardGenerateAssetsMutation } from "#/__generated__/core/AzureConnectWizardGenerateAssetsMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const generateAssetsMutation = graphql`
  mutation AzureConnectWizardGenerateAssetsMutation(
    $input: GenerateCloudAccountInstallAssetsInput!
  ) {
    generateCloudAccountInstallAssets(input: $input) {
      assets {
        __typename
        ... on AzureInstallAssets {
          steps {
            title
            body
            code
          }
          requiredRbacRoles
          requiredGraphPermissions
        }
      }
    }
  }
`;

// IMPORTANT: this mutation contains tenant_id, client_id, and the scope
// identifier — but NEVER the client_secret. The client_secret is uploaded
// through the dedicated /api/console/v1/cloud-accounts/credentials/upload
// endpoint after this mutation succeeds.
const createMutation = graphql`
  mutation AzureConnectWizardCreateMutation(
    $input: CreateCloudAccountInput!
    $connections: [ID!]!
  ) {
    createCloudAccount(input: $input) {
      cloudAccount
        @prependNode(
          connections: $connections
          edgeTypeName: "CloudAccountEdge"
        ) {
        id
        provider
        label
        status
        scope {
          kind
          identifier
        }
        lastVerifiedAt
      }
      verifyStatus
      lastProbeError
    }
  }
`;

type Props = {
  connectionId: string;
  onComplete: () => void;
  onBack: () => void;
};

const ADMIN_CONSENT_NOTE =
  "Grant admin consent for Directory.Read.All. " +
  "Without admin consent, authentication will succeed but the first Microsoft Graph " +
  "call will fail with an authorization error — a non-obvious failure mode.";

export function AzureConnectWizard(props: Props) {
  const { connectionId, onComplete, onBack } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [state, dispatch] = useReducer(
    reduceAzureWizard,
    azureWizardInitialState,
  );

  // client_secret bytes live exclusively in this ref. Never in state, never
  // in a Relay payload. See cloudAccountWizard.ts for the full rationale.
  const credentialRef = useRef<string>("");
  const credentialNodeRef = useRef<HTMLInputElement | null>(null);

  const [generateAssets, isGenerating] =
    useMutation<AzureConnectWizardGenerateAssetsMutation>(
      generateAssetsMutation,
    );
  const [createCloudAccount] =
    useMutation<AzureConnectWizardCreateMutation>(createMutation);

  const clearCredential = () => {
    credentialRef.current = "";
    if (credentialNodeRef.current) {
      credentialNodeRef.current.value = "";
    }
    dispatch({ type: "set-has-credential-payload", value: false });
  };

  const handleGenerateAssets = () => {
    generateAssets({
      variables: {
        input: {
          organizationId,
          provider: "AZURE",
          scopeKind: state.scopeKind,
          scopeIdentifier: state.scopeIdentifier,
          modules: ["ACCESS_REVIEW"],
        },
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          dispatch({
            type: "install-assets-error",
            message: errors.map((e) => e.message).join(", "),
          });
          return;
        }
        const assets = response.generateCloudAccountInstallAssets.assets;
        if (assets.__typename !== "AzureInstallAssets") {
          dispatch({
            type: "install-assets-error",
            message: __("Unexpected install-assets shape from the server."),
          });
          return;
        }
        dispatch({
          type: "install-assets-success",
          steps: assets.steps.map((s) => ({
            title: s.title,
            body: s.body,
            code: s.code ?? null,
          })),
          requiredRbacRoles: assets.requiredRbacRoles,
          requiredGraphPermissions: assets.requiredGraphPermissions,
        });
      },
      onError(error) {
        dispatch({
          type: "install-assets-error",
          message: formatError(
            __("Could not generate Azure install assets"),
            error as GraphQLError,
          ),
        });
      },
    });
  };

  const handleAdvanceFromStep1 = () => {
    if (!canAdvanceAzureStep(state)) return;
    dispatch({ type: "next-step" });
    handleGenerateAssets();
  };

  const handleSubmit = () => {
    if (!canAdvanceAzureStep(state)) return;
    const secret = credentialRef.current;
    if (!secret.trim()) {
      dispatch({
        type: "submit-error",
        message: __("Client secret is required."),
      });
      return;
    }
    dispatch({ type: "submitting" });

    const azureSubscriptionId =
      state.scopeKind === "AZURE_SUBSCRIPTION" ? state.scopeIdentifier : null;
    const azureManagementGroupId =
      state.scopeKind === "AZURE_MANAGEMENT_GROUP"
        ? state.scopeIdentifier
        : null;

    createCloudAccount({
      variables: {
        input: {
          organizationId,
          provider: "AZURE",
          credentialKind: "AZURE_CLIENT_SECRET",
          label: state.label,
          scopeKind: state.scopeKind,
          scopeIdentifier: state.scopeIdentifier,
          enabledAuditModules: ["ACCESS_REVIEW"],
          azureTenantId: state.tenantId,
          azureClientId: state.clientId,
          azureSubscriptionId,
          azureManagementGroupId,
        },
        connections: [connectionId],
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          dispatch({
            type: "submit-error",
            message: errors.map((e) => e.message).join(", "),
          });
          return;
        }
        const cloudAccountId = response.createCloudAccount.cloudAccount.id;
        const formData = new FormData();
        formData.append("cloud_account_id", cloudAccountId);
        formData.append("payload", secret);
        fetch("/api/console/v1/cloud-accounts/credentials/upload", {
          method: "POST",
          body: formData,
          credentials: "same-origin",
        })
          .then(async (res) => {
            clearCredential();
            if (!res.ok) {
              const text = await res.text().catch(() => "");
              dispatch({
                type: "submit-error",
                message: text || __("Credential upload failed."),
              });
              return;
            }
            toast({
              title: __("Success"),
              description: __(
                "Azure subscription connected. Verification will run momentarily.",
              ),
              variant: "success",
            });
            onComplete();
          })
          .catch((error) => {
            clearCredential();
            dispatch({
              type: "submit-error",
              message:
                error instanceof Error
                  ? error.message
                  : __("Credential upload failed."),
            });
          });
      },
      onError(error) {
        dispatch({
          type: "submit-error",
          message: formatError(
            __("Could not create Azure cloud account"),
            error as GraphQLError,
          ),
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <ol className="flex items-center gap-2 text-xs text-txt-tertiary">
        <li
          className={state.step === 1 ? "font-semibold text-txt-primary" : ""}
        >
          {__("1. Scope")}
        </li>
        <li>•</li>
        <li
          className={state.step === 2 ? "font-semibold text-txt-primary" : ""}
        >
          {__("2. Walkthrough")}
        </li>
        <li>•</li>
        <li
          className={state.step === 3 ? "font-semibold text-txt-primary" : ""}
        >
          {__("3. App registration IDs")}
        </li>
        <li>•</li>
        <li
          className={state.step === 4 ? "font-semibold text-txt-primary" : ""}
        >
          {__("4. Paste client secret")}
        </li>
      </ol>

      {state.errorMessage && (
        <div className="rounded-md border border-border-danger bg-danger p-3 text-sm text-txt-danger">
          {state.errorMessage}
        </div>
      )}

      {state.step === 1 && (
        <div className="space-y-4">
          <Field
            label={__("Label")}
            placeholder={__("e.g. Production Azure")}
            value={state.label}
            onValueChange={(value: string) =>
              dispatch({ type: "set-label", value })
            }
          />
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium" htmlFor="azure-scope-kind">
              {__("Scope")}
            </label>
            <Select
              value={state.scopeKind}
              onValueChange={(value: AzureWizardScopeKind) =>
                dispatch({ type: "set-scope-kind", value })
              }
            >
              <Option value="AZURE_SUBSCRIPTION">
                {__("Azure subscription")}
              </Option>
              <Option value="AZURE_MANAGEMENT_GROUP">
                {__("Azure management group")}
              </Option>
              <Option value="AZURE_TENANT">{__("Azure tenant")}</Option>
            </Select>
          </div>
          <Field
            label={__("Scope identifier")}
            placeholder="00000000-0000-0000-0000-000000000000"
            help={__(
              "Subscription, management-group, or tenant ID depending on the scope.",
            )}
            value={state.scopeIdentifier}
            onValueChange={(value: string) =>
              dispatch({ type: "set-scope-identifier", value })
            }
          />
        </div>
      )}

      {state.step === 2 && (
        <div className="space-y-4">
          {isGenerating && (
            <p className="text-sm text-txt-tertiary">
              {__("Generating walkthrough…")}
            </p>
          )}
          <div className="flex items-start gap-2 rounded-md border border-border-warning bg-warning p-3 text-xs text-txt-warning">
            <IconCircleInfo size={14} />
            <span>{__(ADMIN_CONSENT_NOTE)}</span>
          </div>
          <ol className="space-y-4">
            {state.installSteps.map((step, idx) => (
              <li
                key={`${idx}-${step.title}`}
                className="rounded-md border border-border-low p-3 space-y-2"
              >
                <h4 className="text-sm font-medium">
                  {idx + 1}. {step.title}
                </h4>
                <div className="text-sm text-txt-secondary">
                  <Markdown content={step.body} />
                </div>
                {step.code && (
                  <pre className="text-xs font-mono bg-subtle border border-border-low rounded p-2 overflow-x-auto whitespace-pre">
                    {step.code}
                  </pre>
                )}
              </li>
            ))}
          </ol>
          {state.requiredRbacRoles.length > 0 && (
            <div className="text-xs text-txt-tertiary">
              <strong>{__("RBAC roles")}:</strong>{" "}
              {state.requiredRbacRoles.join(", ")}
            </div>
          )}
          {state.requiredGraphPermissions.length > 0 && (
            <div className="text-xs text-txt-tertiary">
              <strong>{__("Graph permissions")}:</strong>{" "}
              {state.requiredGraphPermissions.join(", ")}
            </div>
          )}
        </div>
      )}

      {state.step === 3 && (
        <div className="space-y-4">
          <Field
            label={__("Tenant ID")}
            placeholder="00000000-0000-0000-0000-000000000000"
            value={state.tenantId}
            onValueChange={(value: string) =>
              dispatch({ type: "set-tenant-id", value })
            }
            autoComplete="off"
            spellCheck={false}
          />
          <Field
            label={__("Client (application) ID")}
            placeholder="00000000-0000-0000-0000-000000000000"
            value={state.clientId}
            onValueChange={(value: string) =>
              dispatch({ type: "set-client-id", value })
            }
            autoComplete="off"
            spellCheck={false}
          />
        </div>
      )}

      {state.step === 4 && (
        <div className="space-y-4">
          <div className="flex items-start gap-2 rounded-md border border-border-low bg-subtle p-3 text-xs text-txt-secondary">
            <IconShield size={14} />
            <span>
              {__(
                "The client_secret never enters React state or a GraphQL variable. It is uploaded over a dedicated endpoint and encrypted server-side.",
              )}
            </span>
          </div>
          <label className="flex flex-col gap-1">
            <span className="text-sm font-medium">{__("Client secret")}</span>
            <input
              ref={credentialNodeRef}
              type="text"
              autoComplete="off"
              spellCheck={false}
              data-1p-ignore="true"
              style={{ WebkitTextSecurity: "disc" } as React.CSSProperties}
              className="font-mono text-xs rounded-md border border-border-low p-3"
              placeholder={__("Paste the client secret value")}
              onChange={(e) => {
                credentialRef.current = e.target.value;
                dispatch({
                  type: "set-has-credential-payload",
                  value: e.target.value.trim().length > 0,
                });
              }}
            />
          </label>
        </div>
      )}

      <footer className="flex justify-between">
        <Button
          variant="tertiary"
          onClick={() => {
            if (state.step === 1) {
              clearCredential();
              onBack();
            } else {
              dispatch({ type: "previous-step" });
            }
          }}
          disabled={state.submitting}
        >
          {state.step === 1 ? __("Back") : __("Previous")}
        </Button>
        {state.step < 4 && (
          <Button
            disabled={!canAdvanceAzureStep(state) || isGenerating}
            onClick={
              state.step === 1
                ? handleAdvanceFromStep1
                : () => dispatch({ type: "next-step" })
            }
          >
            {__("Continue")}
          </Button>
        )}
        {state.step === 4 && (
          <Button disabled={!canAdvanceAzureStep(state)} onClick={handleSubmit}>
            {state.submitting ? __("Connecting…") : __("Connect Azure account")}
          </Button>
        )}
      </footer>
    </div>
  );
}
