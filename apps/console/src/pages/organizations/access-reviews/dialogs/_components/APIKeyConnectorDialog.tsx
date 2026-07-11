// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useEffect, useState } from "react";
import { useMutation, useRelayEnvironment } from "react-relay";
import { fetchQuery, graphql } from "relay-runtime";

import type { APIKeyConnectorDialogCreateAPIKeyConnectorMutation } from "#/__generated__/core/APIKeyConnectorDialogCreateAPIKeyConnectorMutation.graphql";
import type { APIKeyConnectorDialogCrispVerificationCodeQuery } from "#/__generated__/core/APIKeyConnectorDialogCrispVerificationCodeQuery.graphql";

import { useCreateAccessReviewSource } from "../_hooks/useCreateAccessReviewSource";
import {
  buildExtraFields,
  hasRequiredExtraSettings,
  mapAPIKeyExtraSettingToField,
} from "../_lib/connectorSettings";
import type { ProviderInfo } from "../AddAccessReviewSourceDialog";
import {
  isPostHogDeploymentSelected,
  PostHogDeploymentField,
} from "../PostHogDeploymentField";

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
  provider: ProviderInfo | null;
  organizationId: string;
  connectionId: string;
  onClose: () => void;
  onSuccess: () => void;
};

export function APIKeyConnectorDialog({
  provider,
  organizationId,
  connectionId,
  onClose,
  onSuccess,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const environment = useRelayEnvironment();

  const [apiKeyValue, setApiKeyValue] = useState("");
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
  }, [provider]);

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

  const connectAPIKeyProvider = () => {
    // Managed providers (Model B, e.g. Crisp) supply no customer key: the
    // server injects Probo's own credential, so only the extra settings
    // are required.
    if (!provider || (!provider.apiKeyManaged && !apiKeyValue.trim())) {
      return;
    }

    const requiredSettings = provider.extraSettings.filter(s => s.required);
    if (!hasRequiredExtraSettings(requiredSettings, extraSettingValues)) {
      return;
    }

    setIsConnectingAPIKey(true);

    const extraFields = buildExtraFields(
      provider,
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
          title: __("Connection failed"),
          // Managed providers (e.g. Crisp) never show an API key field, so
          // pointing the user at their key would be misleading; send them to
          // the settings and verification step instead.
          description: provider.apiKeyManaged
            ? __("Failed to connect provider. Please check your settings and verification code, then try again.")
            : __("Failed to connect provider. Please check your API key and try again."),
          variant: "error",
        });
      },
    });
  };

  // PostHog renders a dedicated deployment selector (Cloud region or
  // self-hosted URL); every other provider falls back to generic fields.
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

    return provider.extraSettings.map((setting) => {
      const value = extraSettingValues[setting.key] ?? "";
      return (
        <Field
          key={setting.key}
          label={__(setting.label)}
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
    ? hasRequiredExtraSettings(provider.extraSettings, extraSettingValues)
    : true;

  return (
    <Dialog
      ref={dialogRef}
      onClose={() => {
        // Reset on dismiss so the next open starts fresh (the imperative
        // close() on success does not fire onClose, so success resets inline).
        setApiKeyValue("");
        setExtraSettingValues({});
        setCrispCode(null);
        setIsConnectingAPIKey(false);
        onClose();
      }}
      title={provider
        ? sprintf(__("Connect %s"), provider.displayName)
        : __("Connect provider")}
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
              ? sprintf(
                  __("Connect %s as an access source. Provide the details below to finish."),
                  provider?.displayName ?? "",
                )
              : sprintf(
                  __("Enter the API key for %s to connect it as an access source."),
                  provider?.displayName ?? "",
                )}
          </p>
          {!provider?.apiKeyManaged && (
            <Field
              label={__("API Key")}
              type="password"
              value={apiKeyValue}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setApiKeyValue(e.target.value)}
              required
              autoFocus
            />
          )}
          {renderAPIKeyExtraSettings()}
          {isCrispManaged && (
            <div className="space-y-1.5">
              <label className="text-sm font-medium">
                {__("Verification code")}
              </label>
              <p className="text-txt-tertiary text-sm">
                {__("Paste this code into the Probo plugin settings for this website in your Crisp dashboard, then click Connect. It proves you control this website.")}
              </p>
              {!crispWebsiteId
                ? (
                    <p className="text-txt-tertiary text-sm">
                      {__("Enter your Website ID above to get your verification code.")}
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
                                title: __("Copy failed"),
                                description: __("Copy the verification code manually."),
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
                                    title: __("Copied to clipboard"),
                                    description: __("Verification code"),
                                    variant: "success",
                                  }),
                                onCopyFailure,
                              );
                            } catch {
                              onCopyFailure();
                            }
                          }}
                        >
                          {__("Copy")}
                        </Button>
                      </div>
                    )
                  : crispCodeState?.status === "error"
                    ? (
                        <div className="flex items-center gap-2">
                          <p className="text-txt-danger text-sm">
                            {__("Couldn't generate a verification code. Check the Website ID and try again.")}
                          </p>
                          <Button
                            type="button"
                            variant="secondary"
                            onClick={() => setCrispRetry(n => n + 1)}
                          >
                            {__("Retry")}
                          </Button>
                        </div>
                      )
                    : (
                        <p className="text-txt-tertiary text-sm">
                          {__("Generating code...")}
                        </p>
                      )}
            </div>
          )}
        </DialogContent>
        <DialogFooter>
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
            {isConnectingAPIKey ? __("Connecting...") : __("Connect")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
