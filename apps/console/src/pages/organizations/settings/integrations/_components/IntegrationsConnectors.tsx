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

import { PlusIcon } from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useCallback, useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useRefetchableFragment } from "react-relay";

import type { IntegrationsConnectors_organization$key } from "#/__generated__/core/IntegrationsConnectors_organization.graphql";
import type { IntegrationsConnectors_query$key } from "#/__generated__/core/IntegrationsConnectors_query.graphql";
import type { IntegrationsConnectorsRefetchQuery } from "#/__generated__/core/IntegrationsConnectorsRefetchQuery.graphql";
import { SearchField } from "#/components/SearchField/SearchField";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { ConnectorConnectionStatus } from "#/pages/organizations/_lib/connectorStatus";

import { marketplacePath } from "../_lib/integrationPath";
import {
  connectionStatuses,
  integrationsListFilter,
  isConnectionStatus,
  useIntegrationsListFilters,
} from "../_lib/useIntegrationsListFilters";
import { useIntegrationsListSearch } from "../_lib/useIntegrationsListSearch";
import { MarketplaceThirdPartyMatches } from "../marketplace/_components/MarketplaceThirdPartyMatches";
import { MarketplaceThirdPartyNotice } from "../marketplace/_components/MarketplaceThirdPartyNotice";
import {
  connectableThirdPartyProviders,
  credentialCountByProvider,
} from "../marketplace/_lib/thirdPartyConnectorMatch";
import { integrationsList, integrationsPage } from "../variants";

import { ConnectorGroupList } from "./ConnectorGroupList";

const integrationsConnectorsFragment = graphql`
  fragment IntegrationsConnectors_query on Query {
    ...ConnectorGroupList_query
    ...MarketplaceThirdPartyMatches_query
    connectorProviders {
      provider
      displayName
    }
  }
`;

const integrationsConnectorsOrganizationFragment = graphql`
  fragment IntegrationsConnectors_organization on Organization
  @refetchable(queryName: "IntegrationsConnectorsRefetchQuery")
  @argumentDefinitions(
    filter: { type: "ConnectorFilter", defaultValue: null }
  ) {
    canCreateConnector: permission(action: "core:connector:create")
    canListThirdParties: permission(action: "core:thirdParty:list")
    # Unfiltered, so a search or status filter cannot mark a connected vendor as available.
    connectedConnectors: connectors {
      provider
    }
    ...ConnectorGroupList_organization @arguments(filter: $filter)
  }
`;

interface IntegrationsConnectorsProps {
  queryKey: IntegrationsConnectors_query$key;
  organizationKey: IntegrationsConnectors_organization$key;
}

export function IntegrationsConnectors({
  queryKey,
  organizationKey,
}: IntegrationsConnectorsProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const [matchedProviders, setMatchedProviders] = useState<ReadonlySet<string>>(() => new Set());
  const handleMatchedProviders = useCallback((next: ReadonlySet<string>) => {
    setMatchedProviders((current) => {
      if (current.size === next.size && [...next].every(provider => current.has(provider))) {
        return current;
      }

      return next;
    });
  }, []);
  const integrationsQuery = useFragment(integrationsConnectorsFragment, queryKey);
  const { query, status, setStatus } = useIntegrationsListFilters();
  const [searchInput, setSearchInput] = useIntegrationsListSearch();
  const [organization, refetch] = useRefetchableFragment<
    IntegrationsConnectorsRefetchQuery,
    IntegrationsConnectors_organization$key
  >(integrationsConnectorsOrganizationFragment, organizationKey);
  const [isPending, startTransition] = useTransition();
  const skipFirstRefetch = useRef(true);

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    startTransition(() => {
      refetch(
        { filter: integrationsListFilter(query) },
        { fetchPolicy: "store-or-network" },
      );
    });
  }, [query, refetch]);

  const isSearching = query.trim() !== "";
  const isFiltering = isSearching || status != null;
  const connectableThirdParties = connectableThirdPartyProviders(
    integrationsQuery.connectorProviders.map(provider => ({
      provider: provider.provider,
      displayName: provider.displayName,
    })),
    matchedProviders,
    credentialCountByProvider(organization.connectedConnectors.map(connector => ({
      provider: connector.provider,
    }))),
  );

  const { root, header, intro } = integrationsPage();
  const {
    root: list,
    tools,
    search,
    filters,
    filter,
    results,
    empty,
  } = integrationsList({ pending: isPending });

  return (
    <div className={root()}>
      {organization.canCreateConnector && organization.canListThirdParties && (
        <MarketplaceThirdPartyMatches
          queryKey={integrationsQuery}
          onMatched={handleMatchedProviders}
        />
      )}
      {connectableThirdParties.length > 0 && (
        <MarketplaceThirdPartyNotice
          providers={connectableThirdParties}
          description={t("listPage.suggestDescription")}
          control={(
            <ButtonLink
              to={marketplacePath(organizationId)}
              size={2}
              variant="solid"
              color="neutral"
              highContrast
              iconStart={<PlusIcon />}
            >
              {t("listPage.addMore")}
            </ButtonLink>
          )}
        />
      )}
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("listPage.title")}
          </Heading>
          <Text size={2} color="faint">
            {t("listPage.description")}
          </Text>
        </div>
      </div>
      <div className={list()}>
        <div className={tools()}>
          <div className={search()}>
            <SearchField
              value={searchInput}
              onValueChange={setSearchInput}
              placeholder={t("listPage.searchConnectorsPlaceholder")}
              aria-label={t("listPage.searchConnectorsPlaceholder")}
            />
          </div>
          <div className={filters()}>
            <div className={filter()}>
              <Select
                value={status}
                onValueChange={(value: string | null) => {
                  if (value == null) {
                    setStatus(null);
                    return;
                  }
                  if (isConnectionStatus(value)) {
                    setStatus(value);
                  }
                }}
              >
                <SelectTrigger
                  size={2}
                  placeholder={t("listPage.filters.allStatuses")}
                  aria-label={t("listPage.filters.status")}
                >
                  {(value: ConnectorConnectionStatus | null) => (
                    value != null
                      ? t(`detailsPage.status.${value}`)
                      : t("listPage.filters.allStatuses")
                  )}
                </SelectTrigger>
                <SelectPopup align="start">
                  <SelectItem value={null}>{t("listPage.filters.allStatuses")}</SelectItem>
                  {connectionStatuses.map(connectionStatus => (
                    <SelectItem key={connectionStatus} value={connectionStatus}>
                      {t(`detailsPage.status.${connectionStatus}`)}
                    </SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </div>
          </div>
        </div>
        <div className={results()} aria-busy={isPending}>
          <Suspense
            fallback={(
              <div className={empty()}>
                <Spinner size={2} aria-label={t("listPage.checkingStatuses")} />
              </div>
            )}
          >
            <ConnectorGroupList
              queryKey={integrationsQuery}
              organizationKey={organization}
              status={status}
              isFiltering={isFiltering}
            />
          </Suspense>
        </div>
      </div>
    </div>
  );
}
