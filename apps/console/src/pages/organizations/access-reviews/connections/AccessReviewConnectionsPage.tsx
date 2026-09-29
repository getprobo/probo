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

import { FileCsvIcon, MagnifyingGlassIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePaginationFragment, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import type { AccessReviewConnectionsPageFragment$key } from "#/__generated__/core/AccessReviewConnectionsPageFragment.graphql";
import type { AccessReviewConnectionsPagePaginationQuery } from "#/__generated__/core/AccessReviewConnectionsPagePaginationQuery.graphql";
import type { AccessReviewConnectionsPageQuery } from "#/__generated__/core/AccessReviewConnectionsPageQuery.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { MarketplaceEntryCard } from "#/pages/organizations/settings/integrations/_components/MarketplaceEntryCard";

import { AccessReviewSourceListItem } from "../_components/AccessReviewSourceListItem";

import { AddableConnectorListItem } from "./_components/AddableConnectorListItem";
import { sourcesPage } from "./_components/variants";
import { groupConnectorsByProvider, listedConnectorAccounts } from "./_lib/listedConnectorAccounts";

function clearOAuthCallbackParams(params: URLSearchParams) {
  params.delete("connector_id");
  params.delete("provider");
  params.delete("error");
  return params;
}

export const accessReviewConnectionsPageQuery = graphql`
  query AccessReviewConnectionsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateSource: permission(action: "access-review:source:create")
        canCreateConnector: permission(action: "core:connector:create")
        connectors {
          id
          displayName
          provider
          accounts(first: 50) {
            edges {
              node {
                name
              }
            }
          }
          ...AddableConnectorListItem_connector
        }
        ...AccessReviewConnectionsPageFragment
      }
    }
  }
`;

const sourcesFragment = graphql`
  fragment AccessReviewConnectionsPageFragment on Organization
  @refetchable(queryName: "AccessReviewConnectionsPagePaginationQuery")
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
    ) @connection(key: "AccessReviewConnectionsPage_accessReviewSources") {
      __id
      edges {
        node {
          id
          name
          connector {
            provider
          }
          ...AccessReviewSourceListItem_source
        }
      }
    }
  }
`;

interface AccessReviewConnectionsPageProps {
  queryRef: PreloadedQuery<AccessReviewConnectionsPageQuery>;
}

export function AccessReviewConnectionsPage({ queryRef }: AccessReviewConnectionsPageProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchQuery, setSearchQuery] = useState("");

  usePageTitle(t("accessReviewConnectionsPage.title"));

  const { organization } = usePreloadedQuery<AccessReviewConnectionsPageQuery>(
    accessReviewConnectionsPageQuery,
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
    AccessReviewConnectionsPagePaginationQuery,
    AccessReviewConnectionsPageFragment$key
  >(sourcesFragment, organization);

  const normalizedSearch = searchQuery.trim().toLowerCase();
  const isSearching = normalizedSearch !== "";
  // accessReviewSources has no server-side filter, so the search matches only
  // the pages already in the store. Pull the remaining ones while a search is
  // active, otherwise a source past the first page is reported as absent.
  // hasNext stays true when a page fails to load, so the failing term is
  // latched to stop the effect from retrying in a loop; the load-more button
  // stays available on that term to retry, and success clears the latch.
  const [failedSearch, setFailedSearch] = useState<string | null>(null);
  const loadMoreSources = useCallback(() => {
    loadNext(50, {
      onComplete: error => setFailedSearch(error ? normalizedSearch : null),
    });
  }, [loadNext, normalizedSearch]);
  const hasFailedSearchLoad
    = isSearching && hasNext && failedSearch === normalizedSearch;
  const isLoadingRemainingSources
    = isSearching && hasNext && !hasFailedSearchLoad;
  useEffect(() => {
    if (!isLoadingRemainingSources || isLoadingNext) return;
    loadMoreSources();
  }, [isLoadingRemainingSources, isLoadingNext, loadMoreSources]);
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
  const addableVendors = useMemo(
    () => groupConnectorsByProvider(organization.connectors).filter(group =>
      group.some(connector =>
        listedConnectorAccounts(
          connector.accounts.edges.map(({ node }) => node),
          {
            displayName: connector.displayName,
            provider: connector.provider,
          },
          normalizedSearch,
        ) != null,
      ),
    ),
    [organization.connectors, normalizedSearch],
  );
  const showCSV = !normalizedSearch
    || "csv".includes(normalizedSearch)
    || t("addAccessReviewSourceDialog.csv.title")
      .toLowerCase()
      .includes(normalizedSearch);
  const showAddMore = organization.canCreateConnector
    && (addableVendors.length > 0 || showCSV);

  const callbackError = searchParams.get("error");

  useEffect(() => {
    if (!callbackError) {
      return;
    }

    toast({
      title: t("accessReviewConnectionsPage.messages.error"),
      description: callbackError,
      variant: "error",
    });
    setSearchParams(clearOAuthCallbackParams, { replace: true });
  }, [callbackError, setSearchParams, t, toast]);

  const {
    root,
    header,
    intro,
    list,
    tools,
    search,
    section,
    sectionTitle,
    grid,
    empty,
  } = sourcesPage();
  const sourcesEmpty = isLoadingRemainingSources
    ? t("accessReviewConnectionsPage.actions.loading")
    : hasFailedSearchLoad
      ? t("accessReviewConnectionsPage.searchLoadFailed")
      : isSearching
        ? t("accessReviewConnectionsPage.emptyConnectedSearch")
        : t("accessReviewConnectionsPage.emptyConnected");
  const hasAddable = addableVendors.length > 0 || showCSV;
  const availableCount = addableVendors.length + (showCSV ? 1 : 0);

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("accessReviewConnectionsPage.title")}
          </Heading>
          <Text size={2} color="faint">
            {t("accessReviewConnectionsPage.description")}
          </Text>
        </div>
      </div>
      <div className={list()}>
        <div className={tools()}>
          <div className={search()}>
            <TextField
              icon={<MagnifyingGlassIcon />}
              value={searchQuery}
              onValueChange={setSearchQuery}
              placeholder={t("accessReviewConnectionsPage.searchPlaceholder")}
              aria-label={t("accessReviewConnectionsPage.searchPlaceholder")}
            />
          </div>
        </div>
        <section className={section()}>
          <div className={sectionTitle()}>
            <Heading level={2} size={3} weight="medium">
              {t("accessReviewConnectionsPage.sections.connected")}
            </Heading>
            <Text size={2} color="faint">{filteredSources.length}</Text>
          </div>
          {filteredSources.length === 0
            ? (
                <Card variant="soft" size={2}>
                  <div className={empty()}>
                    <Text size={2} color="faint">{sourcesEmpty}</Text>
                  </div>
                </Card>
              )
            : (
                <div className={grid()}>
                  {filteredSources.map(({ node }) => (
                    <AccessReviewSourceListItem
                      key={node.id}
                      sourceKey={node}
                      connectionId={accessReviewSources.__id}
                      organizationId={organizationId}
                    />
                  ))}
                </div>
              )}
          {hasNext && (!isSearching || hasFailedSearchLoad) && (
            <Button
              variant="soft"
              color="neutral"
              onClick={loadMoreSources}
              disabled={isLoadingNext}
              className="self-start"
            >
              {isLoadingNext
                ? t("accessReviewConnectionsPage.actions.loading")
                : hasFailedSearchLoad
                  ? t("accessReviewConnectionsPage.actions.retry")
                  : t("accessReviewConnectionsPage.actions.loadMore")}
            </Button>
          )}
        </section>
        {organization.canCreateSource && (
          <section className={section()}>
            <div className={sectionTitle()}>
              <Heading level={2} size={3} weight="medium">
                {t("accessReviewConnectionsPage.sections.addSource")}
              </Heading>
              <Text size={2} color="faint">{availableCount}</Text>
            </div>
            {!hasAddable
              ? (
                  <Card variant="soft" size={2}>
                    <div className={empty()}>
                      <Text size={2} color="faint">
                        {t("accessReviewConnectionsPage.emptyAccounts")}
                      </Text>
                    </div>
                  </Card>
                )
              : (
                  <div className={grid()}>
                    {showAddMore && (
                      <MarketplaceEntryCard organizationId={organizationId} />
                    )}
                    {addableVendors.map(connectors => (
                      <AddableConnectorListItem
                        key={connectors[0].provider}
                        connectorKeys={connectors}
                        organizationId={organizationId}
                        connectionId={accessReviewSources.__id}
                        normalizedSearch={normalizedSearch}
                      />
                    ))}
                    {showCSV && (
                      <CsvSourceCard organizationId={organizationId} />
                    )}
                  </div>
                )}
          </section>
        )}
      </div>
    </div>
  );
}

function CsvSourceCard({ organizationId }: { organizationId: string }) {
  const { t } = useTranslation();

  return (
    <TonedCard
      tone="sand"
      iconSize={14}
      icon={<FileCsvIcon className="size-8" />}
      lead={(
        <Heading level={2} size={3} weight="medium" highContrast>
          {t("addAccessReviewSourceDialog.csv.title")}
        </Heading>
      )}
      control={(
        <ButtonLink
          to={`/organizations/${organizationId}/access-reviews/connections/new/csv`}
          variant="solid"
          size={1}
        >
          {t("addAccessReviewSourceDialog.actions.open")}
        </ButtonLink>
      )}
    >
      <Text size={2} color="faint">
        {t("addAccessReviewSourceDialog.csv.description")}
      </Text>
    </TonedCard>
  );
}
