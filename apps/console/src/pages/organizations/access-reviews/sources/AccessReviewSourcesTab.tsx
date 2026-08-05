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

import { formatError } from "@probo/helpers";
import { Button, Input, useToast } from "@probo/ui";
import { type ReactNode, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, useMutation, usePaginationFragment, usePreloadedQuery } from "react-relay";
import { Link, useSearchParams } from "react-router";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { AccessReviewSourcesTabFragment$key } from "#/__generated__/core/AccessReviewSourcesTabFragment.graphql";
import type { AccessReviewSourcesTabPaginationQuery } from "#/__generated__/core/AccessReviewSourcesTabPaginationQuery.graphql";
import type { AccessReviewSourcesTabQuery } from "#/__generated__/core/AccessReviewSourcesTabQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AccessReviewSourceListItem } from "../_components/AccessReviewSourceListItem";
import { createAccessReviewSourceMutation } from "../dialogs/accessReviewSourceMutations";

import {
  AccessReviewSourceProviderListItem,
} from "./_components/AccessReviewSourceProviderListItem";
import { accessReviewSourceSection } from "./_components/variants";

function clearOAuthCallbackParams(params: URLSearchParams) {
  params.delete("connector_id");
  params.delete("provider");
  params.delete("error");
  return params;
}

export const accessReviewSourcesTabQuery = graphql`
  query AccessReviewSourcesTabQuery($organizationId: ID!) {
    accessReviewDrivers {
      provider
      displayName
      ...AccessReviewSourceProviderListItem_provider
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateSource: permission(action: "access-review:source:create")
        ...AccessReviewSourcesTabFragment
      }
    }
  }
`;

const sourcesFragment = graphql`
  fragment AccessReviewSourcesTabFragment on Organization
  @refetchable(queryName: "AccessReviewSourcesTabPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "AccessReviewSourceOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    accessReviewSources(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "AccessReviewSourcesTab_accessReviewSources") {
      __id
      edges {
        node {
          id
          name
          connectorId
          connector {
            provider
          }
          ...AccessReviewSourceListItem_source
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<AccessReviewSourcesTabQuery>;
};

export default function AccessReviewSourcesTab({ queryRef }: Props) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchQuery, setSearchQuery] = useState("");
  const processedConnectorIdRef = useRef<string | null>(null);

  const { organization, accessReviewDrivers } = usePreloadedQuery<AccessReviewSourcesTabQuery>(
    accessReviewSourcesTabQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const {
    data: { accessReviewSources },
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    AccessReviewSourcesTabPaginationQuery,
    AccessReviewSourcesTabFragment$key
  >(sourcesFragment, organization);

  const existingSourceProviders = useMemo(
    () =>
      accessReviewSources.edges
        .map(edge => edge.node.connector?.provider)
        .filter((p): p is NonNullable<typeof p> => p != null),
    [accessReviewSources.edges],
  );
  const connectedProviderSet = useMemo(
    () => new Set(existingSourceProviders),
    [existingSourceProviders],
  );
  const normalizedSearch = searchQuery.trim().toLowerCase();
  const filteredSources = useMemo(
    () => accessReviewSources.edges.filter(({ node }) => {
      if (!normalizedSearch) return true;
      return node.name.toLowerCase().includes(normalizedSearch)
        || (node.connector?.provider
          ?.replaceAll("_", " ")
          .toLowerCase()
          .includes(normalizedSearch) ?? false);
    }),
    [accessReviewSources.edges, normalizedSearch],
  );
  const availableProviders = useMemo(
    () => accessReviewDrivers
      .filter(provider =>
        !connectedProviderSet.has(provider.provider)
        && (!normalizedSearch
          || provider.displayName.toLowerCase().includes(normalizedSearch)
          || provider.provider
            .replaceAll("_", " ")
            .toLowerCase()
            .includes(normalizedSearch)),
      )
      .sort((a, b) => a.displayName.localeCompare(b.displayName)),
    [accessReviewDrivers, connectedProviderSet, normalizedSearch],
  );
  const showCSV = !normalizedSearch
    || "csv".includes(normalizedSearch)
    || t("addAccessReviewSourceDialog.csv.title")
      .toLowerCase()
      .includes(normalizedSearch);

  const [createAccessReviewSource, isCreatingSource]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourceMutation,
    );

  // Handle OAuth callback: after the provider redirects back with connector_id,
  // automatically create the access source for that connector. Missing scopes
  // arrive as a backend error query param and are toasted like other errors.
  const callbackConnectorId = searchParams.get("connector_id");
  const callbackProvider = searchParams.get("provider");
  const callbackError = searchParams.get("error");
  const hasSourceForCallback = !!callbackConnectorId
    && accessReviewSources?.edges.some(edge => edge.node.connectorId === callbackConnectorId);

  useEffect(() => {
    if (!callbackConnectorId) return;

    if (hasSourceForCallback) {
      // Create sets processedConnectorIdRef before the mutation; when Relay
      // inserts the edge mid-callback, skip toasting here so onCompleted is
      // the only toast. Reconnect never sets that ref, so it still toasts.
      const createInFlight
        = processedConnectorIdRef.current === callbackConnectorId;
      if (callbackError && !createInFlight) {
        toast({
          title: t("accessReviewSourcesTab.messages.error"),
          description: callbackError,
          variant: "error",
        });
      }
      if (!createInFlight) {
        processedConnectorIdRef.current = null;
        setSearchParams(clearOAuthCallbackParams, { replace: true });
      }
      return;
    }

    if (processedConnectorIdRef.current === callbackConnectorId || isCreatingSource) {
      return;
    }
    processedConnectorIdRef.current = callbackConnectorId;

    const providerInfo = callbackProvider
      ? accessReviewDrivers.find(p => p.provider === callbackProvider)
      : null;
    const sourceName = providerInfo?.displayName ?? callbackProvider ?? "Source";

    createAccessReviewSource({
      variables: {
        input: {
          organizationId,
          connectorId: callbackConnectorId,
          name: sourceName,
          csvData: null,
        },
        connections: [accessReviewSources.__id],
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          processedConnectorIdRef.current = null;
          setSearchParams(clearOAuthCallbackParams, { replace: true });
          toast({
            title: t("accessReviewSourcesTab.messages.error"),
            description: formatError(
              t("accessReviewSourcesTab.errors.create"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        if (callbackError) {
          toast({
            title: t("accessReviewSourcesTab.messages.error"),
            description: callbackError,
            variant: "error",
          });
        } else {
          toast({
            title: t("accessReviewSourcesTab.messages.success"),
            description: t("accessReviewSourcesTab.messages.created"),
            variant: "success",
          });
        }
        processedConnectorIdRef.current = null;
        setSearchParams(clearOAuthCallbackParams, { replace: true });
      },
      onError(error) {
        processedConnectorIdRef.current = null;
        setSearchParams(clearOAuthCallbackParams, { replace: true });
        toast({
          title: t("accessReviewSourcesTab.messages.error"),
          description: formatError(
            t("accessReviewSourcesTab.errors.create"),
            error,
          ),
          variant: "error",
        });
      },
    });
  }, [
    callbackConnectorId,
    callbackProvider,
    callbackError,
    accessReviewDrivers,
    createAccessReviewSource,
    hasSourceForCallback,
    isCreatingSource,
    organizationId,
    accessReviewSources.__id,
    setSearchParams,
    toast,
    t,
  ]);

  return (
    <div className="flex flex-col gap-8">
      <Input
        value={searchQuery}
        onChange={event => setSearchQuery(event.target.value)}
        placeholder={t("accessReviewSourcesTab.searchPlaceholder")}
        className="max-w-sm"
      />

      <SourceSection
        title={t("accessReviewSourcesTab.sections.connected")}
        count={filteredSources.length}
        empty={normalizedSearch
          ? t("accessReviewSourcesTab.emptyConnectedSearch")
          : t("accessReviewSourcesTab.emptyConnected")}
      >
        {filteredSources.map(({ node }) => (
          <AccessReviewSourceListItem
            key={node.id}
            sourceKey={node}
            connectionId={accessReviewSources.__id}
            organizationId={organizationId}
          />
        ))}
      </SourceSection>

      {hasNext && (
        <Button
          variant="secondary"
          onClick={() => loadNext(50)}
          disabled={isLoadingNext}
          className="self-start"
        >
          {isLoadingNext
            ? t("accessReviewSourcesTab.actions.loading")
            : t("accessReviewSourcesTab.actions.loadMore")}
        </Button>
      )}

      {organization.canCreateSource && (
        <SourceSection
          title={t("accessReviewSourcesTab.sections.notConnected")}
          count={availableProviders.length + (showCSV ? 1 : 0)}
          empty={t("accessReviewSourcesTab.emptyAvailableSearch")}
        >
          {availableProviders.map(provider => (
            <AccessReviewSourceProviderListItem
              key={provider.provider}
              providerKey={provider}
              organizationId={organizationId}
              connectionId={accessReviewSources.__id}
            />
          ))}
          {showCSV && (
            <CSVSourceListItem organizationId={organizationId} />
          )}
        </SourceSection>
      )}
    </div>
  );
}

function SourceSection({
  title,
  count,
  empty,
  children,
}: {
  title: string;
  count: number;
  empty: string;
  children: ReactNode;
}) {
  const { root, header, title: titleClass, count: countClass, list, item }
    = accessReviewSourceSection();

  return (
    <section className={root()}>
      <div className={header()}>
        <h2 className={titleClass()}>{title}</h2>
        <span className={countClass()}>{count}</span>
      </div>
      <ul className={list()}>
        {count > 0
          ? children
          : (
              <li className={item()}>
                <span className="text-sm text-txt-tertiary">{empty}</span>
              </li>
            )}
      </ul>
    </section>
  );
}

function CSVSourceListItem({ organizationId }: { organizationId: string }) {
  const { t } = useTranslation();
  const { item, content, trailing } = accessReviewSourceSection();

  return (
    <li className={item()}>
      <div className={content()}>
        <span className="text-sm font-medium text-txt-primary">
          {t("addAccessReviewSourceDialog.csv.title")}
        </span>
        <span className="text-xs text-txt-tertiary">
          {t("addAccessReviewSourceDialog.csv.description")}
        </span>
      </div>
      <div className={trailing()}>
        <Button variant="primary" asChild>
          <Link
            to={`/organizations/${organizationId}/access-reviews/sources/new/csv`}
          >
            {t("addAccessReviewSourceDialog.actions.open")}
          </Link>
        </Button>
      </div>
    </li>
  );
}
