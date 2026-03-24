import { sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  PageHeader,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";
import { Link, useSearchParams } from "react-router";

import type { CreateAccessSourcePageQuery } from "#/__generated__/core/CreateAccessSourcePageQuery.graphql";
import type { CreateAccessSourcePageSourcesFragment$key } from "#/__generated__/core/CreateAccessSourcePageSourcesFragment.graphql";
import type { CreateAccessSourcePageSourcesPaginationQuery } from "#/__generated__/core/CreateAccessSourcePageSourcesPaginationQuery.graphql";
import type { AccessSourceRowDeleteMutation } from "#/__generated__/core/AccessSourceRowDeleteMutation.graphql";
import type { CreateAccessSourceDialogMutation } from "#/__generated__/core/CreateAccessSourceDialogMutation.graphql";
import type { CreateAccessSourcePageCreateAPIKeyConnectorMutation } from "#/__generated__/core/CreateAccessSourcePageCreateAPIKeyConnectorMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { deleteAccessSourceMutation } from "./_components/AccessSourceRow";
import { createAccessSourceMutation } from "./dialogs/CreateAccessSourceDialog";

export const createAccessSourcePageQuery = graphql`
  query CreateAccessSourcePageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateSource: permission(action: "core:access-source:create")
        ...CreateAccessSourcePageSourcesFragment
      }
    }
  }
`;

export const createAccessSourcePageSourcesFragment = graphql`
  fragment CreateAccessSourcePageSourcesFragment on Organization
  @refetchable(queryName: "CreateAccessSourcePageSourcesPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "AccessSourceOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    accessSources(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "CreateAccessSourcePage_accessSources") {
      __id
      edges {
        node {
          id
          name
          connectorId
          connector {
            provider
          }
        }
      }
    }
  }
`;

const createAPIKeyConnectorMutation = graphql`
  mutation CreateAccessSourcePageCreateAPIKeyConnectorMutation(
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

type OAuthProvider = "GOOGLE_WORKSPACE" | "LINEAR" | "SLACK";
type APIKeyProvider = "BREX" | "TALLY" | "CLOUDFLARE";
type SourceProvider = OAuthProvider | APIKeyProvider;

const OAUTH_PROVIDERS: OAuthProvider[] = ["GOOGLE_WORKSPACE", "LINEAR", "SLACK"];
const API_KEY_PROVIDERS: APIKeyProvider[] = ["BREX", "TALLY", "CLOUDFLARE"];

function providerLabel(provider: SourceProvider): string {
  switch (provider) {
    case "GOOGLE_WORKSPACE":
      return "Google Workspace";
    case "LINEAR":
      return "Linear";
    case "SLACK":
      return "Slack";
    case "BREX":
      return "Brex";
    case "TALLY":
      return "Tally";
    case "CLOUDFLARE":
      return "Cloudflare";
    default:
      return provider;
  }
}

function providerDescription(provider: SourceProvider): string {
  switch (provider) {
    case "GOOGLE_WORKSPACE":
      return "Connect Google Workspace to import users and access data.";
    case "LINEAR":
      return "Connect Linear to review team and project access.";
    case "SLACK":
      return "Connect Slack to review workspace member access.";
    case "BREX":
      return "Connect Brex to review financial platform access.";
    case "TALLY":
      return "Connect Tally to review form and workspace access.";
    case "CLOUDFLARE":
      return "Connect Cloudflare to review infrastructure access.";
    default:
      return "Connect this provider to sync access data.";
  }
}

function isConnectedProvider(provider: string | null): provider is SourceProvider {
  return provider === "GOOGLE_WORKSPACE"
    || provider === "LINEAR"
    || provider === "SLACK"
    || provider === "BREX"
    || provider === "TALLY"
    || provider === "CLOUDFLARE";
}

function defaultSourceName(provider: SourceProvider, __: (input: string) => string): string {
  switch (provider) {
    case "GOOGLE_WORKSPACE":
      return __("Google Workspace source");
    case "LINEAR":
      return __("Linear source");
    case "SLACK":
      return __("Slack source");
    case "BREX":
      return __("Brex source");
    case "TALLY":
      return __("Tally source");
    case "CLOUDFLARE":
      return __("Cloudflare source");
    default:
      return __("Connected source");
  }
}

export default function CreateAccessSourcePage({
  queryRef,
}: {
  queryRef: PreloadedQuery<CreateAccessSourcePageQuery>;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [searchParams, setSearchParams] = useSearchParams();
  const processedConnectorIdRef = useRef<string | null>(null);
  const apiKeyDialogRef = useDialogRef();
  const [apiKeyProvider, setApiKeyProvider] = useState<APIKeyProvider | null>(null);
  const [apiKeyValue, setApiKeyValue] = useState("");
  const [providerSettingValue, setProviderSettingValue] = useState("");
  const [isConnectingAPIKey, setIsConnectingAPIKey] = useState(false);

  usePageTitle(__("Add Access Source"));

  const { organization } = usePreloadedQuery(createAccessSourcePageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const {
    data: { accessSources },
  } = usePaginationFragment<
    CreateAccessSourcePageSourcesPaginationQuery,
    CreateAccessSourcePageSourcesFragment$key
  >(createAccessSourcePageSourcesFragment, organization);

  const [deleteAccessSource, isDeletingSource]
    = useMutationWithToasts<AccessSourceRowDeleteMutation>(
      deleteAccessSourceMutation,
      {
        successMessage: __("Access source disconnected successfully."),
        errorMessage: __("Failed to disconnect source"),
      },
    );
  const [createAccessSource, isCreatingSource]
    = useMutationWithToasts<CreateAccessSourceDialogMutation>(
      createAccessSourceMutation,
      {
        successMessage: __("Access source created successfully."),
        errorMessage: __("Failed to create access source"),
      },
    );
  const [createAPIKeyConnector]
    = useMutation<CreateAccessSourcePageCreateAPIKeyConnectorMutation>(
      createAPIKeyConnectorMutation,
    );

  const sourceIdsByProvider = useMemo<Record<SourceProvider, string[]>>(
    () => {
      const result: Record<string, string[]> = {};
      for (const p of [...OAUTH_PROVIDERS, ...API_KEY_PROVIDERS]) {
        result[p] = accessSources?.edges
          .filter(edge => edge.node.connector?.provider === p)
          .map(edge => edge.node.id) ?? [];
      }
      return result as Record<SourceProvider, string[]>;
    },
    [accessSources?.edges],
  );
  const callbackConnectorId = searchParams.get("connector_id");
  const callbackProvider = searchParams.get("provider");
  const hasSourceForCallbackConnector = !!callbackConnectorId
    && accessSources?.edges.some(edge => edge.node.connectorId === callbackConnectorId);

  useEffect(() => {
    if (!callbackConnectorId) {
      return;
    }

    if (hasSourceForCallbackConnector) {
      setSearchParams((params) => {
        params.delete("connector_id");
        return params;
      }, { replace: true });
      return;
    }

    if (processedConnectorIdRef.current === callbackConnectorId || isCreatingSource) {
      return;
    }
    processedConnectorIdRef.current = callbackConnectorId;

    void createAccessSource({
      variables: {
        input: {
          organizationId,
          connectorId: callbackConnectorId,
          name: isConnectedProvider(callbackProvider)
            ? defaultSourceName(callbackProvider, __)
            : __("Connected source"),
          csvData: null,
        },
        connections: accessSources?.__id ? [accessSources.__id] : [],
      },
      onCompleted: () => {
        setSearchParams((params) => {
          params.delete("connector_id");
          return params;
        }, { replace: true });
      },
    });
  }, [
    __,
    accessSources?.__id,
    accessSources?.edges,
    callbackConnectorId,
    callbackProvider,
    createAccessSource,
    hasSourceForCallbackConnector,
    isCreatingSource,
    organizationId,
    setSearchParams,
  ]);

  if (!organization.canCreateSource) {
    return (
      <Card padded>
        <p className="text-txt-secondary text-sm">
          {__("You do not have permission to create access sources.")}
        </p>
      </Card>
    );
  }

  const connectOAuthProvider = (provider: OAuthProvider, extraParams?: Record<string, string>) => {
    const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseURL);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", provider);
    url.searchParams.append(
      "continue",
      `/organizations/${organizationId}/access-reviews/sources/new`,
    );
    if (extraParams) {
      for (const [key, value] of Object.entries(extraParams)) {
        url.searchParams.append(key, value);
      }
    }
    window.location.assign(url.toString());
  };

  const openAPIKeyDialog = (provider: APIKeyProvider) => {
    setApiKeyProvider(provider);
    setApiKeyValue("");
    setProviderSettingValue("");
    apiKeyDialogRef.current?.open();
  };

  const providerNeedsExtraSetting = (provider: APIKeyProvider | null): boolean => {
    return provider === "TALLY";
  };

  const connectAPIKeyProvider = () => {
    if (!apiKeyProvider || !apiKeyValue.trim()) {
      return;
    }

    if (providerNeedsExtraSetting(apiKeyProvider) && !providerSettingValue.trim()) {
      return;
    }

    setIsConnectingAPIKey(true);

    const extraSettings: Record<string, string> = {};
    if (apiKeyProvider === "TALLY" && providerSettingValue.trim()) {
      extraSettings.tallyOrganizationId = providerSettingValue.trim();
    }
    createAPIKeyConnector({
      variables: {
        input: {
          organizationId,
          provider: apiKeyProvider,
          apiKey: apiKeyValue.trim(),
          ...extraSettings,
        },
      },
      onCompleted: (response) => {
        const connectorId = response.createAPIKeyConnector.connector.id;

        void createAccessSource({
          variables: {
            input: {
              organizationId,
              connectorId,
              name: defaultSourceName(apiKeyProvider, __),
              csvData: null,
            },
            connections: accessSources?.__id ? [accessSources.__id] : [],
          },
          onCompleted: () => {
            setIsConnectingAPIKey(false);
            setApiKeyValue("");
            setProviderSettingValue("");
            setApiKeyProvider(null);
            apiKeyDialogRef.current?.close();
          },
          onError: () => {
            setIsConnectingAPIKey(false);
          },
        });
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

  const disconnectProviderSource = async (sourceIds: string[]) => {
    if (sourceIds.length === 0) {
      toast({
        title: __("Nothing to disconnect"),
        description: __("No access source is currently linked to this provider."),
      });
      return;
    }

    for (const sourceId of sourceIds) {
      await deleteAccessSource({
        variables: {
          input: {
            accessSourceId: sourceId,
          },
          connections: [accessSources.__id],
        },
      });
    }

    window.location.assign(`/organizations/${organizationId}/access-reviews/sources/new`);
  };

  const renderProviderCard = (
    provider: SourceProvider,
    onConnect: () => void,
  ) => {
    const sourceIds = sourceIdsByProvider[provider];
    const isConnected = sourceIds.length > 0;

    return (
      <Card key={provider} padded className="flex items-center gap-3">
        <div className="mr-auto">
          <h3 className="font-medium">{providerLabel(provider)}</h3>
          <p className="text-sm text-txt-secondary">
            {__(providerDescription(provider))}
          </p>
        </div>
        {isConnected
          ? (
              <div className="flex items-center gap-2">
                <Badge variant="success" size="md">
                  {__("Connected")}
                </Badge>
                <Button
                  variant="danger"
                  disabled={isDeletingSource}
                  onClick={() => void disconnectProviderSource(sourceIds)}
                >
                  {__("Disconnect")}
                </Button>
              </div>
            )
          : (
              <Button
                variant="secondary"
                onClick={onConnect}
              >
                {__("Connect")}
              </Button>
            )}
      </Card>
    );
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Add access source")}
        description={__(
          "Connect a provider or upload CSV data to add an access source.",
        )}
      />

      <div className="space-y-3">
        {OAUTH_PROVIDERS.map(provider =>
          renderProviderCard(
            provider,
            () => connectOAuthProvider(provider),
          ),
        )}
        {API_KEY_PROVIDERS.map(provider =>
          renderProviderCard(provider, () => openAPIKeyDialog(provider)),
        )}

        <Card padded className="flex items-center gap-3">
          <div className="mr-auto">
            <h3 className="font-medium">{__("CSV")}</h3>
            <p className="text-sm text-txt-secondary">
              {__("Upload CSV data directly as an access source.")}
            </p>
          </div>
          <Button variant="secondary" asChild>
            <Link to={`/organizations/${organizationId}/access-reviews/sources/new/csv`}>
              {__("Open")}
            </Link>
          </Button>
        </Card>
      </div>

      <Dialog
        ref={apiKeyDialogRef}
        title={apiKeyProvider ? sprintf(__("Connect %s"), providerLabel(apiKeyProvider)) : __("Connect provider")}
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            connectAPIKeyProvider();
          }}
        >
          <DialogContent padded className="space-y-4">
            <p className="text-txt-secondary text-sm">
              {sprintf(__("Enter the API key for %s to connect it as an access source."), apiKeyProvider ? providerLabel(apiKeyProvider) : "")}
            </p>
            <Field
              label={__("API Key")}
              type="password"
              value={apiKeyValue}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setApiKeyValue(e.target.value)}
              required
              autoFocus
            />
            {apiKeyProvider === "TALLY" && (
              <Field
                label={__("Organization ID")}
                value={providerSettingValue}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setProviderSettingValue(e.target.value)}
                required
              />
            )}
          </DialogContent>
          <DialogFooter>
            <Button
              type="submit"
              disabled={
                isConnectingAPIKey
                || !apiKeyValue.trim()
                || (providerNeedsExtraSetting(apiKeyProvider) && !providerSettingValue.trim())
              }
            >
              {isConnectingAPIKey ? __("Connecting...") : __("Connect")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

    </div>
  );
}
