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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useRefetchableFragment } from "react-relay";

import type { IntegrationsConnectors_organization$key } from "#/__generated__/core/IntegrationsConnectors_organization.graphql";
import type { IntegrationsConnectors_query$key } from "#/__generated__/core/IntegrationsConnectors_query.graphql";
import type { IntegrationsConnectorsRefetchQuery } from "#/__generated__/core/IntegrationsConnectorsRefetchQuery.graphql";
import {
  type ConnectorConnectionStatus,
  groupByProvider,
} from "#/pages/organizations/_lib/connectorStatus";

import {
  connectionStatuses,
  integrationsListFilter,
  isConnectionStatus,
  useIntegrationsListFilters,
} from "../_lib/useIntegrationsListFilters";
import { useIntegrationsListSearch } from "../_lib/useIntegrationsListSearch";
import { integrationsList, integrationsPage } from "../variants";

import { ConnectorGroupListItem } from "./ConnectorGroupListItem";
import { MarketplaceEntryCard } from "./MarketplaceEntryCard";

const integrationsConnectorsFragment = graphql`
  fragment IntegrationsConnectors_query on Query {
    connectorProviders {
      provider
      ...ConnectorGroupListItem_provider
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
    connectors(filter: $filter) {
      id
      provider
      connectionStatus
      ...ConnectorGroupListItem_connector
    }
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
  const { connectorProviders } = useFragment(integrationsConnectorsFragment, queryKey);
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
  const connectors = organization.connectors.filter(connector =>
    status == null || connector.connectionStatus === status,
  );

  const { root, header, intro } = integrationsPage();
  const {
    root: list,
    tools,
    search,
    filters,
    filter,
    results,
    section,
    sectionTitle,
    grid,
    empty,
  } = integrationsList({ pending: isPending });

  return (
    <div className={root()}>
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
            <TextField
              icon={<MagnifyingGlassIcon />}
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
          <section className={section()}>
            <div className={sectionTitle()}>
              <Heading level={2} size={3} weight="medium">
                {t("listPage.sections.connected")}
              </Heading>
              <Text size={2} color="faint">{connectors.length}</Text>
            </div>
            {connectors.length === 0 && isFiltering
              ? (
                  <Card variant="soft" size={2}>
                    <div className={empty()}>
                      <Text size={2} color="faint">
                        {t("listPage.emptyConnectedSearch")}
                      </Text>
                    </div>
                  </Card>
                )
              : (
                  <div className={grid()}>
                    {organization.canCreateConnector && (
                      <MarketplaceEntryCard />
                    )}
                    {groupByProvider(connectors).map((group) => {
                      const face = group[0];
                      const providerKey = connectorProviders.find(
                        driver => driver.provider === face.provider,
                      );

                      return (
                        <ConnectorGroupListItem
                          key={group.length > 1 ? face.provider : face.id}
                          connectorKeys={group}
                          providerKey={providerKey}
                          canConnect={organization.canCreateConnector}
                        />
                      );
                    })}
                    {connectors.length === 0 && !organization.canCreateConnector && (
                      <Card variant="soft" size={2}>
                        <div className={empty()}>
                          <Text size={2} color="faint">
                            {t("listPage.emptyConnected")}
                          </Text>
                        </div>
                      </Card>
                    )}
                  </div>
                )}
          </section>
        </div>
      </div>
    </div>
  );
}
