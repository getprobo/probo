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
import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AccessReviewSourceProviderListItem_provider$key } from "#/__generated__/core/AccessReviewSourceProviderListItem_provider.graphql";
import type { ConnectorListItem_connector$key } from "#/__generated__/core/ConnectorListItem_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import type { TonedCardTone } from "#/components/TonedCard/variants";

import { connectorCard } from "../variants";

import { ConnectorConnectMore } from "./ConnectorConnectMore";
import { ConnectorDeleteDialog } from "./ConnectorDeleteDialog";
import { ConnectorModules, ConnectorModulesSkeleton } from "./ConnectorModules";
import { ConnectorOrganizationSelect } from "./ConnectorOrganizationSelect";

// connectionStatus probes the vendor on every read, so this list pays one
// outbound call per connected connector.
const connectorListItemFragment = graphql`
  fragment ConnectorListItem_connector on Connector {
    id
    provider
    displayName
    connectionStatus
    createdAt
    canDelete: permission(action: "core:connector:delete")
    ...ConnectorOrganizationSelect_connector
    ...ConnectorDeleteDialog_connector
  }
`;

function connectionStatusTone(
  status: "CONNECTED" | "DISCONNECTED" | "NOT_AUTHORIZED" | "RECONNECT_REQUIRED",
): TonedCardTone {
  if (status === "CONNECTED") {
    return "green";
  }
  if (status === "RECONNECT_REQUIRED" || status === "NOT_AUTHORIZED") {
    return "amber";
  }
  return "red";
}

function connectionStatusColor(
  status: "CONNECTED" | "DISCONNECTED" | "NOT_AUTHORIZED" | "RECONNECT_REQUIRED",
) {
  if (status === "CONNECTED") {
    return "green" as const;
  }
  if (status === "RECONNECT_REQUIRED" || status === "NOT_AUTHORIZED") {
    return "amber" as const;
  }
  return "red" as const;
}

interface ConnectorListItemProps {
  connectorKey: ConnectorListItem_connector$key;
  providerKey?: AccessReviewSourceProviderListItem_provider$key;
  organizationId: string;
  canConnect: boolean;
}

export function ConnectorListItem({
  connectorKey,
  providerKey,
  organizationId,
  canConnect,
}: ConnectorListItemProps) {
  const { t, i18n } = useTranslation("organizations/settings/integrations");
  const [deleteOpen, setDeleteOpen] = useState(false);
  const connector = useFragment(connectorListItemFragment, connectorKey);
  const { controls, metaRow } = connectorCard();
  const menu = (connector.canDelete || canConnect)
    ? (
        <div className={controls()}>
          {canConnect && providerKey != null && (
            <ConnectorConnectMore
              providerKey={providerKey}
              organizationId={organizationId}
            />
          )}
          {connector.canDelete && (
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
    <TonedCard
      tone={connectionStatusTone(connector.connectionStatus)}
      icon={(
        <ThirdPartyLogo
          thirdParty={connector.provider}
          className="size-8"
        />
      )}
      lead={(
        <Heading level={2} size={3} weight="medium" highContrast className="truncate">
          {connector.displayName}
        </Heading>
      )}
      control={menu}
    >
      <div className={metaRow()}>
        <Badge
          variant="soft"
          color={connectionStatusColor(connector.connectionStatus)}
          size={1}
        >
          {t(`detailsPage.status.${connector.connectionStatus}`)}
        </Badge>
        <ConnectorOrganizationSelect connectorKey={connector} />
      </div>
      <Text size={1} color="faint">
        {t("listPage.created", {
          date: dateFormat(i18n.language, connector.createdAt),
        })}
      </Text>
      <Suspense fallback={<ConnectorModulesSkeleton />}>
        <ConnectorModules
          connectorId={connector.id}
          organizationId={organizationId}
        />
      </Suspense>
      <ConnectorDeleteDialog
        connectorKey={connector}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
      />
    </TonedCard>
  );
}
