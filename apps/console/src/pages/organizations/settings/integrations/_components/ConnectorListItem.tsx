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
import { CardButton } from "@probo/ui/src/v2/Card/CardButton";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorListItem_connector$data, ConnectorListItem_connector$key } from "#/__generated__/core/ConnectorListItem_connector.graphql";
import type { ConnectorListItemStatus_connector$key } from "#/__generated__/core/ConnectorListItemStatus_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import {
  connectionIssueKeys,
  connectionSignalFrom,
  connectionTone,
  type ConnectorConnectionStatus,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";

import { buildConnectorInitiateURL } from "../_lib/connectorSettings";
import { connectorCard } from "../variants";

import { ConnectorConnectionBadge } from "./ConnectorConnectionBadge";
import { ConnectorDeleteDialog } from "./ConnectorDeleteDialog";
import { ConnectorNameHeading } from "./ConnectorNameHeading";
import { ConnectorOrganizationSelect } from "./ConnectorOrganizationSelect";
import { ConnectorProbeError } from "./ConnectorProbeError";
import { ConnectorTypeMark } from "./ConnectorTypeMark";

const connectorListItemFragment = graphql`
  fragment ConnectorListItem_connector on Connector
    @argumentDefinitions(
      deferConnectionStatus: { type: "Boolean!", defaultValue: true }
    ) {
    id
    name
    provider
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
    ...ConnectorNameHeading_connector
    ...ConnectorTypeMark_connector
    ...ConnectorProbeError_connector
    ...ConnectorOrganizationSelect_connector
    ...ConnectorDeleteDialog_connector
    ...ConnectorListItemStatus_connector
      @defer(if: $deferConnectionStatus, label: "$defer$ConnectorListItemStatus")
  }
`;

const connectorListItemStatusFragment = graphql`
  fragment ConnectorListItemStatus_connector on Connector {
    connectionStatus
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
  const [deleteOpen, setDeleteOpen] = useState(false);
  const connector = useFragment(connectorListItemFragment, connectorKey);
  const bodyProps = {
    connector,
    organizationId,
    onSelect,
    onDeleted,
    deleteOpen,
    onDeleteOpenChange: setDeleteOpen,
  };

  return (
    <Suspense
      fallback={(
        <ConnectorListItemBody
          connectionStatus={null}
          {...bodyProps}
        />
      )}
    >
      <ConnectorListItemResolved {...bodyProps} />
    </Suspense>
  );
}

function ConnectorListItemResolved({
  connector,
  ...bodyProps
}: {
  connector: ConnectorListItem_connector$data;
} & Omit<ConnectorListItemBodyProps, "connectionStatus" | "connector">) {
  const status = useFragment(
    connectorListItemStatusFragment,
    connector as unknown as ConnectorListItemStatus_connector$key,
  );

  return (
    <ConnectorListItemBody
      connector={connector}
      connectionStatus={status.connectionStatus}
      {...bodyProps}
    />
  );
}

interface ConnectorListItemBodyProps {
  connector: ConnectorListItem_connector$data;
  connectionStatus: ConnectorConnectionStatus | null;
  organizationId: string;
  onSelect: (connectorId: string) => void;
  onDeleted?: () => void;
  deleteOpen: boolean;
  onDeleteOpenChange: (open: boolean) => void;
}

function ConnectorListItemBody({
  connector,
  connectionStatus,
  organizationId,
  onSelect,
  onDeleted,
  deleteOpen,
  onDeleteOpenChange,
}: ConnectorListItemBodyProps) {
  const { t, i18n } = useTranslation("organizations/settings/integrations");
  const { card, controls, identity, metaRow, name, tags } = connectorCard();
  const signal = connectionSignalFrom({
    connectionStatus,
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
            onClick={() => onDeleteOpenChange(true)}
          >
            <TrashIcon />
          </IconButton>
        </div>
      )
    : undefined;

  return (
    <div className={card()}>
      <TonedCard
        tone={tone}
        size={2}
        className={connector.canGet ? "pointer-events-none h-full" : "h-full"}
        stretch={connector.canGet
          ? (
              <CardButton
                type="button"
                aria-label={connector.name}
                onClick={() => onSelect(connector.id)}
              />
            )
          : undefined}
        icon={(
          <ConnectorTypeMark connectorKey={connector} />
        )}
        lead={(
          <div className={identity()}>
            <div className={name()}>
              <ConnectorNameHeading connectorKey={connector} />
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
          <div className="pointer-events-auto relative z-1 ml-auto">
            <ConnectorOrganizationSelect connectorKey={connector} />
          </div>
        </div>
        {connector.canReconnect && (
          <div className="pointer-events-auto relative z-1">
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
            connectorKey={connector}
            issues={connectionIssues}
          />
        )}
        {connector.canDelete && (
          <ConnectorDeleteDialog
            connectorKey={connector}
            open={deleteOpen}
            onOpenChange={onDeleteOpenChange}
            onDeleted={onDeleted}
          />
        )}
      </TonedCard>
    </div>
  );
}
