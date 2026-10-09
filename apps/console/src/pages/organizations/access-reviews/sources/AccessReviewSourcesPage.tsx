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

import { CaretDownIcon, MagnifyingGlassIcon, PlusIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePaginationFragment, usePreloadedQuery } from "react-relay";
import { useLocation, useSearchParams } from "react-router";

import type { AccessReviewSourcesPageFragment$key } from "#/__generated__/core/AccessReviewSourcesPageFragment.graphql";
import type { AccessReviewSourcesPagePaginationQuery } from "#/__generated__/core/AccessReviewSourcesPagePaginationQuery.graphql";
import type { AccessReviewSourcesPageQuery } from "#/__generated__/core/AccessReviewSourcesPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { AccessReviewSourceListItem } from "../_components/AccessReviewSourceListItem";

import { sourcesPage } from "./_components/variants";

function clearOAuthCallbackParams(params: URLSearchParams) {
  params.delete("connector_id");
  params.delete("provider");
  params.delete("error");
  return params;
}

export const accessReviewSourcesPageQuery = graphql`
  query AccessReviewSourcesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateSource: permission(action: "access-review:source:create")
        ...AccessReviewSourcesPageFragment
      }
    }
  }
`;

const sourcesFragment = graphql`
  fragment AccessReviewSourcesPageFragment on Organization
  @refetchable(queryName: "AccessReviewSourcesPagePaginationQuery")
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
    ) @connection(key: "AccessReviewSourcesPage_accessReviewSources", filters: []) {
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

interface AccessReviewSourcesPageProps {
  queryRef: PreloadedQuery<AccessReviewSourcesPageQuery>;
}

export function AccessReviewSourcesPage({ queryRef }: AccessReviewSourcesPageProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchQuery, setSearchQuery] = useState("");

  usePageTitle(t("accessReviewSourcesPage.title"));

  const { organization } = usePreloadedQuery<AccessReviewSourcesPageQuery>(
    accessReviewSourcesPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new NotFoundError("Organization not found");
  }

  const {
    data: { accessReviewSources },
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    AccessReviewSourcesPagePaginationQuery,
    AccessReviewSourcesPageFragment$key
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
  const callbackError = searchParams.get("error");

  useEffect(() => {
    if (!callbackError) {
      return;
    }

    toast.add({
      title: t("accessReviewSourcesPage.messages.error"),
      description: callbackError,
      type: "error",
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
    ? t("accessReviewSourcesPage.actions.loading")
    : hasFailedSearchLoad
      ? t("accessReviewSourcesPage.searchLoadFailed")
      : isSearching
        ? t("accessReviewSourcesPage.emptyConnectedSearch")
        : t("accessReviewSourcesPage.emptyConnected");
  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("accessReviewSourcesPage.title")}
          </Heading>
          <Text size={2} color="faint">
            {t("accessReviewSourcesPage.description")}
          </Text>
        </div>
        {organization.canCreateSource && (
          <ButtonLink
            to={{ pathname: "new", search: location.search }}
            variant="solid"
            iconStart={<PlusIcon />}
          >
            {t("accessReviewSourcesPage.actions.addSources")}
          </ButtonLink>
        )}
      </div>
      <div className={list()}>
        <div className={tools()}>
          <div className={search()}>
            <TextField
              icon={<MagnifyingGlassIcon />}
              value={searchQuery}
              onValueChange={setSearchQuery}
              placeholder={t("accessReviewSourcesPage.searchPlaceholder")}
              aria-label={t("accessReviewSourcesPage.searchPlaceholder")}
            />
          </div>
        </div>
        <section className={section()}>
          <div className={sectionTitle()}>
            <Heading level={2} size={3} weight="medium">
              {t("accessReviewSourcesPage.sections.connected")}
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
                  ? t("accessReviewSourcesPage.actions.retry")
                  : t("accessReviewSourcesPage.actions.loadMore")}
              </Button>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
