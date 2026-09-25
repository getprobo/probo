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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { formatError } from "@probo/helpers";
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
import { graphql, useMutation, usePaginationFragment, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import type { AccessReviewConnectionsPageFragment$key } from "#/__generated__/core/AccessReviewConnectionsPageFragment.graphql";
import type { AccessReviewConnectionsPagePaginationQuery } from "#/__generated__/core/AccessReviewConnectionsPagePaginationQuery.graphql";
import type { AccessReviewConnectionsPageQuery } from "#/__generated__/core/AccessReviewConnectionsPageQuery.graphql";
import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AccessReviewSourceListItem } from "../_components/AccessReviewSourceListItem";
import { createAccessReviewSourceMutation, prependCreatedSourceEdge } from "../dialogs/accessReviewSourceMutations";

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
        connectors {
          id
          displayName
          accounts(first: 50) {
            edges {
              node {
                id
                name
                externalAccountId
              }
            }
          }
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
  const addableAccounts = useMemo(() => {
    const accounts = organization.connectors.flatMap(connector =>
      connector.accounts.edges.map(({ node }) => ({
        id: node.id,
        name: node.name,
        connectorName: connector.displayName,
        needsOrganization: node.externalAccountId === connector.id,
      })),
    );

    return accounts.filter(account =>
      !normalizedSearch
      || account.name.toLowerCase().includes(normalizedSearch)
      || account.connectorName.toLowerCase().includes(normalizedSearch),
    );
  }, [organization.connectors, normalizedSearch]);
  const showCSV = !normalizedSearch
    || "csv".includes(normalizedSearch)
    || t("addAccessReviewSourceDialog.csv.title")
      .toLowerCase()
      .includes(normalizedSearch);

  const [createAccessReviewSource, isCreatingSource]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourceMutation,
    );

  const callbackError = searchParams.get("error");

  const addSource = (accountId: string, name: string) => {
    if (isCreatingSource) {
      return;
    }

    createAccessReviewSource({
      variables: {
        input: {
          organizationId,
          connectorAccountId: accountId,
          name,
          csvData: null,
        },
      },
      updater: store => prependCreatedSourceEdge(store, accessReviewSources.__id),
      onCompleted(_data, errors) {
        if (errors?.length) {
          toast({
            title: t("accessReviewConnectionsPage.messages.error"),
            description: formatError(
              t("accessReviewConnectionsPage.errors.create"),
              errors,
            ),
            variant: "error",
          });
          return;
        }

        toast({
          title: t("accessReviewConnectionsPage.messages.success"),
          description: t("accessReviewConnectionsPage.messages.created"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: t("accessReviewConnectionsPage.messages.error"),
          description: formatError(
            t("accessReviewConnectionsPage.errors.create"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

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
  const addableCount = addableAccounts.length + (showCSV ? 1 : 0);

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
          <Heading level={2} size={3} weight="medium">
            {t("accessReviewConnectionsPage.sections.connected")}
          </Heading>
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
            <Heading level={2} size={3} weight="medium">
              {t("accessReviewConnectionsPage.sections.addSource")}
            </Heading>
            {addableCount === 0
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
                    {addableAccounts.map(account => (
                      <Card key={account.id} variant="soft" size={2} className="flex h-full flex-col gap-3">
                        <Heading level={2} size={3} weight="medium" highContrast>
                          {account.connectorName}
                        </Heading>
                        <Text size={2} color="faint">
                          {account.needsOrganization
                            ? t("accessReviewConnectionsPage.needsOrganization")
                            : account.name}
                        </Text>
                        <Button
                          variant="soft"
                          color="neutral"
                          disabled={account.needsOrganization || isCreatingSource}
                          onClick={() => addSource(account.id, account.name)}
                          className="self-end"
                        >
                          {t("accessReviewConnectionsPage.actions.add")}
                        </Button>
                      </Card>
                    ))}
                    {showCSV && (
                      <Card variant="soft" size={2} className="flex h-full flex-col gap-3">
                        <Heading level={2} size={3} weight="medium" highContrast>
                          {t("addAccessReviewSourceDialog.csv.title")}
                        </Heading>
                        <Text size={2} color="faint">
                          {t("addAccessReviewSourceDialog.csv.description")}
                        </Text>
                        <ButtonLink
                          to={`/organizations/${organizationId}/access-reviews/connections/new/csv`}
                          variant="solid"
                          className="self-end"
                        >
                          {t("addAccessReviewSourceDialog.actions.open")}
                        </ButtonLink>
                      </Card>
                    )}
                  </div>
                )}
          </section>
        )}
      </div>
    </div>
  );
}
