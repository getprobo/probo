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
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MarketplacePageQuery } from "#/__generated__/core/MarketplacePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import { AccessReviewSourceProviderListItem } from "#/pages/organizations/access-reviews/connections/_components/AccessReviewSourceProviderListItem";

import { integrationListPath } from "./_lib/integrationPath";
import { integrationsList, integrationsPage } from "./variants";

export const marketplacePageQuery = graphql`
  query MarketplacePageQuery($organizationId: ID!) {
    accessReviewDrivers {
      provider
      displayName
      ...AccessReviewSourceProviderListItem_provider
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateConnector: permission(action: "core:connector:create")
      }
    }
  }
`;

interface MarketplacePageProps {
  queryRef: PreloadedQuery<MarketplacePageQuery>;
}

export function MarketplacePage({ queryRef }: MarketplacePageProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const [searchQuery, setSearchQuery] = useState("");
  const organizationId = useOrganizationId();
  const { organization, accessReviewDrivers }
    = usePreloadedQuery<MarketplacePageQuery>(marketplacePageQuery, queryRef);

  usePageTitle(t("marketplacePage.title"));

  if (organization.__typename !== "Organization") {
    throw new NotFoundError(t("listPage.notFound"));
  }

  const normalizedSearch = searchQuery.trim().toLowerCase();
  const isSearching = normalizedSearch !== "";
  const providers = organization.canCreateConnector
    ? accessReviewDrivers
        .filter((provider) => {
          if (!isSearching) {
            return true;
          }

          return provider.displayName.toLowerCase().includes(normalizedSearch)
            || provider.provider.replaceAll("_", " ").toLowerCase().includes(normalizedSearch);
        })
        .sort((a, b) => a.displayName.localeCompare(b.displayName))
    : [];

  const { root, header, intro } = integrationsPage();
  const { root: list, tools, search, marketplaceGrid, empty } = integrationsList();

  return (
    <div className={root()}>
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
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("marketplacePage.title")}
          </Heading>
          <Text size={2} color="faint">
            {t("marketplacePage.description")}
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
              placeholder={t("listPage.searchPlaceholder")}
              aria-label={t("listPage.searchPlaceholder")}
              autoFocus
            />
          </div>
        </div>
        {providers.length === 0
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
                  <AccessReviewSourceProviderListItem
                    key={provider.provider}
                    providerKey={provider}
                    organizationId={organizationId}
                  />
                ))}
              </div>
            )}
      </div>
    </div>
  );
}
