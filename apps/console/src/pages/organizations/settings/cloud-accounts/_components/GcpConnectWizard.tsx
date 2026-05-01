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
  canAdvanceGcpStep,
  formatError,
  gcpWizardInitialState,
  type GcpWizardScopeKind,
  type GraphQLError,
  reduceGcpWizard,
} from "@probo/helpers";
import { useCopy } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Field,
  IconCheckmark1,
  IconShield,
  Option,
  Select,
  useToast,
} from "@probo/ui";
import { useReducer, useRef } from "react";
import { graphql, useMutation } from "react-relay";

import type { GcpConnectWizardCreateMutation } from "#/__generated__/core/GcpConnectWizardCreateMutation.graphql";
import type { GcpConnectWizardGenerateAssetsMutation } from "#/__generated__/core/GcpConnectWizardGenerateAssetsMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const generateAssetsMutation = graphql`
  mutation GcpConnectWizardGenerateAssetsMutation(
    $input: GenerateCloudAccountInstallAssetsInput!
  ) {
    generateCloudAccountInstallAssets(input: $input) {
      assets {
        __typename
        ... on GCPInstallAssets {
          setupScript
          requiredRoles
          requiredApis
        }
      }
    }
  }
`;

// IMPORTANT: this mutation contains NO credential payload. Only metadata
// (label, scope, etc.) — the GCP service-account JSON key body is uploaded
// via the dedicated /api/console/v1/cloud-accounts/credentials/upload
// endpoint after this mutation succeeds. See the credential-upload audit
// note in `cloudAccountWizard.ts`.
const createMutation = graphql`
  mutation GcpConnectWizardCreateMutation(
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

export function GcpConnectWizard(props: Props) {
  const { connectionId, onComplete, onBack } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [state, dispatch] = useReducer(reduceGcpWizard, gcpWizardInitialState);
  const [isCopied, copy] = useCopy();

  // Secret bytes (the JSON service-account key) live in this ref so that they
  // never enter React state and are never serialised into a Relay payload.
  // The textarea uses an uncontrolled-like pattern: onChange writes the
  // current value here and merely flips a "hasCredentialPayload" boolean in
  // reducer state for gating.
  const credentialRef = useRef<string>("");
  const credentialNodeRef = useRef<HTMLTextAreaElement | null>(null);

  const [generateAssets, isGenerating]
    = useMutation<GcpConnectWizardGenerateAssetsMutation>(generateAssetsMutation);
  const [createCloudAccount]
    = useMutation<GcpConnectWizardCreateMutation>(createMutation);

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
          provider: "GCP",
          scopeKind: state.scopeKind,
          scopeIdentifier: state.scopeIdentifier,
          modules: ["ACCESS_REVIEW"],
        },
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          dispatch({
            type: "install-assets-error",
            message: errors.map(e => e.message).join(", "),
          });
          return;
        }
        const assets = response.generateCloudAccountInstallAssets.assets;
        if (assets.__typename !== "GCPInstallAssets") {
          dispatch({
            type: "install-assets-error",
            message: __("Unexpected install-assets shape from the server."),
          });
          return;
        }
        dispatch({
          type: "install-assets-success",
          setupScript: assets.setupScript,
          requiredRoles: assets.requiredRoles,
          requiredApis: assets.requiredApis,
        });
      },
      onError(error) {
        dispatch({
          type: "install-assets-error",
          message: formatError(
            __("Could not generate GCP install assets"),
            error as GraphQLError,
          ),
        });
      },
    });
  };

  const handleAdvanceFromStep1 = () => {
    if (!canAdvanceGcpStep(state)) return;
    dispatch({ type: "next-step" });
    handleGenerateAssets();
  };

  const handleDownloadScript = () => {
    if (!state.setupScript) return;
    const blob = new Blob([state.setupScript], { type: "text/x-shellscript" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "setup.sh";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const validateJson = (value: string): boolean => {
    if (!value.trim()) return false;
    try {
      JSON.parse(value);
      return true;
    } catch {
      return false;
    }
  };

  const handleSubmit = () => {
    if (!canAdvanceGcpStep(state)) return;
    const payload = credentialRef.current;
    if (!validateJson(payload)) {
      dispatch({
        type: "submit-error",
        message: __("Service-account key body is not valid JSON."),
      });
      return;
    }
    dispatch({ type: "submitting" });

    const gcpProjectId
      = state.scopeKind === "GCP_PROJECT" ? state.scopeIdentifier : null;
    const gcpOrganizationId
      = state.scopeKind === "GCP_ORGANIZATION" ? state.scopeIdentifier : null;

    createCloudAccount({
      variables: {
        input: {
          organizationId,
          provider: "GCP",
          credentialKind: "GCP_SERVICE_ACCOUNT_KEY",
          label: state.label,
          scopeKind: state.scopeKind,
          scopeIdentifier: state.scopeIdentifier,
          enabledAuditModules: ["ACCESS_REVIEW"],
          gcpProjectId,
          gcpOrganizationId,
        },
        connections: [connectionId],
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          dispatch({
            type: "submit-error",
            message: errors.map(e => e.message).join(", "),
          });
          return;
        }
        const cloudAccountId = response.createCloudAccount.cloudAccount.id;
        // Now upload the credential body via the dedicated multipart endpoint.
        // The credential never travels through a GraphQL variable; it goes in
        // the multipart body, on a route excluded from request logging.
        const formData = new FormData();
        formData.append("cloud_account_id", cloudAccountId);
        formData.append("payload", payload);
        fetch("/api/console/v1/cloud-accounts/credentials/upload", {
          method: "POST",
          body: formData,
          credentials: "same-origin",
        })
          .then(async (res) => {
            // Clear the secret immediately, regardless of outcome.
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
                "GCP project connected. Verification will run momentarily.",
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
            __("Could not create GCP cloud account"),
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
          {__("2. Run setup script")}
        </li>
        <li>•</li>
        <li
          className={state.step === 3 ? "font-semibold text-txt-primary" : ""}
        >
          {__("3. Paste service-account key")}
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
            placeholder={__("e.g. Production GCP")}
            value={state.label}
            onValueChange={(value: string) =>
              dispatch({ type: "set-label", value })}
          />
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium" htmlFor="gcp-scope-kind">
              {__("Scope")}
            </label>
            <Select
              value={state.scopeKind}
              onValueChange={(value: GcpWizardScopeKind) =>
                dispatch({ type: "set-scope-kind", value })}
            >
              <Option value="GCP_PROJECT">{__("GCP project")}</Option>
              <Option value="GCP_ORGANIZATION">{__("GCP organization")}</Option>
            </Select>
          </div>
          <Field
            label={
              state.scopeKind === "GCP_PROJECT"
                ? __("Project ID")
                : __("Organization ID")
            }
            placeholder={
              state.scopeKind === "GCP_PROJECT"
                ? "my-project-123456"
                : "1234567890"
            }
            value={state.scopeIdentifier}
            onValueChange={(value: string) =>
              dispatch({ type: "set-scope-identifier", value })}
          />
        </div>
      )}

      {state.step === 2 && (
        <div className="space-y-4">
          <p className="text-sm text-txt-secondary">
            {__(
              "Run the script below in Google Cloud Shell or locally with the gcloud CLI installed — both are supported.",
            )}
          </p>
          <div className="flex gap-3 text-xs">
            <a
              className="text-txt-accent underline"
              href="https://shell.cloud.google.com"
              target="_blank"
              rel="noopener noreferrer"
            >
              {__("Open Cloud Shell")}
            </a>
            <a
              className="text-txt-accent underline"
              href="https://cloud.google.com/sdk/docs/install"
              target="_blank"
              rel="noopener noreferrer"
            >
              {__("Install gcloud CLI")}
            </a>
          </div>
          {isGenerating && (
            <p className="text-sm text-txt-tertiary">
              {__("Generating install script…")}
            </p>
          )}
          {state.setupScript && (
            <>
              <div className="relative">
                <pre className="text-xs font-mono bg-subtle border border-border-low rounded-md p-3 overflow-x-auto whitespace-pre">
                  {state.setupScript}
                </pre>
                <div className="absolute top-2 right-2 flex gap-2">
                  <Button
                    variant="secondary"
                    onClick={() => copy(state.setupScript ?? "")}
                  >
                    {isCopied
                      ? (
                          <>
                            <IconCheckmark1 size={14} />
                            {__("Copied")}
                          </>
                        )
                      : (
                          __("Copy")
                        )}
                  </Button>
                  <Button variant="secondary" onClick={handleDownloadScript}>
                    {__("Download as setup.sh")}
                  </Button>
                </div>
              </div>
              {state.requiredRoles.length > 0 && (
                <div className="text-xs text-txt-tertiary">
                  <strong>
                    {__("Roles granted")}
                    :
                  </strong>
                  {" "}
                  {state.requiredRoles.join(", ")}
                </div>
              )}
            </>
          )}
        </div>
      )}

      {state.step === 3 && (
        <div className="space-y-4">
          <div className="flex items-start gap-2 rounded-md border border-border-low bg-subtle p-3 text-xs text-txt-secondary">
            <IconShield size={14} />
            <span>
              {__(
                "Delete the key file from disk after submitting. Probo encrypts and stores the credential server-side.",
              )}
            </span>
          </div>
          <label className="flex flex-col gap-1">
            <span className="text-sm font-medium">
              {__("Service-account JSON key")}
            </span>
            <textarea
              ref={credentialNodeRef}
              className="font-mono text-xs rounded-md border border-border-low p-3 min-h-[200px]"
              autoComplete="off"
              spellCheck={false}
              data-1p-ignore="true"
              style={{ WebkitTextSecurity: "disc" } as React.CSSProperties}
              placeholder={__("Paste the full JSON key body here")}
              onChange={(e) => {
                credentialRef.current = e.target.value;
                dispatch({
                  type: "set-has-credential-payload",
                  value: e.target.value.trim().length > 0,
                });
              }}
            />
            <span className="text-xs text-txt-tertiary">
              {__(
                "Must be valid JSON. The key never enters React state or a GraphQL variable; it is uploaded over a dedicated endpoint.",
              )}
            </span>
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
        {state.step < 3 && (
          <Button
            disabled={!canAdvanceGcpStep(state) || isGenerating}
            onClick={
              state.step === 1
                ? handleAdvanceFromStep1
                : () => dispatch({ type: "next-step" })
            }
          >
            {__("Continue")}
          </Button>
        )}
        {state.step === 3 && (
          <Button
            disabled={!canAdvanceGcpStep(state)}
            onClick={() => void handleSubmit()}
          >
            {state.submitting ? __("Connecting…") : __("Connect GCP account")}
          </Button>
        )}
      </footer>
    </div>
  );
}
