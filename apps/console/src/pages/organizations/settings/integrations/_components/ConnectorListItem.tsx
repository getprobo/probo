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
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorListItem_connector$key } from "#/__generated__/core/ConnectorListItem_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import {
  connectionIssueKeys,
  connectionSignalFrom,
  connectionTone,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";

import { buildConnectorInitiateURL } from "../_lib/connectorSettings";
import { connectorCard } from "../variants";

import { ConnectorConnectionBadge } from "./ConnectorConnectionBadge";
import { ConnectorDeleteDialog } from "./ConnectorDeleteDialog";
import { ConnectorMethodIcon } from "./ConnectorMethodIcon";
import { ConnectorNameHeading } from "./ConnectorNameHeading";
import { ConnectorOrganizationSelect } from "./ConnectorOrganizationSelect";
import { ConnectorProbeError } from "./ConnectorProbeError";
import { ConnectorTypeMark } from "./ConnectorTypeMark";

const connectorListItemFragment = graphql`
  fragment ConnectorListItem_connector on Connector {
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
    canUpdate: permission(action: "core:connector:update")
    canDelete: permission(action: "core:connector:delete")
    ...ConnectorOrganizationSelect_connector
    ...ConnectorDeleteDialog_connector
  }
`;

interface ConnectorListItemProps {
  connectorKey: ConnectorListItem_connector$key;
  organizationId: string;
  onSelect: (connectorId: string) => void;
  onDeleted?: () => void;
}

export function ConnectorListItem({
  connectorKey,
  organizationId,
  onSelect,
  onDeleted,
}: ConnectorListItemProps) {
  const { t, i18n } = useTranslation("organizations/settings/integrations");
  const [deleteOpen, setDeleteOpen] = useState(false);
  const connector = useFragment(connectorListItemFragment, connectorKey);
  const { card, controls, identity, metaRow, name, tags } = connectorCard();
  const signal = connectionSignalFrom({
    connectionStatus: connector.connectionStatus,
    canReconnect: connector.canReconnect,
    providerOrganizations: {
      status: connector.providerOrganizations.status,
    },
  });
  const solo = signal == null ? null : presentConnection(signal);
  const connectionIssues = connectionIssueKeys(solo == null ? [] : [solo]);
  const tone = solo == null ? "sand" : connectionTone(solo.status);
  const menu = connector.canDelete
    ? (
        <div className={controls({ className: "pointer-events-auto" })}>
          <IconButton
            variant="ghost"
            color="red"
            size={1}
            aria-label={t("detailsPage.actions.delete")}
            onClick={() => setDeleteOpen(true)}
          >
            <TrashIcon />
          </IconButton>
        </div>
      )
    : undefined;

  return (
    <div className={card()}>
      {connector.canGet && (
        <button
          type="button"
          aria-label={connector.name}
          className="absolute inset-0 z-0 cursor-pointer"
          onClick={() => onSelect(connector.id)}
        />
      )}
      <TonedCard
        tone={tone}
        size={2}
        className={connector.canGet ? "pointer-events-none relative z-0 h-full" : "relative z-0 h-full"}
        icon={(
          <ConnectorTypeMark
            protocol={connector.protocol}
            canReconnect={connector.canReconnect}
          />
        )}
        lead={(
          <div className={identity()}>
            <div className={name()}>
              <ConnectorNameHeading
                connectorId={connector.id}
                name={connector.name}
                canUpdate={connector.canUpdate}
              />
            </div>
          </div>
        )}
        control={menu}
      >
        <div className={tags()}>
          <ConnectorConnectionBadge
            aggregated={false}
            aggregatedTotal={null}
            connectedCount={solo?.status === "CONNECTED" ? 1 : 0}
            tone={tone}
            solo={solo}
          />
          <Badge variant="soft" color="neutral" size={1}>
            {t("listPage.accountCount", { count: connector.accounts.totalCount })}
          </Badge>
        </div>
        <div className={metaRow()}>
          <div className="pointer-events-auto ml-auto">
            <ConnectorOrganizationSelect connectorKey={connector} />
          </div>
        </div>
        {connector.canReconnect && (
          <div className="pointer-events-auto">
            <ButtonAnchor
              href={buildConnectorInitiateURL(
                organizationId,
                connector.provider,
                connector.protocol,
                {
                  connectorId: connector.id,
                  oauth2Scopes: connector.oauth2Scopes,
                },
              )}
              variant="solid"
              size={1}
            >
              {t("detailsPage.actions.reconnect")}
            </ButtonAnchor>
          </div>
        )}
        <Text size={1} color="faint">
          {t("listPage.created", {
            date: dateFormat(i18n.language, connector.createdAt),
          })}
        </Text>
        {connectionIssues.length > 0 && (
          <ConnectorProbeError
            issues={connectionIssues}
            provider={connector.displayName}
          />
        )}
        {connector.canDelete && (
          <ConnectorDeleteDialog
            connectorKey={connector}
            open={deleteOpen}
            onOpenChange={setDeleteOpen}
            onDeleted={onDeleted}
          />
        )}
      </TonedCard>
    </div>
  );
}
