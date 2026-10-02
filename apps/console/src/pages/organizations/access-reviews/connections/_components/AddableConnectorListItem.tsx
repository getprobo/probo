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
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { AddableConnectorListItem_connector$data } from "#/__generated__/core/AddableConnectorListItem_connector.graphql";
import type { AddableConnectorListItem_connector$key } from "#/__generated__/core/AddableConnectorListItem_connector.graphql";
import type { AddableConnectorListItemStatus_connector$key } from "#/__generated__/core/AddableConnectorListItemStatus_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useMutation } from "#/lib/relay/useMutation";
import {
  aggregateConnectionTone,
  type ConnectorConnectionStatus,
  connectionSignalFrom,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";
import { connectorDetailsPath } from "#/pages/organizations/settings/integrations/_lib/integrationPath";

import {
  createAccessReviewSourcesMutation,
  prependCreatedSourceEdges,
} from "../../dialogs/accessReviewSourceMutations";
import { listedConnectorAccounts } from "../_lib/listedConnectorAccounts";

const fragment = graphql`
  fragment AddableConnectorListItem_connector on Connector
    @argumentDefinitions(
      deferConnectionStatus: { type: "Boolean!", defaultValue: true }
    )
    @relay(plural: true) {
    id
    displayName
    provider
    canReconnect
    providerOrganizations {
      status
    }
    accounts(first: 50) {
      totalCount
      edges {
        node {
          id
          name
          externalAccountId
        }
      }
    }
    ...AddableConnectorListItemStatus_connector
      @defer(if: $deferConnectionStatus, label: "$defer$AddableConnectorListItemStatus")
  }
`;

const statusFragment = graphql`
  fragment AddableConnectorListItemStatus_connector on Connector @relay(plural: true) {
    connectionStatus
  }
`;

interface AddableConnectorListItemProps {
  connectorKeys: AddableConnectorListItem_connector$key;
  organizationId: string;
  connectionId: string;
  normalizedSearch: string;
}

function accountNeedsOrganization(externalAccountId: string, connectorId: string) {
  return externalAccountId === connectorId;
}

export function AddableConnectorListItem({
  connectorKeys,
  organizationId,
  connectionId,
  normalizedSearch,
}: AddableConnectorListItemProps) {
  const connectors = useFragment(fragment, connectorKeys);
  const [isAdding, setIsAdding] = useState(false);
  const [createAccessReviewSources, isCreating]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourcesMutation,
    );
  const [face] = connectors;
  const accountCount = connectors.length === 1
    ? connectors[0].accounts.totalCount
    : new Set(
      connectors.flatMap(connector =>
        connector.accounts.edges.map(({ node }) => node.externalAccountId),
      ),
    ).size;
  const listed = connectors.flatMap(connector =>
    (listedConnectorAccounts(
      connector.accounts.edges.map(({ node }) => node),
      connector,
      normalizedSearch,
    ) ?? []).map(account => ({
      id: account.id,
      name: account.name,
      needsOrganization: accountNeedsOrganization(account.externalAccountId, connector.id),
    })),
  );
  if (face == null || listed.length === 0) {
    return null;
  }

  const cardProps = {
    connectors,
    organizationId,
    connectionId,
    face,
    accountCount,
    listed,
    isAdding,
    isCreating,
    setIsAdding,
    createAccessReviewSources,
  };

  return (
    <Suspense fallback={<AddableConnectorCard statuses={null} {...cardProps} />}>
      <AddableConnectorResolved {...cardProps} />
    </Suspense>
  );
}

function AddableConnectorResolved(
  props: Omit<AddableConnectorCardProps, "statuses">,
) {
  const statuses = useFragment(
    statusFragment,
    props.connectors as unknown as AddableConnectorListItemStatus_connector$key,
  );

  return (
    <AddableConnectorCard
      statuses={statuses.map(status => status.connectionStatus)}
      {...props}
    />
  );
}

interface AddableConnectorCardProps {
  connectors: AddableConnectorListItem_connector$data;
  statuses: readonly ConnectorConnectionStatus[] | null;
  organizationId: string;
  connectionId: string;
  face: AddableConnectorListItem_connector$data[number];
  accountCount: number;
  listed: { id: string; name: string; needsOrganization: boolean }[];
  isAdding: boolean;
  isCreating: boolean;
  setIsAdding: (adding: boolean) => void;
  createAccessReviewSources: ReturnType<
    typeof useMutation<accessReviewSourceMutationsCreateMutation>
  >[0];
}

function AddableConnectorCard({
  connectors,
  statuses,
  organizationId,
  connectionId,
  face,
  accountCount,
  listed,
  isAdding,
  isCreating,
  setIsAdding,
  createAccessReviewSources,
}: AddableConnectorCardProps) {
  const { t } = useTranslation();
  const { t: tConnector } = useTranslation("organizations/settings/integrations");
  const busy = isAdding || isCreating;
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
  const [firstPresented] = presented;
  const sharedStatus = firstPresented != null
    && presented.every(item => item.status === firstPresented.status)
    ? firstPresented.status
    : null;
  const tone = presented.length === 0 ? "sand" : aggregateConnectionTone(presented);
  const addable = listed.filter(account => !account.needsOrganization);

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
    <TonedCard
      tone={tone}
      iconSize={14}
      icon={(
        <ThirdPartyLogo
          thirdParty={face.provider}
          className="size-12"
        />
      )}
      lead={(
        <Heading level={2} size={3} weight="medium" highContrast className="min-w-0 truncate">
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
              void addSources(addable);
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
      {addable.length === 0 && (
        <Text size={2} color="faint">
          {t("accessReviewConnectionsPage.needsOrganization")}
        </Text>
      )}
      <div className="mt-auto flex flex-wrap items-center gap-2">
        <Badge variant="soft" color={tone === "sand" ? "neutral" : tone} size={1}>
          {statuses == null
            ? tConnector("detailsPage.accounts.pending")
            : sharedStatus != null
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
  );
}
