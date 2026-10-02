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

import { PencilSimpleIcon, PlusIcon } from "@phosphor-icons/react";
import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { iconButton } from "@probo/ui/src/v2/IconButton/variants";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { AddableConnectorListItem_account$key } from "#/__generated__/core/AddableConnectorListItem_account.graphql";
import type {
  AddableConnectorListItem_connector$data,
  AddableConnectorListItem_connector$key,
  ProviderOrganizationsStatus,
} from "#/__generated__/core/AddableConnectorListItem_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useMutation } from "#/lib/relay/useMutation";
import {
  aggregateConnectionTone,
  connectionSignalFrom,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";
import { connectorDetailsPath } from "#/pages/organizations/settings/integrations/_lib/integrationPath";

import {
  createAccessReviewSourcesMutation,
  prependCreatedSourceEdges,
} from "../../dialogs/accessReviewSourceMutations";
import { listedConnectorAccounts } from "../_lib/listedConnectorAccounts";

import {
  ConnectorAccountPages,
  type LoadedConnectorAccounts,
} from "./AddableConnectorAccounts";

const fragment = graphql`
  fragment AddableConnectorListItem_connector on Connector @relay(plural: true) {
    id
    displayName
    provider
    connectionStatus
    canReconnect
    providerOrganizations {
      status
    }
    distinctAccountCount
    ...AddableConnectorAccounts_connector
  }
`;

const accountFragment = graphql`
  fragment AddableConnectorListItem_account on ConnectorAccount @relay(plural: true) {
    id
    name
    externalAccountId
  }
`;

export interface AddableConnectorCard {
  provider: string;
  card: ReactNode;
}

interface AddableConnectorListItemProps {
  connectorKeys: AddableConnectorListItem_connector$key;
  organizationId: string;
  connectionId: string;
  normalizedSearch: string;
  children: (card: AddableConnectorCard | null) => ReactNode;
}

interface AddableConnectorGroupsProps {
  groups: readonly AddableConnectorListItem_connector$key[];
  normalizedSearch: string;
  organizationId: string;
  connectionId: string;
  children: (cards: readonly AddableConnectorCard[]) => ReactNode;
}

function accountNeedsOrganization(
  externalAccountId: string,
  connectorId: string,
  organizationsStatus: ProviderOrganizationsStatus,
) {
  return organizationsStatus !== "NOT_APPLICABLE"
    && externalAccountId === connectorId;
}

export function AddableConnectorGroups({
  groups,
  normalizedSearch,
  organizationId,
  connectionId,
  children,
}: AddableConnectorGroupsProps) {
  return (
    <ResolveAddableConnectorGroups
      groups={groups}
      index={0}
      cards={[]}
      normalizedSearch={normalizedSearch}
      organizationId={organizationId}
      connectionId={connectionId}
    >
      {children}
    </ResolveAddableConnectorGroups>
  );
}

function ResolveAddableConnectorGroups({
  groups,
  index,
  cards,
  normalizedSearch,
  organizationId,
  connectionId,
  children,
}: AddableConnectorGroupsProps & {
  index: number;
  cards: readonly AddableConnectorCard[];
}) {
  if (index >= groups.length) {
    return children(cards);
  }

  const group = groups[index];
  return (
    <AddableConnectorListItem
      connectorKeys={group}
      organizationId={organizationId}
      connectionId={connectionId}
      normalizedSearch={normalizedSearch}
    >
      {card => (
        <ResolveAddableConnectorGroups
          groups={groups}
          index={index + 1}
          cards={card == null ? cards : [...cards, card]}
          normalizedSearch={normalizedSearch}
          organizationId={organizationId}
          connectionId={connectionId}
        >
          {children}
        </ResolveAddableConnectorGroups>
      )}
    </AddableConnectorListItem>
  );
}

export function AddableConnectorListItem({
  connectorKeys,
  organizationId,
  connectionId,
  normalizedSearch,
  children,
}: AddableConnectorListItemProps) {
  const { t } = useTranslation();
  const connectors = useFragment(fragment, connectorKeys);
  const [isAdding, setIsAdding] = useState(false);
  const [createAccessReviewSources, isCreating]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourcesMutation,
    );
  const busy = isAdding || isCreating;

  async function addSources(accounts: { id: string; name: string }[]) {
    if (busy || accounts.length === 0) {
      return;
    }

    setIsAdding(true);
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
          prependCreatedSourceEdges(store, connectionId);
        },
      }, {
        successMessage: t("accessReviewConnectionsPage.messages.created"),
        errorToast: t("accessReviewConnectionsPage.errors.create"),
      });
    } catch {
      // The mutation hook already reported the error.
    } finally {
      setIsAdding(false);
    }
  }

  return (
    <ConnectorAccountPages connectorKeys={connectors}>
      {pages => (
        <ResolvedAddableConnectorCard
          pages={pages}
          connectors={connectors}
          normalizedSearch={normalizedSearch}
          organizationId={organizationId}
          busy={busy}
          onAdd={(accounts) => {
            void addSources(accounts);
          }}
        >
          {children}
        </ResolvedAddableConnectorCard>
      )}
    </ConnectorAccountPages>
  );
}

function ResolvedAddableConnectorCard({
  pages,
  connectors,
  normalizedSearch,
  organizationId,
  busy,
  onAdd,
  children,
}: {
  pages: readonly LoadedConnectorAccounts[];
  connectors: AddableConnectorListItem_connector$data;
  normalizedSearch: string;
  organizationId: string;
  busy: boolean;
  onAdd: (accounts: { id: string; name: string }[]) => void;
  children: (card: AddableConnectorCard | null) => ReactNode;
}) {
  const { t } = useTranslation();
  const { t: tConnector } = useTranslation("organizations/settings/integrations");
  const accountKeys: AddableConnectorListItem_account$key = pages.flatMap(
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
          <div className="flex items-center gap-1">
            <IconButton
              variant="ghost"
              color="neutral"
              size={1}
              loading={busy}
              disabled={addable.length === 0}
              aria-label={addable.length > 1
                ? t("accessReviewConnectionsPage.actions.addAll")
                : t("accessReviewConnectionsPage.actions.add")}
              onClick={() => {
                void onAdd(addable);
              }}
            >
              <PlusIcon />
            </IconButton>
            <Link
              to={connectorDetailsPath(organizationId, face.provider)}
              aria-label={t("accessReviewConnectionsPage.actions.edit", {
                connector: face.displayName,
              })}
              className={iconButton({ variant: "ghost", color: "neutral", size: 1 })}
            >
              <PencilSimpleIcon />
            </Link>
          </div>
        )}
      >
        {listed.length > 0 && addable.length === 0 && (
          <Text size={2} color="faint">
            {t("accessReviewConnectionsPage.needsOrganization")}
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
            {t("accessReviewConnectionsPage.accountCount", {
              count: accountCount,
            })}
          </Badge>
        </div>
      </TonedCard>
    ),
  });
}
