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

import { CaretDownIcon, FileCsvIcon, MagnifyingGlassIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Fragment, useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePaginationFragment, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import type { AccessReviewConnectionsPageFragment$key } from "#/__generated__/core/AccessReviewConnectionsPageFragment.graphql";
import type { AccessReviewConnectionsPagePaginationQuery } from "#/__generated__/core/AccessReviewConnectionsPagePaginationQuery.graphql";
import type { AccessReviewConnectionsPageQuery } from "#/__generated__/core/AccessReviewConnectionsPageQuery.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { groupByProvider } from "#/pages/organizations/_lib/connectorStatus";
import { MarketplaceEntryCard } from "#/pages/organizations/settings/integrations/_components/MarketplaceEntryCard";

import { AccessReviewSourceListItem } from "../_components/AccessReviewSourceListItem";

import {
  type AddableConnectorCard,
  AddableConnectorGroups,
} from "./_components/AddableConnectorListItem";
import { sourcesPage } from "./_components/variants";

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
          provider
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
    first: { type: "Int", defaultValue: 20 }
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
      totalCount
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
    loadNext(20, {
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
  // Server total, not adjusted after create or delete. A search count
  // waits until every page has loaded.
  const connectedCount = isSearching
    ? (hasNext ? null : filteredSources.length)
    : accessReviewSources.totalCount;
  const vendorGroups = useMemo(
    () => groupByProvider(organization.connectors),
    [organization.connectors],
  );
  const showCSV = !normalizedSearch
    || "csv".includes(normalizedSearch)
    || t("addAccessReviewSourceDialog.csv.title")
      .toLowerCase()
      .includes(normalizedSearch);
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
    pager,
  } = sourcesPage();
  const sourcesEmpty = isLoadingRemainingSources
    ? t("accessReviewConnectionsPage.actions.loading")
    : hasFailedSearchLoad
      ? t("accessReviewConnectionsPage.searchLoadFailed")
      : isSearching
        ? t("accessReviewConnectionsPage.emptyConnectedSearch")
        : t("accessReviewConnectionsPage.emptyConnected");
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
            {connectedCount != null && (
              <Text size={2} color="faint">{connectedCount}</Text>
            )}
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
            <div className={pager()}>
              <Button
                type="button"
                size={2}
                variant="soft"
                color="neutral"
                loading={isLoadingNext}
                iconStart={<CaretDownIcon />}
                onClick={loadMoreSources}
              >
                {hasFailedSearchLoad
                  ? t("accessReviewConnectionsPage.actions.retry")
                  : t("accessReviewConnectionsPage.actions.loadMore")}
              </Button>
            </div>
          )}
        </section>
        {organization.canCreateSource && (
          <AddableConnectorGroups
            groups={vendorGroups}
            normalizedSearch={normalizedSearch}
            organizationId={organizationId}
            connectionId={accessReviewSources.__id}
          >
            {cards => (
              <AddSourceSection
                cards={cards}
                showCSV={showCSV}
                canCreateConnector={organization.canCreateConnector}
                organizationId={organizationId}
              />
            )}
          </AddableConnectorGroups>
        )}
      </div>
    </div>
  );
}

function AddSourceSection({
  cards,
  showCSV,
  canCreateConnector,
  organizationId,
}: {
  cards: readonly AddableConnectorCard[];
  showCSV: boolean;
  canCreateConnector: boolean;
  organizationId: string;
}) {
  const { t } = useTranslation();
  const { section, sectionTitle, grid, empty } = sourcesPage();
  const hasAddable = cards.length > 0 || showCSV;
  const showAddMore = canCreateConnector && hasAddable;
  const availableCount = cards.length + (showCSV ? 1 : 0);

  return (
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
              {cards.map(({ provider, card }) => (
                <Fragment key={provider}>{card}</Fragment>
              ))}
              {showCSV && (
                <CsvSourceCard organizationId={organizationId} />
              )}
            </div>
          )}
    </section>
  );
}

function CsvSourceCard({ organizationId }: { organizationId: string }) {
  const { t } = useTranslation();

  return (
    <TonedCard
      tone="sand"
      size={2}
      icon={<FileCsvIcon />}
      lead={(
        <Heading level={3} size={3} weight="medium" highContrast>
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
