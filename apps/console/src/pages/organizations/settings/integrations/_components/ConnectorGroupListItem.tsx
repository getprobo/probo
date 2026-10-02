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

import { TrashIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { IconWarning, ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { CardLink } from "@probo/ui/src/v2/Card/CardLink";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorGroupListItem_connector$data } from "#/__generated__/core/ConnectorGroupListItem_connector.graphql";
import type { ConnectorGroupListItem_connector$key } from "#/__generated__/core/ConnectorGroupListItem_connector.graphql";
import type { ConnectorGroupListItemStatus_connector$key } from "#/__generated__/core/ConnectorGroupListItemStatus_connector.graphql";
import type { ConnectorProviderListItem_provider$key } from "#/__generated__/core/ConnectorProviderListItem_provider.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import {
  aggregateConnectionTone,
  type ConnectionIssueKey,
  type ConnectorConnectionStatus,
  connectionIssueKeys,
  connectionSignalFrom,
  connectionTone,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";

import { buildConnectorInitiateURL } from "../_lib/connectorSettings";
import { connectorDetailsPath } from "../_lib/integrationPath";
import { connectorCard } from "../variants";

import { ConnectorConnectionBadge } from "./ConnectorConnectionBadge";
import { ConnectorConnectMore } from "./ConnectorConnectMore";
import { ConnectorDeleteDialog } from "./ConnectorDeleteDialog";
import { UsedBy } from "./UsedBy";

// connectionStatus probes the vendor on every read. It is deferred so the
// card can paint before that call returns. Mutations opt out: a mutation
// response is a single payload, and a deferred field there would never arrive.
const connectorGroupListItemFragment = graphql`
  fragment ConnectorGroupListItem_connector on Connector
    @argumentDefinitions(
      deferConnectionStatus: { type: "Boolean!", defaultValue: true }
    )
    @relay(plural: true) {
    id
    name
    provider
    displayName
    canReconnect
    protocol
    oauth2Scopes
    providerOrganizations {
      status
    }
    accounts(first: 1) {
      totalCount
    }
    createdAt
    modules
    canGet: permission(action: "core:connector:get")
    canDelete: permission(action: "core:connector:delete")
    ...ConnectorDeleteDialog_connector
    ...ConnectorGroupListItemStatus_connector
      @defer(if: $deferConnectionStatus, label: "$defer$ConnectorGroupListItemStatus")
  }
`;

const connectorGroupListItemStatusFragment = graphql`
  fragment ConnectorGroupListItemStatus_connector on Connector @relay(plural: true) {
    id
    connectionStatus
  }
`;

interface ConnectorGroupListItemProps {
  connectorKeys: ConnectorGroupListItem_connector$key;
  providerKey?: ConnectorProviderListItem_provider$key;
  organizationId: string;
  canConnect: boolean;
  statusFilter: ConnectorConnectionStatus | null;
  onStatus: (id: string, status: ConnectorConnectionStatus) => void;
}

export function ConnectorGroupListItem({
  connectorKeys,
  providerKey,
  organizationId,
  canConnect,
  statusFilter,
  onStatus,
}: ConnectorGroupListItemProps) {
  const [deleteOpen, setDeleteOpen] = useState(false);
  const connectors = useFragment(connectorGroupListItemFragment, connectorKeys);
  if (connectors.length === 0) {
    return null;
  }

  const cardProps = {
    providerKey,
    organizationId,
    canConnect,
    deleteOpen,
    onDeleteOpenChange: setDeleteOpen,
  };

  return (
    <Suspense
      fallback={statusFilter == null
        ? (
            <ConnectorGroupCard
              connectors={connectors}
              statuses={null}
              {...cardProps}
            />
          )
        : null}
    >
      <ConnectorGroupResolved
        connectors={connectors}
        statusFilter={statusFilter}
        onStatus={onStatus}
        {...cardProps}
      />
    </Suspense>
  );
}

function ConnectorGroupResolved({
  connectors,
  statusFilter,
  onStatus,
  ...cardProps
}: {
  connectors: ConnectorGroupListItem_connector$data;
  statusFilter: ConnectorConnectionStatus | null;
  onStatus: (id: string, status: ConnectorConnectionStatus) => void;
} & ConnectorGroupCardProps) {
  const statuses = useFragment(
    connectorGroupListItemStatusFragment,
    connectors as unknown as ConnectorGroupListItemStatus_connector$key,
  );

  useEffect(() => {
    for (const status of statuses) {
      onStatus(status.id, status.connectionStatus);
    }
  }, [onStatus, statuses]);

  const visibleConnectors: ConnectorGroupListItem_connector$data[number][] = [];
  const visibleStatuses: ConnectorConnectionStatus[] = [];
  connectors.forEach((connector, index) => {
    const connectionStatus = statuses[index]?.connectionStatus;
    if (connectionStatus == null) {
      return;
    }
    if (statusFilter != null && connectionStatus !== statusFilter) {
      return;
    }
    visibleConnectors.push(connector);
    visibleStatuses.push(connectionStatus);
  });
  if (visibleConnectors.length === 0) {
    return null;
  }

  return (
    <ConnectorGroupCard
      connectors={visibleConnectors}
      statuses={visibleStatuses}
      {...cardProps}
    />
  );
}

interface ConnectorGroupCardProps {
  providerKey?: ConnectorProviderListItem_provider$key;
  organizationId: string;
  canConnect: boolean;
  deleteOpen: boolean;
  onDeleteOpenChange: (open: boolean) => void;
}

function ConnectorGroupCard({
  connectors,
  statuses,
  providerKey,
  organizationId,
  canConnect,
  deleteOpen,
  onDeleteOpenChange,
}: {
  connectors: readonly ConnectorGroupListItem_connector$data[number][];
  statuses: readonly ConnectorConnectionStatus[] | null;
} & ConnectorGroupCardProps) {
  const { t, i18n } = useTranslation("organizations/settings/integrations");
  const { card, controls, identity, name, tags, title } = connectorCard();
  const [face] = connectors;
  if (face == null) {
    return null;
  }

  const single = connectors.length === 1 ? face : null;
  const presented = connectors.flatMap((connector, index) => {
    const signal = connectionSignalFrom({
      connectionStatus: statuses?.[index] ?? null,
      canReconnect: connector.canReconnect,
      providerOrganizations: {
        status: connector.providerOrganizations.status,
      },
    });
    return signal == null ? [] : [presentConnection(signal)];
  });
  const connectionIssues = connectionIssueKeys(presented);
  const solo = single == null ? null : presented[0] ?? null;
  const tone = single == null
    ? (presented.length === 0 ? "sand" : aggregateConnectionTone(presented))
    : (solo == null ? "sand" : connectionTone(solo.status));
  const connectedCount = presented.filter(item => item.status === "CONNECTED").length;
  const accountTotal = connectors.reduce(
    (sum, connector) => sum + connector.accounts.totalCount,
    0,
  );
  const showDelete = single != null && single.canDelete;
  const menu = (showDelete || canConnect)
    ? (
        <div className={controls({ className: "pointer-events-auto" })}>
          {canConnect && providerKey != null && (
            <ConnectorConnectMore
              providerKey={providerKey}
              organizationId={organizationId}
            />
          )}
          {showDelete && (
            <IconButton
              variant="ghost"
              color="red"
              size={1}
              aria-label={t("detailsPage.actions.delete")}
              onClick={() => onDeleteOpenChange(true)}
            >
              <TrashIcon />
            </IconButton>
          )}
        </div>
      )
    : undefined;

  return (
    <div className={card()}>
      <TonedCard
        tone={tone}
        iconSize={14}
        className={face.canGet ? "pointer-events-none h-full" : "h-full"}
        stretch={face.canGet
          ? (
              <CardLink
                to={connectorDetailsPath(organizationId, face.provider)}
                aria-label={face.displayName}
              />
            )
          : undefined}
        icon={(
          <ThirdPartyLogo
            thirdParty={face.provider}
            className="size-12"
          />
        )}
        lead={(
          <div className={identity()}>
            <div className={name()}>
              <div className="flex min-w-0 flex-col">
                <Heading level={2} size={3} weight="medium" highContrast className={title()}>
                  {face.displayName}
                </Heading>
                {single != null && (
                  <Text size={1} color="faint" className="truncate">
                    {single.name}
                  </Text>
                )}
              </div>
            </div>
          </div>
        )}
        control={menu}
      >
        <div className={tags()}>
          <ConnectorConnectionBadge
            aggregated={single == null}
            aggregatedTotal={single == null ? connectors.length : null}
            connectedCount={connectedCount}
            tone={tone}
            solo={solo}
          />
          <Badge variant="soft" color="neutral" size={1}>
            {t("listPage.accountCount", { count: accountTotal })}
          </Badge>
        </div>
        {single != null && single.canReconnect && (
          <div className="pointer-events-auto relative z-1">
            <ButtonAnchor
              href={buildConnectorInitiateURL(
                organizationId,
                single.provider,
                single.protocol,
                {
                  connectorId: single.id,
                  oauth2Scopes: single.oauth2Scopes,
                },
              )}
              variant="solid"
              size={1}
            >
              {t("detailsPage.actions.reconnect")}
            </ButtonAnchor>
          </div>
        )}
        <div className="pointer-events-auto relative z-1">
          <UsedBy modules={listedModules(connectors)} />
        </div>
        {single != null && (
          <Text size={1} color="faint">
            {t("listPage.created", {
              date: dateFormat(i18n.language, single.createdAt),
            })}
          </Text>
        )}
        {showDelete && (
          <ConnectorDeleteDialog
            connectorKey={face}
            open={deleteOpen}
            onOpenChange={onDeleteOpenChange}
          />
        )}
      </TonedCard>
      <ConnectionIssueMark issues={connectionIssues} />
    </div>
  );
}

function listedModules(
  connectors: ReadonlyArray<{ readonly modules: ReadonlyArray<string> }>,
): string[] {
  const seen = new Set<string>();
  return connectors.flatMap(connector =>
    connector.modules.filter((module) => {
      if (seen.has(module)) {
        return false;
      }
      seen.add(module);
      return true;
    }),
  );
}

function ConnectionIssueMark({
  issues,
}: {
  issues: readonly ConnectionIssueKey[];
}) {
  if (issues.length === 0) {
    return null;
  }

  return (
    <span
      className="pointer-events-none absolute right-3 bottom-3 z-1 text-red-11"
      aria-hidden
    >
      <IconWarning size={16} className="shrink-0" />
    </span>
  );
}
