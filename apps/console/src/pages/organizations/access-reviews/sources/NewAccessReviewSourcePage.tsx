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

import { CaretLeftIcon, MagnifyingGlassIcon, PlugIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation, useNavigate } from "react-router";
import { ConnectionHandler } from "relay-runtime";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { NewAccessReviewSourcePageQuery } from "#/__generated__/core/NewAccessReviewSourcePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";
import { groupByProvider } from "#/pages/organizations/_lib/connectorStatus";
import { marketplacePath } from "#/pages/organizations/settings/integrations/_lib/integrationPath";

import {
  createAccessReviewSourcesMutation,
  prependCreatedSourceEdges,
} from "../dialogs/accessReviewSourceMutations";

import { AddableConnectorGroups } from "./_components/AddableConnectorGroups";
import { AddSourceSection } from "./_components/AddSourceSection";
import type { AddableConnectorCard } from "./_components/ResolvedAddableConnectorCard";
import { sourcesPage } from "./_components/variants";

// The page query spreads the connector fragment owned by the list item.
import "./_components/AddableConnectorListItem";

export const newAccessReviewSourcePageQuery = graphql`
  query NewAccessReviewSourcePageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateSource: permission(action: "access-review:source:create")
        canCreateConnector: permission(action: "core:connector:create")
        connectors {
          id
          provider
          ...AddableConnectorListItem_connector
        }
      }
    }
  }
`;

interface NewAccessReviewSourcePageProps {
  queryRef: PreloadedQuery<NewAccessReviewSourcePageQuery>;
}

export function NewAccessReviewSourcePage({ queryRef }: NewAccessReviewSourcePageProps) {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const toast = useToast();
  const organizationId = useOrganizationId();
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedProviders, setSelectedProviders] = useState<ReadonlySet<string>>(
    () => new Set(),
  );
  const title = t("newAccessReviewSourcePage.title");
  const {
    root,
    back,
    header,
    intro,
    actions,
    list,
    search,
    sectionTitle,
    empty,
  } = sourcesPage();
  const sourcesPath = `/organizations/${organizationId}/access-reviews/sources`;

  usePageTitle(title);

  const { organization } = usePreloadedQuery<NewAccessReviewSourcePageQuery>(
    newAccessReviewSourcePageQuery,
    queryRef,
  );
  const [createAccessReviewSources, isCreating]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourcesMutation,
    );
  const vendorGroups = useMemo(() => {
    if (organization.__typename !== "Organization") {
      return [];
    }
    return groupByProvider(
      organization.connectors.map(connector => ({
        provider: connector.provider,
        connector,
      })),
    ).map(group => group.map(entry => entry.connector));
  }, [organization]);
  if (organization.__typename !== "Organization") {
    throw new NotFoundError("Organization not found");
  }

  const normalizedSearch = searchQuery.trim().toLowerCase();
  const isSearching = normalizedSearch !== "";
  const showCSV = !normalizedSearch
    || "csv".includes(normalizedSearch)
    || t("addAccessReviewSourceDialog.csv.title")
      .toLowerCase()
      .includes(normalizedSearch);
  const connectionId = ConnectionHandler.getConnectionID(
    organization.id,
    "AccessReviewSourcesPage_accessReviewSources",
  );

  function selectProvider(provider: string, selected: boolean) {
    setSelectedProviders((current) => {
      const next = new Set(current);
      if (selected) {
        next.add(provider);
      } else {
        next.delete(provider);
      }
      return next;
    });
  }

  async function addSelected(cards: readonly AddableConnectorCard[]) {
    const accounts = cards
      .filter(card => card.selectable && selectedProviders.has(card.provider))
      .flatMap(card => card.accounts);
    if (isCreating || accounts.length === 0) {
      return;
    }

    try {
      await createAccessReviewSources({
        variables: {
          input: {
            organizationId,
            sources: accounts.map(account => ({
              connectorAccountId: account.id,
              name: account.name,
              connectorId: null,
              csvData: null,
            })),
          },
        },
        updater: (store) => {
          if (connectionId) {
            prependCreatedSourceEdges(store, connectionId);
          }
        },
      }, {
        errorToast: t("accessReviewSourcesPage.errors.create"),
      });
      toast.add({
        title: t("accessReviewSourcesPage.messages.created"),
        type: "success",
      });
      void navigate({ pathname: sourcesPath, search: location.search });
    } catch {
      // The mutation hook already reported the error.
    }
  }

  function pageIntro(count: number | null) {
    return (
      <div className={intro()}>
        <div className={sectionTitle()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {title}
          </Heading>
          {count != null && (
            <Text size={2} color="faint">{count}</Text>
          )}
        </div>
        <Text size={2} color="faint">
          {t("newAccessReviewSourcePage.description")}
        </Text>
      </div>
    );
  }

  return (
    <div className={root()}>
      <Link
        to={{ pathname: sourcesPath, search: location.search }}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("accessReviewSourcesPage.title")}
      </Link>
      {organization.canCreateSource
        ? (
            <AddableConnectorGroups
              groups={vendorGroups}
              normalizedSearch={normalizedSearch}
              selectedProviders={selectedProviders}
              onSelectedChange={selectProvider}
            >
              {(cards) => {
                const selectedAccountCount = cards
                  .filter(card => card.selectable && selectedProviders.has(card.provider))
                  .reduce((sum, card) => sum + card.accounts.length, 0);
                return (
                  <>
                    <div className={header()}>
                      {pageIntro(cards.length + (showCSV ? 1 : 0))}
                      <div className={actions()}>
                        {organization.canCreateConnector && (
                          <ButtonLink
                            to={marketplacePath(organizationId)}
                            variant="soft"
                            iconStart={<PlugIcon />}
                          >
                            {t("newAccessReviewSourcePage.actions.new")}
                          </ButtonLink>
                        )}
                        <Button
                          variant="solid"
                          disabled={selectedAccountCount === 0}
                          loading={isCreating}
                          onClick={() => {
                            void addSelected(cards);
                          }}
                        >
                          {selectedAccountCount === 0
                            ? t("newAccessReviewSourcePage.actions.add")
                            : t("newAccessReviewSourcePage.actions.addCount", {
                                count: selectedAccountCount,
                              })}
                        </Button>
                      </div>
                    </div>
                    <div className={list()}>
                      <div className={search()}>
                        <TextField
                          icon={<MagnifyingGlassIcon />}
                          value={searchQuery}
                          onValueChange={setSearchQuery}
                          placeholder={t("newAccessReviewSourcePage.searchPlaceholder")}
                          aria-label={t("newAccessReviewSourcePage.searchPlaceholder")}
                        />
                      </div>
                      <AddSourceSection
                        cards={cards}
                        showCSV={showCSV}
                        isSearching={isSearching}
                      />
                    </div>
                  </>
                );
              }}
            </AddableConnectorGroups>
          )
        : (
            <>
              {pageIntro(null)}
              <Card variant="soft" size={2}>
                <div className={empty()}>
                  <Text size={2} color="faint">
                    {t("newAccessReviewSourcePage.permissionDenied")}
                  </Text>
                </div>
              </Card>
            </>
          )}
    </div>
  );
}
