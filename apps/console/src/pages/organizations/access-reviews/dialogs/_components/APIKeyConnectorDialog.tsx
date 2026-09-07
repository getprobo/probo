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

import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useMutation, useRelayEnvironment } from "react-relay";
import { fetchQuery, graphql } from "relay-runtime";

import type { APIKeyConnectorDialog_provider$key } from "#/__generated__/core/APIKeyConnectorDialog_provider.graphql";
import type { APIKeyConnectorDialogCreateAPIKeyConnectorMutation } from "#/__generated__/core/APIKeyConnectorDialogCreateAPIKeyConnectorMutation.graphql";
import type { APIKeyConnectorDialogCrispVerificationCodeQuery } from "#/__generated__/core/APIKeyConnectorDialogCrispVerificationCodeQuery.graphql";

import { useCreateAccessReviewSource } from "../_hooks/useCreateAccessReviewSource";
import {
  buildExtraFields,
  hasRequiredExtraSettings,
  mapAPIKeyExtraSettingToField,
} from "../_lib/connectorSettings";
import {
  isPostHogDeploymentSelected,
  PostHogDeploymentField,
} from "../PostHogDeploymentField";

import { ConnectorDocumentationLink } from "./ConnectorDocumentationLink";

const apiKeyConnectorDialogFragment = graphql`
  fragment APIKeyConnectorDialog_provider on ConnectorProviderInfo {
    provider
    displayName
    documentationUrl
    apiKeyManaged
    apiKeyExtraSettings {
      key
      label
      required
    }
    apiKeyFormat {
      pattern
      example
    }
  }
`;

const createAPIKeyConnectorMutation = graphql`
  mutation APIKeyConnectorDialogCreateAPIKeyConnectorMutation(
    $input: CreateAPIKeyConnectorInput!
  ) {
    createAPIKeyConnector(input: $input) {
      connector {
        id
        provider
      }
    }
  }
`;

// Crisp (a managed provider) proves website ownership before connecting: Probo
// derives a per-(organization, website) code the customer pastes into the plugin
// settings in their Crisp dashboard. The code depends on the typed Website ID,
// so it is fetched on demand rather than carried on ConnectorProviderInfo.
const crispVerificationCodeQuery = graphql`
  query APIKeyConnectorDialogCrispVerificationCodeQuery(
    $organizationId: ID!
    $websiteId: String!
  ) {
    crispVerificationCode(organizationId: $organizationId, websiteId: $websiteId)
  }
`;

type CrispCodeState = {
  websiteId: string;
  status: "loading" | "ok" | "error";
  code: string;
};

type Props = {
  providerKey: APIKeyConnectorDialog_provider$key | null;
  organizationId: string;
  connectionId: string;
  onClose: () => void;
  onSuccess: () => void;
};

export function APIKeyConnectorDialog({
  providerKey,
  organizationId,
  connectionId,
  onClose,
  onSuccess,
}: Props) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const provider = useFragment(apiKeyConnectorDialogFragment, providerKey);
  const dialogRef = useDialogRef();
  const environment = useRelayEnvironment();

  const [apiKeyValue, setApiKeyValue] = useState("");
  // The shape check is shown once the customer has left the field or tried to
  // connect. Judging a key while it is still being typed would mark every
  // partial paste as wrong.
  const [apiKeyBlurred, setApiKeyBlurred] = useState(false);
  const [extraSettingValues, setExtraSettingValues] = useState<Record<string, string>>({});
  const [isConnectingAPIKey, setIsConnectingAPIKey] = useState(false);
  const [crispCode, setCrispCode] = useState<CrispCodeState | null>(null);
  const [crispRetry, setCrispRetry] = useState(0);

  const [createAPIKeyConnector]
    = useMutation<APIKeyConnectorDialogCreateAPIKeyConnectorMutation>(
      createAPIKeyConnectorMutation,
    );

  const createSourceAfterConnector = useCreateAccessReviewSource({
    organizationId,
    connectionId,
    onSuccess,
  });

  // Opening is driven imperatively by the parent's active-provider state. Form
  // state is reset when the dialog closes (onClose for a dismiss, the success
  // callback otherwise), so opening only shows the dialog — no setState here.
  useEffect(() => {
    if (provider) {
      dialogRef.current?.open();
    }
  }, [dialogRef, provider]);

  const isCrispManaged
    = provider?.provider === "CRISP" && !!provider.apiKeyManaged;
  const crispWebsiteId = extraSettingValues.websiteId?.trim() ?? "";

  // The fetch result is stored keyed by the Website ID it was minted for and
  // read only while it still matches the current Website ID, so a slow response
  // for a previous ID is ignored. crispVerificationCode is non-null only once a
  // code has actually been minted (status "ok"); loading and error are distinct
  // states so the UI shows progress or an actionable error, never a blank code.
  const crispCodeState
    = crispCode && crispCode.websiteId === crispWebsiteId ? crispCode : null;
  const crispVerificationCode
    = crispCodeState?.status === "ok" ? crispCodeState.code : null;

  // Fetch the ownership-verification code once a Website ID is entered, debounced
  // so it is not recomputed on every keystroke. The cleanup cancels a superseded
  // request so a late response for a previous Website ID cannot clobber the
  // current one, and crispRetry lets the user re-trigger a fetch after a failure.
  useEffect(() => {
    if (!isCrispManaged || !crispWebsiteId) {
      return;
    }

    let cancelled = false;

    const handle = setTimeout(() => {
      setCrispCode({ websiteId: crispWebsiteId, status: "loading", code: "" });

      fetchQuery<APIKeyConnectorDialogCrispVerificationCodeQuery>(
        environment,
        crispVerificationCodeQuery,
        { organizationId, websiteId: crispWebsiteId },
      )
        .toPromise()
        .then((res) => {
          if (cancelled) {
            return;
          }
          const code = res?.crispVerificationCode ?? "";
          setCrispCode({
            websiteId: crispWebsiteId,
            status: code ? "ok" : "error",
            code,
          });
        })
        .catch(() => {
          if (!cancelled) {
            setCrispCode({ websiteId: crispWebsiteId, status: "error", code: "" });
          }
        });
    }, 400);

    return () => {
      cancelled = true;
      clearTimeout(handle);
    };
  }, [environment, isCrispManaged, crispWebsiteId, organizationId, crispRetry]);

  // The provider states the shape it mints keys in; the server applies the
  // same rule at create, so this is what saves a round trip, not what enforces
  // it. A pattern that will not compile here is treated as no check at all —
  // the server still has the real one.
  const apiKeyPattern = useMemo(() => {
    const pattern = provider?.apiKeyFormat?.pattern;
    if (!pattern) {
      return null;
    }

    try {
      return new RegExp(pattern);
    } catch {
      return null;
    }
  }, [provider?.apiKeyFormat?.pattern]);

  const trimmedAPIKey = apiKeyValue.trim();
  const apiKeyMalformed
    = !!apiKeyPattern && trimmedAPIKey !== "" && !apiKeyPattern.test(trimmedAPIKey);

  const connectAPIKeyProvider = () => {
    // Managed providers (Model B, e.g. Crisp) supply no customer key: the
    // server injects Probo's own credential, so only the extra settings
    // are required.
    if (!provider || (!provider.apiKeyManaged && !trimmedAPIKey)) {
      return;
    }

    if (apiKeyMalformed) {
      setApiKeyBlurred(true);

      return;
    }

    const requiredSettings = provider.apiKeyExtraSettings.filter(s => s.required);
    if (!hasRequiredExtraSettings(requiredSettings, extraSettingValues)) {
      return;
    }

    setIsConnectingAPIKey(true);

    const extraFields = buildExtraFields(
      provider.provider,
      provider.apiKeyExtraSettings,
      extraSettingValues,
      mapAPIKeyExtraSettingToField,
    );

    createAPIKeyConnector({
      variables: {
        input: {
          organizationId,
          provider: provider.provider,
          apiKey: provider.apiKeyManaged ? null : apiKeyValue.trim(),
          ...extraFields,
        },
      },
      onCompleted: (response) => {
        const connectorId = response.createAPIKeyConnector.connector.id;
        createSourceAfterConnector(
          connectorId,
          provider.displayName,
          () => {
            setIsConnectingAPIKey(false);
            setApiKeyValue("");
            setApiKeyBlurred(false);
            setExtraSettingValues({});
            setCrispCode(null);
            dialogRef.current?.close();
            onClose();
          },
        );
      },
      onError: () => {
        setIsConnectingAPIKey(false);
        toast({
          title: t("apiKeyConnectorDialog.messages.connectionFailed"),
          // Managed providers (e.g. Crisp) never show an API key field, so
          // pointing the user at their key would be misleading; send them to
          // the settings and verification step instead.
          description: provider.apiKeyManaged
            ? t("apiKeyConnectorDialog.errors.managedConnect")
            : t("apiKeyConnectorDialog.errors.apiKeyConnect"),
          variant: "error",
        });
      },
    });
  };

  // PostHog renders a dedicated deployment selector (Cloud region or
  // self-hosted URL) and Segment a region selector; every other provider falls
  // back to generic fields.
  const renderAPIKeyExtraSettings = () => {
    if (!provider) {
      return null;
    }

    if (provider.provider === "POSTHOG") {
      return (
        <PostHogDeploymentField
          values={extraSettingValues}
          onChange={setExtraSettingValues}
        />
      );
    }

    // The server maps this to a regional API host and rejects anything outside
    // the allow-list, so it must not be typed by hand.
    if (provider.provider === "SEGMENT") {
      return (
        <div className="space-y-1.5">
          <label className="text-sm font-medium">{t("accessReviewSource.regions.label")}</label>
          <Select
            value={extraSettingValues.region ?? ""}
            onValueChange={(val: string) =>
              setExtraSettingValues(prev => ({ ...prev, region: val }))}
            placeholder={t("accessReviewSource.regions.placeholder")}
          >
            <Option value="US">{t("accessReviewSource.regions.unitedStates")}</Option>
            <Option value="EU">{t("accessReviewSource.regions.europe")}</Option>
          </Select>
        </div>
      );
    }

    return provider.apiKeyExtraSettings.map((setting) => {
      const value = extraSettingValues[setting.key] ?? "";
      return (
        <Field
          key={setting.key}
          label={setting.label}
          value={value}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
            setExtraSettingValues(prev => ({ ...prev, [setting.key]: e.target.value }))}
          required={setting.required}
        />
      );
    });
  };

  // PostHog's extra settings are individually optional (region OR instance
  // URL), so the generic required-field check can't gate it.
  const postHogAPIKeyValid
    = provider?.provider !== "POSTHOG"
      || isPostHogDeploymentSelected(extraSettingValues);

  const apiKeyExtraSettingsValid = provider
    ? hasRequiredExtraSettings(provider.apiKeyExtraSettings, extraSettingValues)
    : true;

  return (
    <Dialog
      ref={dialogRef}
      onClose={() => {
        // Reset on dismiss so the next open starts fresh (the imperative
        // close() on success does not fire onClose, so success resets inline).
        setApiKeyValue("");
        setApiKeyBlurred(false);
        setExtraSettingValues({});
        setCrispCode(null);
        setIsConnectingAPIKey(false);
        onClose();
      }}
      title={provider
        ? t("apiKeyConnectorDialog.titleWithProvider", {
            provider: provider.displayName,
          })
        : t("apiKeyConnectorDialog.title")}
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          connectAPIKeyProvider();
        }}
      >
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            {provider?.apiKeyManaged
              ? t("apiKeyConnectorDialog.managedDescription", {
                  provider: provider?.displayName ?? "",
                })
              : t("apiKeyConnectorDialog.apiKeyDescription", {
                  provider: provider?.displayName ?? "",
                })}
          </p>
          {!provider?.apiKeyManaged && (
            <Field
              label={t("apiKeyConnectorDialog.apiKey")}
              type="password"
              value={apiKeyValue}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setApiKeyValue(e.target.value)}
              onBlur={() => setApiKeyBlurred(true)}
              placeholder={provider?.apiKeyFormat?.example}
              error={
                apiKeyBlurred && apiKeyMalformed && provider?.apiKeyFormat
                  ? t("apiKeyConnectorDialog.errors.apiKeyFormat", {
                      example: provider.apiKeyFormat.example,
                    })
                  : undefined
              }
              required
              autoFocus
            />
          )}
          {renderAPIKeyExtraSettings()}
          {isCrispManaged && (
            <div className="space-y-1.5">
              <label className="text-sm font-medium">
                {t("apiKeyConnectorDialog.verificationCode.label")}
              </label>
              <p className="text-txt-tertiary text-sm">
                {t("apiKeyConnectorDialog.verificationCode.description")}
              </p>
              {!crispWebsiteId
                ? (
                    <p className="text-txt-tertiary text-sm">
                      {t("apiKeyConnectorDialog.verificationCode.enterWebsiteId")}
                    </p>
                  )
                : crispCodeState?.status === "ok"
                  ? (
                      <div className="flex items-center gap-2">
                        <code className="rounded border border-border-solid bg-subtle px-2 py-1 font-mono text-sm">
                          {crispCodeState.code}
                        </code>
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={() => {
                            const onCopyFailure = () =>
                              toast({
                                title: t("apiKeyConnectorDialog.messages.copyFailed"),
                                description: t("apiKeyConnectorDialog.errors.copy"),
                                variant: "error",
                              });

                            // navigator.clipboard is undefined in an insecure
                            // context or unsupported embedded browser, where
                            // writeText throws synchronously before .then; guard
                            // so the manual-copy toast still shows.
                            if (!navigator.clipboard?.writeText) {
                              onCopyFailure();
                              return;
                            }

                            // Copying feeds the Crisp connect flow, so only
                            // claim success once the write actually resolves.
                            try {
                              navigator.clipboard.writeText(crispCodeState.code).then(
                                () =>
                                  toast({
                                    title: t("apiKeyConnectorDialog.messages.copied"),
                                    description: t("apiKeyConnectorDialog.verificationCode.label"),
                                    variant: "success",
                                  }),
                                onCopyFailure,
                              );
                            } catch {
                              onCopyFailure();
                            }
                          }}
                        >
                          {t("apiKeyConnectorDialog.actions.copy")}
                        </Button>
                      </div>
                    )
                  : crispCodeState?.status === "error"
                    ? (
                        <div className="flex items-center gap-2">
                          <p className="text-txt-danger text-sm">
                            {t("apiKeyConnectorDialog.errors.generateCode")}
                          </p>
                          <Button
                            type="button"
                            variant="secondary"
                            onClick={() => setCrispRetry(n => n + 1)}
                          >
                            {t("apiKeyConnectorDialog.actions.retry")}
                          </Button>
                        </div>
                      )
                    : (
                        <p className="text-txt-tertiary text-sm">
                          {t("apiKeyConnectorDialog.verificationCode.generating")}
                        </p>
                      )}
            </div>
          )}
        </DialogContent>
        <DialogFooter
          start={
            provider?.documentationUrl
              ? (
                  <ConnectorDocumentationLink
                    url={provider.documentationUrl}
                    variant="button"
                  />
                )
              : undefined
          }
        >
          <Button
            type="submit"
            disabled={
              isConnectingAPIKey
              || (!provider?.apiKeyManaged && !apiKeyValue.trim())
              || !apiKeyExtraSettingsValid
              || !postHogAPIKeyValid
              || (isCrispManaged && !crispVerificationCode)
            }
          >
            {isConnectingAPIKey
              ? t("apiKeyConnectorDialog.actions.connecting")
              : t("apiKeyConnectorDialog.actions.connect")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
