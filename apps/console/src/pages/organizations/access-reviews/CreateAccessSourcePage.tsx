import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  PageHeader,
  useToast,
} from "@probo/ui";
import { useEffect, useMemo, useRef } from "react";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useSearchParams } from "react-router";

import type { AccessReviewPageQuery } from "#/__generated__/core/AccessReviewPageQuery.graphql";
import type { AccessReviewPageSourcesFragment$key } from "#/__generated__/core/AccessReviewPageSourcesFragment.graphql";
import type { AccessReviewPageSourcesPaginationQuery } from "#/__generated__/core/AccessReviewPageSourcesPaginationQuery.graphql";
import type { AccessSourceRowDeleteMutation } from "#/__generated__/core/AccessSourceRowDeleteMutation.graphql";
import type { CreateAccessSourceDialogMutation } from "#/__generated__/core/CreateAccessSourceDialogMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { deleteAccessSourceMutation } from "./_components/AccessSourceRow";
import { accessReviewPageQuery, sourcesPaginatedFragment } from "./AccessReviewPage";
import { createAccessSourceMutation } from "./dialogs/CreateAccessSourceDialog";

type OAuthProvider = "GOOGLE_WORKSPACE" | "LINEAR";
type ConnectedProvider = OAuthProvider | "SLACK";

const OAUTH_PROVIDERS: OAuthProvider[] = ["GOOGLE_WORKSPACE", "LINEAR"];

function providerLabel(provider: OAuthProvider): string {
  switch (provider) {
    case "GOOGLE_WORKSPACE":
      return "Google Workspace";
    case "LINEAR":
      return "Linear";
    default:
      return provider;
  }
}

function providerDescription(provider: OAuthProvider): string {
  switch (provider) {
    case "GOOGLE_WORKSPACE":
      return "Connect Google Workspace to import users and access data.";
    case "LINEAR":
      return "Connect Linear to review team and project access.";
    default:
      return "Connect this provider to sync access data.";
  }
}

function isConnectedProvider(provider: string | null): provider is ConnectedProvider {
  return provider === "GOOGLE_WORKSPACE"
    || provider === "LINEAR"
    || provider === "SLACK";
}

function defaultSourceName(provider: ConnectedProvider, __: (input: string) => string): string {
  switch (provider) {
    case "GOOGLE_WORKSPACE":
      return __("Google Workspace source");
    case "LINEAR":
      return __("Linear source");
    case "SLACK":
      return __("Slack source");
    default:
      return __("Connected source");
  }
}

export default function CreateAccessSourcePage({
  queryRef,
}: {
  queryRef: PreloadedQuery<AccessReviewPageQuery>;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [searchParams, setSearchParams] = useSearchParams();
  const processedConnectorIdRef = useRef<string | null>(null);

  usePageTitle(__("Add Access Source"));

  const { organization } = usePreloadedQuery(accessReviewPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }
  const accessReview = organization.accessReview;

  const {
    data: { accessSources },
  } = usePaginationFragment<
    AccessReviewPageSourcesPaginationQuery,
    AccessReviewPageSourcesFragment$key
  >(sourcesPaginatedFragment, accessReview!);

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

  const sourceIdsByProvider = useMemo<Record<OAuthProvider, string[]>>(
    () => ({
      GOOGLE_WORKSPACE: accessSources?.edges
        .filter(edge => edge.node.connector?.provider === "GOOGLE_WORKSPACE")
        .map(edge => edge.node.id) ?? [],
      LINEAR: accessSources?.edges
        .filter(edge => edge.node.connector?.provider === "LINEAR")
        .map(edge => edge.node.id) ?? [],
    }),
    [accessSources?.edges],
  );
  const callbackConnectorId = searchParams.get("connector_id");
  const callbackProvider = searchParams.get("provider");
  const hasSourceForCallbackConnector = !!callbackConnectorId
    && accessSources?.edges.some(edge => edge.node.connectorId === callbackConnectorId);

  useEffect(() => {
    if (!callbackConnectorId || !accessReview) {
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
          accessReviewId: accessReview.id,
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
    accessReview,
    accessSources?.__id,
    accessSources?.edges,
    callbackConnectorId,
    callbackProvider,
    createAccessSource,
    hasSourceForCallbackConnector,
    isCreatingSource,
    setSearchParams,
  ]);

  if (!accessReview?.canCreateSource) {
    return (
      <Card padded>
        <p className="text-txt-secondary text-sm">
          {__("You do not have permission to create access sources.")}
        </p>
      </Card>
    );
  }

  const connectProvider = (provider: OAuthProvider) => {
    const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseURL);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", provider);
    url.searchParams.append(
      "continue",
      `/organizations/${organizationId}/access-reviews/sources/new`,
    );
    window.location.assign(url.toString());
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

    // Keep user on the same page, but refresh data from server.
    window.location.assign(`/organizations/${organizationId}/access-reviews/sources/new`);
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
        {OAUTH_PROVIDERS.map((provider) => {
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
                      onClick={() => connectProvider(provider)}
                    >
                      {__("Connect")}
                    </Button>
                  )}
            </Card>
          );
        })}

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
    </div>
  );
}
