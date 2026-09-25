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

import { MagnifyingGlassIcon, PlusIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import type { IntegrationsPageQuery } from "#/__generated__/core/IntegrationsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { ConnectorListItem } from "./_components/ConnectorListItem";
import { ConnectorProviderStack } from "./_components/ConnectorProviderStack";
import { MarketplaceEntryCard } from "./_components/MarketplaceEntryCard";
import { marketplacePath } from "./_lib/integrationPath";
import { integrationsList, integrationsPage } from "./variants";

const connectionStatuses = [
  "CONNECTED",
  "DISCONNECTED",
  "NOT_AUTHORIZED",
  "RECONNECT_REQUIRED",
] as const;

type ConnectionStatus = (typeof connectionStatuses)[number];

function isConnectionStatus(value: string): value is ConnectionStatus {
  return (connectionStatuses as readonly string[]).includes(value);
}

function groupByProvider<T extends { provider: string }>(connectors: readonly T[]): T[][] {
  const groups: T[][] = [];
  const indexByProvider = new Map<string, number>();

  for (const connector of connectors) {
    const index = indexByProvider.get(connector.provider);
    if (index == null) {
      indexByProvider.set(connector.provider, groups.length);
      groups.push([connector]);
      continue;
    }

    groups[index].push(connector);
  }

  return groups;
}

export const integrationsPageQuery = graphql`
  query IntegrationsPageQuery($organizationId: ID!) {
    accessReviewDrivers {
      provider
      displayName
      ...AccessReviewSourceProviderListItem_provider
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateConnector: permission(action: "core:connector:create")
        connectors {
          id
          provider
          displayName
          connectionStatus
          ...ConnectorListItem_connector
        }
      }
    }
  }
`;

interface IntegrationsPageProps {
  queryRef: PreloadedQuery<IntegrationsPageQuery>;
}

export function IntegrationsPage({ queryRef }: IntegrationsPageProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { toast } = useToast();
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchQuery, setSearchQuery] = useState("");
  const [status, setStatus] = useState<ConnectionStatus | null>(null);
  const organizationId = useOrganizationId();
  const { organization, accessReviewDrivers }
    = usePreloadedQuery<IntegrationsPageQuery>(integrationsPageQuery, queryRef);

  usePageTitle(t("listPage.title"));

  const callbackConnectorId = searchParams.get("connector_id");
  const callbackError = searchParams.get("error");

  useEffect(() => {
    if (callbackConnectorId) {
      if (callbackError) {
        toast({
          title: t("listPage.messages.error"),
          description: callbackError,
          variant: "error",
        });
      }

      setSearchParams((params) => {
        params.delete("connector_id");
        params.delete("provider");
        params.delete("error");
        return params;
      }, { replace: true });
      return;
    }

    if (callbackError) {
      toast({
        title: t("listPage.messages.error"),
        description: callbackError,
        variant: "error",
      });
      setSearchParams((params) => {
        params.delete("error");
        return params;
      }, { replace: true });
    }
  }, [
    callbackConnectorId,
    callbackError,
    setSearchParams,
    t,
    toast,
  ]);

  if (organization.__typename !== "Organization") {
    throw new NotFoundError(t("listPage.notFound"));
  }

  const normalizedSearch = searchQuery.trim().toLowerCase();
  const isSearching = normalizedSearch !== "";
  const matchesSearch = (displayName: string, provider: string) => {
    if (!isSearching) {
      return true;
    }

    return displayName.toLowerCase().includes(normalizedSearch)
      || provider.replaceAll("_", " ").toLowerCase().includes(normalizedSearch);
  };
  const isFiltering = isSearching || status != null;
  const connectors = organization.connectors.filter(connector =>
    matchesSearch(connector.displayName, connector.provider)
    && (status == null || connector.connectionStatus === status),
  );

  const { root, header, intro } = integrationsPage();
  const { root: list, tools, search, filters, filter, section, sectionTitle, grid, empty } = integrationsList();

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
        {organization.canCreateConnector && (
          <ButtonLink
            to={marketplacePath(organizationId)}
            variant="solid"
            iconStart={<PlusIcon />}
          >
            {t("listPage.actions.add")}
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
                  {(value: ConnectionStatus | null) => (
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
                  {groupByProvider(connectors).map((group) => {
                    const providerKey = accessReviewDrivers.find(
                      driver => driver.provider === group[0].provider,
                    );
                    if (group.length === 1) {
                      return (
                        <ConnectorListItem
                          key={group[0].id}
                          connectorKey={group[0]}
                          providerKey={providerKey}
                          organizationId={organizationId}
                          canConnect={organization.canCreateConnector}
                        />
                      );
                    }

                    return (
                      <ConnectorProviderStack
                        key={group[0].provider}
                        connectors={group}
                        providerKey={providerKey}
                        organizationId={organizationId}
                        canConnect={organization.canCreateConnector}
                      />
                    );
                  })}
                  {organization.canCreateConnector && (
                    <MarketplaceEntryCard
                      organizationId={organizationId}
                      providers={accessReviewDrivers.map(driver => driver.provider)}
                    />
                  )}
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
  );
}
