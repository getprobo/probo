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

import { usePageTitle } from "@probo/hooks";
import { PageHeader } from "@probo/ui";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import type { ReactNode } from "react";
import { Trans, useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";

import type { IntegrationsPageQuery } from "#/__generated__/core/IntegrationsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { ConnectorListItem } from "./_components/ConnectorListItem";
import { ConnectorProviderListItem } from "./_components/ConnectorProviderListItem";
import { connectProviderPath } from "./_lib/integrationPath";
import { integrationSection, integrationsPage } from "./variants";

export const integrationsPageQuery = graphql`
  query IntegrationsPageQuery($organizationId: ID!) {
    accessReviewDrivers {
      provider
      displayName
      ...ConnectorProviderListItem_provider
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateSource: permission(action: "access-review:source:create")
        connectors {
          id
          provider
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
  const organizationId = useOrganizationId();
  const { organization, accessReviewDrivers }
    = usePreloadedQuery<IntegrationsPageQuery>(integrationsPageQuery, queryRef);

  usePageTitle(t("listPage.title"));

  if (organization.__typename !== "Organization") {
    throw new NotFoundError(t("listPage.notFound"));
  }

  const connectors = organization.connectors;
  const connected = new Set(connectors.map(({ provider }) => provider));
  const availableProviders = accessReviewDrivers
    .filter(({ provider }) => !connected.has(provider))
    .sort((a, b) => a.displayName.localeCompare(b.displayName));

  const { root, sections } = integrationsPage();

  return (
    <div className={root()}>
      <PageHeader
        title={t("listPage.title")}
        description={t("listPage.description")}
      />

      <div className={sections()}>
        <IntegrationSection
          title={t("listPage.sections.connected")}
          count={connectors.length}
          empty={t("listPage.emptyConnected")}
        >
          {connectors.map(connector => (
            <ConnectorListItem key={connector.id} connectorKey={connector} />
          ))}
        </IntegrationSection>

        {organization.canCreateSource && (
          <IntegrationSection
            title={t("listPage.sections.available")}
            count={availableProviders.length}
            empty={t("listPage.emptyAvailable")}
            notice={(
              <Callout color="sky">
                <Trans
                  ns="organizations/settings/integrations"
                  i18nKey="listPage.connectElsewhere"
                  components={{
                    connections: (
                      <Link
                        to={connectProviderPath(organizationId)}
                        className="font-medium underline hover:no-underline"
                      />
                    ),
                  }}
                />
              </Callout>
            )}
          >
            {availableProviders.map(provider => (
              <ConnectorProviderListItem
                key={provider.provider}
                providerKey={provider}
              />
            ))}
          </IntegrationSection>
        )}
      </div>
    </div>
  );
}

function IntegrationSection({
  title,
  count,
  empty,
  notice,
  children,
}: {
  title: string;
  count: number;
  empty: string;
  notice?: ReactNode;
  children: ReactNode;
}) {
  const {
    root,
    header,
    title: titleClass,
    count: countClass,
    list,
    item,
    description,
  } = integrationSection();

  return (
    <section className={root()}>
      <div className={header()}>
        <h2 className={titleClass()}>{title}</h2>
        <span className={countClass()}>{count}</span>
      </div>
      {notice}
      <ul className={list()}>
        {count > 0
          ? children
          : (
              <li className={item()}>
                <span className={description()}>{empty}</span>
              </li>
            )}
      </ul>
    </section>
  );
}
