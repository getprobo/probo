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

import { DotsThreeVerticalIcon, PencilSimpleIcon, PlusIcon } from "@phosphor-icons/react";
import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { ResolvedAddableConnectorCard_account$key } from "#/__generated__/core/ResolvedAddableConnectorCard_account.graphql";
import type {
  ProviderOrganizationsStatus,
  ResolvedAddableConnectorCard_connector$key,
} from "#/__generated__/core/ResolvedAddableConnectorCard_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import {
  aggregateConnectionTone,
  connectionSignalFrom,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";
import { connectorDetailsPath } from "#/pages/organizations/settings/integrations/_lib/integrationPath";

import { listedConnectorAccounts } from "../_lib/listedConnectorAccounts";

import type { LoadedConnectorAccounts } from "./AddableConnectorAccounts";

const connectorFragment = graphql`
  fragment ResolvedAddableConnectorCard_connector on Connector @relay(plural: true) {
    id
    displayName
    provider
    connectionStatus
    canReconnect
    providerOrganizations {
      status
    }
    distinctAccountCount
  }
`;

const accountFragment = graphql`
  fragment ResolvedAddableConnectorCard_account on ConnectorAccount @relay(plural: true) {
    id
    name
    externalAccountId
  }
`;

export interface AddableConnectorCard {
  provider: string;
  card: ReactNode;
}

interface ResolvedAddableConnectorCardProps {
  pages: readonly LoadedConnectorAccounts[];
  connectorKeys: ResolvedAddableConnectorCard_connector$key;
  normalizedSearch: string;
  organizationId: string;
  busy: boolean;
  onAdd: (accounts: { id: string; name: string }[]) => void;
  children: (card: AddableConnectorCard | null) => ReactNode;
}

function accountNeedsOrganization(
  externalAccountId: string,
  connectorId: string,
  organizationsStatus: ProviderOrganizationsStatus,
) {
  return organizationsStatus !== "NOT_APPLICABLE"
    && externalAccountId === connectorId;
}

export function ResolvedAddableConnectorCard({
  pages,
  connectorKeys,
  normalizedSearch,
  organizationId,
  busy,
  onAdd,
  children,
}: ResolvedAddableConnectorCardProps) {
  const { t } = useTranslation();
  const { t: tConnector } = useTranslation("organizations/settings/integrations");
  const connectors = useFragment(connectorFragment, connectorKeys);
  const accountKeys: ResolvedAddableConnectorCard_account$key = pages.flatMap(
    page => page.accounts,
  );
  const accounts = useFragment(accountFragment, accountKeys);
  const [face] = connectors;
  const listed = pages.flatMap((page, index) => {
    const connector = connectors[index];
    const start = pages
      .slice(0, index)
      .reduce((sum, item) => sum + item.accounts.length, 0);
    const pageAccounts = accounts.slice(start, start + page.accounts.length);
    if (connector == null) {
      return [];
    }
    return (listedConnectorAccounts(pageAccounts, connector, normalizedSearch) ?? [])
      .map(account => ({
        id: account.id,
        name: account.name,
        needsOrganization: accountNeedsOrganization(
          account.externalAccountId,
          connector.id,
          connector.providerOrganizations.status,
        ),
      }));
  });
  const presented = connectors.flatMap((connector) => {
    const signal = connectionSignalFrom({
      connectionStatus: connector.connectionStatus,
      canReconnect: connector.canReconnect,
      providerOrganizations: {
        status: connector.providerOrganizations.status,
      },
    });
    return signal == null ? [] : [presentConnection(signal)];
  });
  const [firstPresented] = presented;
  const sharedStatus = firstPresented != null
    && presented.every(item => item.status === firstPresented.status)
    ? firstPresented.status
    : null;
  const tone = presented.length === 0 ? "green" : aggregateConnectionTone(presented);
  const accountCount = face?.distinctAccountCount ?? 0;
  const addable = listed.filter(account => !account.needsOrganization);
  const hasNext = pages.some(page => page.hasNext);
  if (face == null || (listed.length === 0 && !hasNext)) {
    return children(null);
  }

  return children({
    provider: face.provider,
    card: (
      <TonedCard
        tone={tone}
        size={2}
        icon={(
          <ThirdPartyLogo thirdParty={face.provider} />
        )}
        lead={(
          <Heading level={3} size={3} weight="medium" highContrast className="min-w-0 truncate">
            {face.displayName}
          </Heading>
        )}
        control={(
          <Dropdown>
            <DropdownTrigger
              render={(
                <IconButton
                  variant="ghost"
                  color="neutral"
                  size={1}
                  aria-label={t("accessReviewSourcesPage.actions.more")}
                >
                  <DotsThreeVerticalIcon />
                </IconButton>
              )}
            />
            <DropdownPopup align="end">
              <DropdownItem
                iconStart={<PlusIcon />}
                disabled={busy || addable.length === 0}
                onClick={() => {
                  void onAdd(addable);
                }}
              >
                {t("accessReviewSourcesPage.actions.addAll")}
              </DropdownItem>
              <DropdownItem
                iconStart={<PencilSimpleIcon />}
                render={(
                  <Link to={connectorDetailsPath(organizationId, face.provider)} />
                )}
              >
                {t("accessReviewSourcesPage.actions.edit")}
              </DropdownItem>
            </DropdownPopup>
          </Dropdown>
        )}
      >
        {listed.length > 0 && addable.length === 0 && (
          <Text size={2} color="faint">
            {t("accessReviewSourcesPage.needsOrganization")}
          </Text>
        )}
        <div className="mt-auto flex flex-wrap items-center gap-2">
          <Badge variant="soft" color={tone} size={1}>
            {sharedStatus != null
              ? tConnector(`detailsPage.status.${sharedStatus}`)
              : tConnector("listPage.connectedCount", {
                  connected: presented.filter(item => item.status === "CONNECTED").length,
                  total: presented.length,
                })}
          </Badge>
          <Badge variant="soft" color="neutral" size={1}>
            {t("accessReviewSourcesPage.accountCount", {
              count: accountCount,
            })}
          </Badge>
        </div>
      </TonedCard>
    ),
  });
}
