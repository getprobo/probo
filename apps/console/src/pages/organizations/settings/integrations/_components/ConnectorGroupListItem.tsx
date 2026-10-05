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
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorGroupListItem_connector$key } from "#/__generated__/core/ConnectorGroupListItem_connector.graphql";
import type { ConnectorGroupListItem_provider$key } from "#/__generated__/core/ConnectorGroupListItem_provider.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import {
  aggregateConnectionTone,
  type ConnectionIssueKey,
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

// connectionStatus probes the vendor on every read, so this list pays one
// outbound call per connected connector. providerOrganizations does too, for
// providers that have an account picker.
const connectorGroupListItemProviderFragment = graphql`
  fragment ConnectorGroupListItem_provider on ConnectorProviderInfo {
    ...ConnectorConnectMore_provider
  }
`;

const connectorGroupListItemFragment = graphql`
  fragment ConnectorGroupListItem_connector on Connector @relay(plural: true) {
    id
    name
    provider
    displayName
    connectionStatus
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
    canGet: permission(action: "core:connector:get")
    canDelete: permission(action: "core:connector:delete")
    ...ConnectorDeleteDialog_connector
  }
`;

interface ConnectorGroupListItemProps {
  connectorKeys: ConnectorGroupListItem_connector$key;
  providerKey?: ConnectorGroupListItem_provider$key | null;
  organizationId: string;
  canConnect: boolean;
}

export function ConnectorGroupListItem({
  connectorKeys,
  providerKey,
  organizationId,
  canConnect,
}: ConnectorGroupListItemProps) {
  const { t, i18n } = useTranslation("organizations/settings/integrations");
  const [deleteOpen, setDeleteOpen] = useState(false);
  const connectors = useFragment(connectorGroupListItemFragment, connectorKeys);
  const provider = useFragment(connectorGroupListItemProviderFragment, providerKey ?? null);
  const { card, controls, identity, name, tags, title } = connectorCard();
  const [face] = connectors;
  if (face == null) {
    return null;
  }

  const single = connectors.length === 1 ? face : null;
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
          {canConnect && provider != null && (
            <ConnectorConnectMore
              providerKey={provider}
              organizationId={organizationId}
            />
          )}
          {showDelete && (
            <IconButton
              variant="ghost"
              color="red"
              size={1}
              aria-label={t("detailsPage.actions.delete")}
              onClick={() => setDeleteOpen(true)}
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
        size={2}
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
          <ThirdPartyLogo thirdParty={face.provider} />
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
            onOpenChange={setDeleteOpen}
          />
        )}
      </TonedCard>
      <ConnectionIssueMark issues={connectionIssues} />
    </div>
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
