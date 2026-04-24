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

import { formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  DropdownItem,
  Field,
  Input,
  Option,
  Select,
  useDialogRef,
  useToast,
  ThirdPartyLogo,
} from "@probo/ui";
import { type ReactNode, useMemo, useState } from "react";
import { useMutation } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { accessSourceMutationsCreateMutation } from "#/__generated__/core/accessSourceMutationsCreateMutation.graphql";
import type { AddAccessSourceDialogConnectorProviderInfoFragment$data } from "#/__generated__/core/AddAccessSourceDialogConnectorProviderInfoFragment.graphql";
import type { AddAccessSourceDialogCreateAPIKeyConnectorMutation } from "#/__generated__/core/AddAccessSourceDialogCreateAPIKeyConnectorMutation.graphql";
import type { AddAccessSourceDialogCreateClientCredentialsConnectorMutation } from "#/__generated__/core/AddAccessSourceDialogCreateClientCredentialsConnectorMutation.graphql";

import { createAccessSourceMutation } from "./accessSourceMutations";

export const addAccessSourceDialogConnectorProviderInfoFragment = graphql`
  fragment AddAccessSourceDialogConnectorProviderInfoFragment on ConnectorProviderInfo @relay(plural: true) {
    provider
    displayName
    oauthConfigured
    apiKeySupported
    clientCredentialsSupported
    oauth2Scopes
    extraSettings {
      key
      label
      required
    }
  }
`;

export type ProviderInfo = AddAccessSourceDialogConnectorProviderInfoFragment$data[number];

type Props = {
  children: ReactNode;
  organizationId: string;
  connectionId: string;
  providerInfos: ReadonlyArray<ProviderInfo>;
  existingSourceProviders: ReadonlyArray<string>;
};

const createAPIKeyConnectorMutation = graphql`
  mutation AddAccessSourceDialogCreateAPIKeyConnectorMutation(
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

const createClientCredentialsConnectorMutation = graphql`
  mutation AddAccessSourceDialogCreateClientCredentialsConnectorMutation(
    $input: CreateClientCredentialsConnectorInput!
  ) {
    createClientCredentialsConnector(input: $input) {
      connector {
        id
        provider
      }
    }
  }
`;

function mapAPIKeyExtraSettingToField(
  provider: string,
  settingKey: string,
): string | null {
  switch (provider) {
    case "TALLY":
      if (settingKey === "organizationId") return "tallyOrganizationId";
      break;
    case "SENTRY":
      if (settingKey === "organizationSlug") return "sentryOrganizationSlug";
      break;
    case "SUPABASE":
      if (settingKey === "organizationSlug") return "supabaseOrganizationSlug";
      break;
    case "GITHUB":
      if (settingKey === "organization") return "githubOrganization";
      break;
    case "ONE_PASSWORD":
      if (settingKey === "scimBridgeUrl") return "onePasswordScimBridgeUrl";
      break;
  }
  return null;
}

function mapClientCredentialsExtraSettingToField(
  provider: string,
  settingKey: string,
): string | null {
  switch (provider) {
    case "ONE_PASSWORD":
      if (settingKey === "accountId") return "onePasswordAccountId";
      if (settingKey === "region") return "onePasswordRegion";
      break;
  }
  return null;
}

function hasRequiredExtraSettings(
  settings: ReadonlyArray<{ readonly key: string; readonly required: boolean }>,
  values: Record<string, string>,
): boolean {
  return settings
    .filter(s => s.required)
    .every(s => values[s.key]?.trim());
}

export function AddAccessSourceDialog({
  children,
  organizationId,
  connectionId,
  providerInfos,
  existingSourceProviders,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const apiKeyDialogRef = useDialogRef();
  const clientCredentialsDialogRef = useDialogRef();

  const [searchQuery, setSearchQuery] = useState("");
  const [activeProvider, setActiveProvider] = useState<ProviderInfo | null>(null);

  const [apiKeyValue, setApiKeyValue] = useState("");
  const [extraSettingValues, setExtraSettingValues] = useState<Record<string, string>>({});
  const [isConnectingAPIKey, setIsConnectingAPIKey] = useState(false);

  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [tokenUrl, setTokenUrl] = useState("");
  const [scope, setScope] = useState("");
  const [clientCredentialsExtraValues, setClientCredentialsExtraValues] = useState<Record<string, string>>({});
  const [isConnectingClientCredentials, setIsConnectingClientCredentials] = useState(false);

  const filteredProviders = useMemo(() => {
    const sorted = [...providerInfos].sort((a, b) =>
      a.displayName.localeCompare(b.displayName),
    );
    if (!searchQuery.trim()) return sorted;
    const q = searchQuery.toLowerCase();
    return sorted.filter(
      info => info.displayName.toLowerCase().includes(q),
    );
  }, [providerInfos, searchQuery]);

  const connectedProviders = useMemo(
    () => new Set(existingSourceProviders),
    [existingSourceProviders],
  );

  const [createAccessSource]
    = useMutation<accessSourceMutationsCreateMutation>(
      createAccessSourceMutation,
    );
  const [createAPIKeyConnector]
    = useMutation<AddAccessSourceDialogCreateAPIKeyConnectorMutation>(
      createAPIKeyConnectorMutation,
    );
  const [createClientCredentialsConnector]
    = useMutation<AddAccessSourceDialogCreateClientCredentialsConnectorMutation>(
      createClientCredentialsConnectorMutation,
    );

  const connectOAuthProvider = (info: ProviderInfo) => {
    const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseURL);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", info.provider);
    for (const scope of info.oauth2Scopes) {
      url.searchParams.append("scope", scope);
    }
    url.searchParams.append(
      "continue",
      `/organizations/${organizationId}/access-reviews/sources`,
    );
    window.location.assign(url.toString());
  };

  const openAPIKeyDialog = (info: ProviderInfo) => {
    setActiveProvider(info);
    setApiKeyValue("");
    setExtraSettingValues({});
    apiKeyDialogRef.current?.open();
  };

  const openClientCredentialsDialog = (info: ProviderInfo) => {
    setActiveProvider(info);
    setClientId("");
    setClientSecret("");
    setTokenUrl("");
    setScope("");
    setClientCredentialsExtraValues({});
    clientCredentialsDialogRef.current?.open();
  };

  const createSourceAfterConnector = (
    connectorId: string,
    displayName: string,
    onDone: () => void,
  ) => {
    createAccessSource({
      variables: {
        input: {
          organizationId,
          connectorId,
          name: displayName,
          csvData: null,
        },
        connections: [connectionId],
      },
      onCompleted(_, errors) {
        onDone();
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to create access source"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Access source created successfully."),
          variant: "success",
        });
        dialogRef.current?.close();
      },
      onError(error) {
        onDone();
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create access source"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const connectAPIKeyProvider = () => {
    if (!activeProvider || !apiKeyValue.trim()) {
      return;
    }

    const requiredSettings = activeProvider.extraSettings.filter(s => s.required);
    if (!hasRequiredExtraSettings(requiredSettings, extraSettingValues)) {
      return;
    }

    setIsConnectingAPIKey(true);

    const extraFields: Record<string, string> = {};
    for (const setting of activeProvider.extraSettings) {
      const value = extraSettingValues[setting.key]?.trim();
      if (value) {
        const fieldName = mapAPIKeyExtraSettingToField(activeProvider.provider, setting.key);
        if (fieldName) {
          extraFields[fieldName] = value;
        }
      }
    }

    createAPIKeyConnector({
      variables: {
        input: {
          organizationId,
          provider: activeProvider.provider,
          apiKey: apiKeyValue.trim(),
          ...extraFields,
        },
      },
      onCompleted: (response) => {
        const connectorId = response.createAPIKeyConnector.connector.id;
        createSourceAfterConnector(
          connectorId,
          activeProvider.displayName,
          () => {
            setIsConnectingAPIKey(false);
            setApiKeyValue("");
            setExtraSettingValues({});
            setActiveProvider(null);
            apiKeyDialogRef.current?.close();
          },
        );
      },
      onError: () => {
        setIsConnectingAPIKey(false);
        toast({
          title: __("Connection failed"),
          description: __("Failed to connect provider. Please check your API key and try again."),
          variant: "error",
        });
      },
    });
  };

  const connectClientCredentialsProvider = () => {
    if (!activeProvider || !clientId.trim() || !clientSecret.trim() || !tokenUrl.trim()) {
      return;
    }

    const requiredSettings = activeProvider.extraSettings.filter(s => s.required);
    if (!hasRequiredExtraSettings(requiredSettings, clientCredentialsExtraValues)) {
      return;
    }

    setIsConnectingClientCredentials(true);

    const extraFields: Record<string, string> = {};
    for (const setting of activeProvider.extraSettings) {
      const value = clientCredentialsExtraValues[setting.key]?.trim();
      if (value) {
        const fieldName = mapClientCredentialsExtraSettingToField(
          activeProvider.provider,
          setting.key,
        );
        if (fieldName) {
          extraFields[fieldName] = value;
        }
      }
    }

    createClientCredentialsConnector({
      variables: {
        input: {
          organizationId,
          provider: activeProvider.provider,
          clientId: clientId.trim(),
          clientSecret: clientSecret.trim(),
          tokenUrl: tokenUrl.trim(),
          scope: scope.trim() || null,
          ...extraFields,
        },
      },
      onCompleted: (response) => {
        const connector = response.createClientCredentialsConnector?.connector;
        if (!connector) {
          setIsConnectingClientCredentials(false);
          toast({
            title: __("Connection failed"),
            description: __("Failed to connect provider. Please check your credentials and try again."),
            variant: "error",
          });
          return;
        }

        createSourceAfterConnector(
          connector.id,
          activeProvider.displayName,
          () => {
            setIsConnectingClientCredentials(false);
            setClientId("");
            setClientSecret("");
            setTokenUrl("");
            setScope("");
            setClientCredentialsExtraValues({});
            setActiveProvider(null);
            clientCredentialsDialogRef.current?.close();
          },
        );
      },
      onError: () => {
        setIsConnectingClientCredentials(false);
        toast({
          title: __("Connection failed"),
          description: __("Failed to connect provider. Please check your credentials and try again."),
          variant: "error",
        });
      },
    });
  };

  const renderProviderCard = (info: ProviderInfo) => {
    const isConnected = connectedProviders.has(info.provider);

    const hasSecondaryOptions = info.oauthConfigured
      && (info.apiKeySupported || info.clientCredentialsSupported);

    const renderPrimaryButton = () => {
      if (info.oauthConfigured) {
        return (
          <Button
            variant="secondary"
            onClick={() => connectOAuthProvider(info)}
          >
            {__("Connect")}
          </Button>
        );
      }
      if (info.apiKeySupported) {
        return (
          <Button
            variant="secondary"
            onClick={() => openAPIKeyDialog(info)}
          >
            {__("API Key")}
          </Button>
        );
      }
      if (info.clientCredentialsSupported) {
        return (
          <Button
            variant="secondary"
            onClick={() => openClientCredentialsDialog(info)}
          >
            {__("Client Credentials")}
          </Button>
        );
      }
      return null;
    };

    return (
      <Card key={info.provider} padded className="flex items-center gap-3">
        <ThirdPartyLogo thirdParty={info.provider} tint className="size-6 shrink-0" />
        <div className="mr-auto">
          <h3 className="font-medium">{info.displayName}</h3>
        </div>
        {isConnected
          ? (
              <Badge variant="success" size="md">
                {__("Connected")}
              </Badge>
            )
          : (
              <div className="flex items-center gap-2">
                {renderPrimaryButton()}
                {hasSecondaryOptions && (
                  <ActionDropdown variant="secondary">
                    {info.apiKeySupported && (
                      <DropdownItem
                        onSelect={() => openAPIKeyDialog(info)}
                      >
                        {__("Connect with API Key")}
                      </DropdownItem>
                    )}
                    {info.clientCredentialsSupported && (
                      <DropdownItem
                        onSelect={() => openClientCredentialsDialog(info)}
                      >
                        {__("Connect with Client Credentials")}
                      </DropdownItem>
                    )}
                  </ActionDropdown>
                )}
              </div>
            )}
      </Card>
    );
  };

  const apiKeyExtraSettingsValid = activeProvider
    ? hasRequiredExtraSettings(activeProvider.extraSettings, extraSettingValues)
    : true;

  const clientCredentialsExtraSettingsValid = activeProvider
    ? hasRequiredExtraSettings(activeProvider.extraSettings, clientCredentialsExtraValues)
    : true;

  return (
    <>
      <Dialog
        ref={dialogRef}
        trigger={children}
        title={(
          <Breadcrumb
            items={[
              __("Access Reviews"),
              __("Add Source"),
            ]}
          />
        )}
      >
        <DialogContent padded className="space-y-4">
          <Input
            placeholder={__("Search providers...")}
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
          />

          <div className="space-y-3">
            {filteredProviders.map(info => renderProviderCard(info))}

            {(!searchQuery.trim() || "csv".includes(searchQuery.toLowerCase())) && (
              <Card padded className="flex items-center gap-3">
                <div className="mr-auto">
                  <h3 className="font-medium">{__("CSV")}</h3>
                  <p className="text-sm text-txt-secondary">
                    {__("Upload CSV data directly as an access source.")}
                  </p>
                </div>
                <Button
                  variant="secondary"
                  asChild
                  onClick={() => dialogRef.current?.close()}
                >
                  <Link to={`/organizations/${organizationId}/access-reviews/sources/new/csv`}>
                    {__("Open")}
                  </Link>
                </Button>
              </Card>
            )}
          </div>
        </DialogContent>
        <DialogFooter exitLabel={__("Close")} />
      </Dialog>

      <Dialog
        ref={apiKeyDialogRef}
        title={activeProvider
          ? sprintf(__("Connect %s"), activeProvider.displayName)
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
              {sprintf(
                __("Enter the API key for %s to connect it as an access source."),
                activeProvider?.displayName ?? "",
              )}
            </p>
            <Field
              label={__("API Key")}
              type="password"
              value={apiKeyValue}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setApiKeyValue(e.target.value)}
              required
              autoFocus
            />
            {activeProvider?.extraSettings.map(setting => (
              <Field
                key={setting.key}
                label={__(setting.label)}
                value={extraSettingValues[setting.key] ?? ""}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setExtraSettingValues(prev => ({
                    ...prev,
                    [setting.key]: e.target.value,
                  }))}
                required={setting.required}
              />
            ))}
          </DialogContent>
          <DialogFooter>
            <Button
              type="submit"
              disabled={
                isConnectingAPIKey
                || !apiKeyValue.trim()
                || !apiKeyExtraSettingsValid
              }
            >
              {isConnectingAPIKey ? __("Connecting...") : __("Connect")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

      <Dialog
        ref={clientCredentialsDialogRef}
        title={activeProvider
          ? sprintf(__("Connect %s"), activeProvider.displayName)
          : __("Connect provider")}
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            connectClientCredentialsProvider();
          }}
        >
          <DialogContent padded className="space-y-4">
            <p className="text-txt-secondary text-sm">
              {sprintf(
                __("Enter the client credentials for %s to connect it as an access source."),
                activeProvider?.displayName ?? "",
              )}
            </p>
            <Field
              label={__("Client ID")}
              value={clientId}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setClientId(e.target.value)}
              required
              autoFocus
            />
            <Field
              label={__("Client Secret")}
              type="password"
              value={clientSecret}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setClientSecret(e.target.value)}
              required
            />
            <Field
              label={__("Token URL")}
              value={tokenUrl}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTokenUrl(e.target.value)}
              required
            />
            <Field
              label={__("Scope")}
              value={scope}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setScope(e.target.value)}
            />
            {activeProvider?.extraSettings.map(setting =>
              setting.key === "region"
                ? (
                    <div key={setting.key} className="space-y-1.5">
                      <label className="text-sm font-medium">{__(setting.label)}</label>
                      <Select
                        value={clientCredentialsExtraValues[setting.key] ?? ""}
                        onValueChange={(val: string) =>
                          setClientCredentialsExtraValues(prev => ({
                            ...prev,
                            [setting.key]: val,
                          }))}
                        placeholder={__("Select a region")}
                      >
                        <Option value="US">United States (US)</Option>
                        <Option value="CA">Canada (CA)</Option>
                        <Option value="EU">Europe (EU)</Option>
                      </Select>
                    </div>
                  )
                : (
                    <Field
                      key={setting.key}
                      label={__(setting.label)}
                      value={clientCredentialsExtraValues[setting.key] ?? ""}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                        setClientCredentialsExtraValues(prev => ({
                          ...prev,
                          [setting.key]: e.target.value,
                        }))}
                      required={setting.required}
                    />
                  ),
            )}
          </DialogContent>
          <DialogFooter>
            <Button
              type="submit"
              disabled={
                isConnectingClientCredentials
                || !clientId.trim()
                || !clientSecret.trim()
                || !tokenUrl.trim()
                || !clientCredentialsExtraSettingsValid
              }
            >
              {isConnectingClientCredentials ? __("Connecting...") : __("Connect")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
