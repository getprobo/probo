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

import { CaretLeftIcon, MagnifyingGlassIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MarketplacePageQuery } from "#/__generated__/core/MarketplacePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { integrationListPath } from "../_lib/integrationPath";
import { integrationsList, integrationsPage } from "../variants";

import { ConnectorProviderListItem } from "./_components/ConnectorProviderListItem";
import { MarketplaceThirdPartyMatches } from "./_components/MarketplaceThirdPartyMatches";
import { MarketplaceThirdPartyNotice } from "./_components/MarketplaceThirdPartyNotice";

export const marketplacePageQuery = graphql`
  query MarketplacePageQuery($organizationId: ID!) {
    ...MarketplaceThirdPartyMatches_query
    connectorProviders {
      provider
      displayName
      ...ConnectorProviderListItem_provider
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateConnector: permission(action: "core:connector:create")
        canListThirdParties: permission(action: "core:thirdParty:list")
        connectors {
          id
          provider
        }
      }
    }
  }
`;

function credentialCountByProvider(
  connectors: readonly { provider: string }[],
): Map<string, number> {
  const counts = new Map<string, number>();

  for (const connector of connectors) {
    counts.set(connector.provider, (counts.get(connector.provider) ?? 0) + 1);
  }

  return counts;
}

interface MarketplacePageProps {
  queryRef: PreloadedQuery<MarketplacePageQuery>;
}

export function MarketplacePage({ queryRef }: MarketplacePageProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const [searchQuery, setSearchQuery] = useState("");
  const [thirdPartyOnly, setThirdPartyOnly] = useState(false);
  const [matchedProviders, setMatchedProviders] = useState<ReadonlySet<string>>(() => new Set());
  const handleMatchedProviders = useCallback((next: ReadonlySet<string>) => {
    setMatchedProviders((current) => {
      if (current.size === next.size && [...next].every(provider => current.has(provider))) {
        return current;
      }

      return next;
    });
  }, []);
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<MarketplacePageQuery>(marketplacePageQuery, queryRef);
  const { organization, connectorProviders } = data;

  usePageTitle(t("marketplacePage.title"));

  if (organization.__typename !== "Organization") {
    throw new NotFoundError(t("listPage.notFound"));
  }

  const credentialCounts = credentialCountByProvider(organization.connectors);
  const canCreateConnector = organization.canCreateConnector;
  const canListThirdParties = organization.canListThirdParties;
  const normalizedSearch = searchQuery.trim().toLowerCase();
  const isSearching = normalizedSearch !== "";
  const isConnectableThirdParty = (provider: string) =>
    matchedProviders.has(provider)
    && (credentialCounts.get(provider) ?? 0) === 0;
  const providers = canCreateConnector
    ? connectorProviders
        .filter((provider) => {
          if (thirdPartyOnly && !isConnectableThirdParty(provider.provider)) {
            return false;
          }

          if (!isSearching) {
            return true;
          }

          return provider.displayName.toLowerCase().includes(normalizedSearch)
            || provider.provider.replaceAll("_", " ").toLowerCase().includes(normalizedSearch);
        })
        .sort((a, b) => a.displayName.localeCompare(b.displayName))
    : [];
  const connectableThirdParties = connectorProviders
    .filter(provider => isConnectableThirdParty(provider.provider))
    .sort((a, b) => a.displayName.localeCompare(b.displayName));
  const { root, header, intro } = integrationsPage();
  const { root: list, tools, search, filters, filter, marketplaceGrid, empty } = integrationsList();

  return (
    <div className={root()}>
      {canCreateConnector && canListThirdParties && (
        <MarketplaceThirdPartyMatches
          queryKey={data}
          onMatched={handleMatchedProviders}
        />
      )}
      {connectableThirdParties.length > 0 && (
        <MarketplaceThirdPartyNotice
          providers={connectableThirdParties}
          onSuggest={() => setThirdPartyOnly(true)}
        />
      )}
      <div className={header()}>
        <div className={intro()}>
          <Link
            to={integrationListPath(organizationId)}
            size={2}
            color="neutral"
            underline={false}
            iconStart={<CaretLeftIcon />}
            className="self-start"
          >
            {t("listPage.title")}
          </Link>
          <div className="flex items-baseline gap-2">
            <Heading level={1} size={6} weight="medium" highContrast>
              {t("marketplacePage.title")}
            </Heading>
            {canCreateConnector && (
              <Text size={2} color="faint">
                {connectorProviders.length}
              </Text>
            )}
          </div>
          <Text size={2} color="faint">
            {t("marketplacePage.description")}
          </Text>
        </div>
      </div>
      <div className={list()}>
        {canCreateConnector && (
          <div className={tools()}>
            <div className={search()}>
              <TextField
                icon={<MagnifyingGlassIcon />}
                value={searchQuery}
                onValueChange={setSearchQuery}
                placeholder={t("listPage.searchPlaceholder")}
                aria-label={t("listPage.searchPlaceholder")}
                autoFocus
              />
            </div>
            {connectableThirdParties.length > 0 && (
              <div className={filters()}>
                <div className={filter()}>
                  <Select
                    value={thirdPartyOnly ? "connectable" : null}
                    onValueChange={(value: string | null) => {
                      setThirdPartyOnly(value === "connectable");
                    }}
                  >
                    <SelectTrigger
                      size={2}
                      placeholder={t("marketplacePage.thirdParties.all")}
                      aria-label={t("marketplacePage.thirdParties.filter")}
                    >
                      {(value: string | null) => (
                        value === "connectable"
                          ? t("marketplacePage.thirdParties.filter")
                          : t("marketplacePage.thirdParties.all")
                      )}
                    </SelectTrigger>
                    <SelectPopup align="end">
                      <SelectItem value={null}>
                        {t("marketplacePage.thirdParties.all")}
                      </SelectItem>
                      <SelectItem value="connectable">
                        {t("marketplacePage.thirdParties.filter")}
                      </SelectItem>
                    </SelectPopup>
                  </Select>
                </div>
              </div>
            )}
          </div>
        )}
        {!canCreateConnector
          ? (
              <Card variant="soft" size={2}>
                <div className={empty()}>
                  <Text size={2} color="faint">
                    {t("marketplacePage.permissionDenied")}
                  </Text>
                </div>
              </Card>
            )
          : providers.length === 0
            ? (
                <Card variant="soft" size={2}>
                  <div className={empty()}>
                    <Text size={2} color="faint">
                      {isSearching
                        ? t("listPage.emptyAvailableSearch")
                        : t("listPage.emptyAvailable")}
                    </Text>
                  </div>
                </Card>
              )
            : (
                <div className={marketplaceGrid()}>
                  {providers.map(provider => (
                    <ConnectorProviderListItem
                      key={provider.provider}
                      providerKey={provider}
                      credentialCount={credentialCounts.get(provider.provider) ?? 0}
                      matchesThirdParty={isConnectableThirdParty(provider.provider)}
                    />
                  ))}
                </div>
              )}
      </div>
    </div>
  );
}
