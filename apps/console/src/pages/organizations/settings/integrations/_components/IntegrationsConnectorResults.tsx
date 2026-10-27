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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import type { IntegrationsConnectors_organization$data } from "#/__generated__/core/IntegrationsConnectors_organization.graphql";
import type { IntegrationsConnectors_query$data } from "#/__generated__/core/IntegrationsConnectors_query.graphql";
import {
  connectionSignalFrom,
  type ConnectorConnectionStatus,
  groupByProvider,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";

import { useConnectorStatuses } from "../_lib/useConnectorStatuses";
import { integrationsList } from "../variants";

import { ConnectorGroupListItem } from "./ConnectorGroupListItem";
import { MarketplaceEntryCard } from "./MarketplaceEntryCard";

interface IntegrationsConnectorResultsProps {
  organization: IntegrationsConnectors_organization$data;
  connectorProviders: IntegrationsConnectors_query$data["connectorProviders"];
  status: ConnectorConnectionStatus | null;
  isFiltering: boolean;
}

// Without a status filter the cards render at once and each card waits for
// its own deferred status. With one, the list itself needs every status, so
// it suspends until the probes finish.
export function IntegrationsConnectorResults({
  organization,
  connectorProviders,
  status,
  isFiltering,
}: IntegrationsConnectorResultsProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const statuses = useConnectorStatuses(status == null ? null : organization.connectors);
  const connectors = statuses == null
    ? organization.connectors
    : organization.connectors.filter((connector, index) => {
        // Match what the card shows: a connected probe with an unreadable
        // organization list is presented as needing attention.
        const signal = connectionSignalFrom({
          connectionStatus: statuses[index]?.connectionStatus,
          canReconnect: connector.canReconnect,
          providerOrganizations: statuses[index]?.providerOrganizations,
        });

        return signal != null && presentConnection(signal).status === status;
      });
  const { section, sectionTitle, grid, empty } = integrationsList();

  return (
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
              <MarketplaceEntryCard organizationKey={organization} />
              {groupByProvider(connectors).map((group) => {
                const face = group[0];
                const providerKey = connectorProviders.find(
                  driver => driver.provider === face.provider,
                );

                return (
                  <ConnectorGroupListItem
                    key={group.length > 1 ? face.provider : face.id}
                    connectorKeys={group}
                    statusKeys={group}
                    organizationKey={organization}
                    providerKey={providerKey}
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
  );
}
