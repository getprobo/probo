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

import { XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { DrawerBody } from "@probo/ui/src/v2/Drawer/DrawerBody";
import { DrawerClose } from "@probo/ui/src/v2/Drawer/DrawerClose";
import { DrawerDescription } from "@probo/ui/src/v2/Drawer/DrawerDescription";
import { DrawerFooter } from "@probo/ui/src/v2/Drawer/DrawerFooter";
import { DrawerHeader } from "@probo/ui/src/v2/Drawer/DrawerHeader";
import { DrawerTitle } from "@probo/ui/src/v2/Drawer/DrawerTitle";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { ConnectorAccountList_accounts$key } from "#/__generated__/core/ConnectorAccountList_accounts.graphql";
import type { ConnectorAccountList_connector$key } from "#/__generated__/core/ConnectorAccountList_connector.graphql";
import type { ConnectorAccountListAccountsQuery } from "#/__generated__/core/ConnectorAccountListAccountsQuery.graphql";
import type { ConnectorAccountListEnableMutation } from "#/__generated__/core/ConnectorAccountListEnableMutation.graphql";
import type { ConnectorAccountListQuery } from "#/__generated__/core/ConnectorAccountListQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";
import {
  connectionIssueKeys,
  connectionSignalFrom,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";

import { connectorAccountsDrawer } from "../variants";

import { ConnectorAccountListItem } from "./ConnectorAccountListItem";
import { ConnectorProbeError } from "./ConnectorProbeError";
import { DiscoveredConnectorAccountListItem } from "./DiscoveredConnectorAccountListItem";

const PAGE_SIZE = 50;

const connectorAccountListFragment = graphql`
  fragment ConnectorAccountList_connector on Connector {
    id
    canEnable: permission(action: "core:connector:create")
    canReconnect
    connectionStatus
    providerOrganizations {
      status
    }
    accountDiscovery {
      status
      nodes {
        externalAccountId
        name
        enabled
        ...DiscoveredConnectorAccountListItem_account
      }
    }
    ...ConnectorAccountList_accounts
    ...DiscoveredConnectorAccountListItem_connector
    ...ConnectorProbeError_connector
    ...ConnectorAccountListItem_connector
  }
`;

export const connectorAccountListQuery = graphql`
  query ConnectorAccountListQuery($connectorId: ID!) {
    connector: node(id: $connectorId) {
      __typename
      ... on Connector {
        ...ConnectorAccountList_connector
      }
    }
  }
`;

const connectorAccountListAccountsFragment = graphql`
  fragment ConnectorAccountList_accounts on Connector
  @refetchable(queryName: "ConnectorAccountListAccountsQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    accounts(
      first: $first
      after: $after
      orderBy: { direction: ASC, field: CREATED_AT }
    ) @connection(key: "ConnectorAccountList_accounts", filters: []) {
      totalCount
      edges {
        node {
          id
          ...ConnectorAccountListItem_account
        }
      }
    }
  }
`;

const enableConnectorAccountsMutation = graphql`
  mutation ConnectorAccountListEnableMutation($input: EnableConnectorAccountsInput!) {
    enableConnectorAccounts(input: $input) {
      connectorAccounts {
        externalAccountId
      }
    }
  }
`;

interface ConnectorAccountListProps {
  queryRef: PreloadedQuery<ConnectorAccountListQuery>;
  onReload: (connectorId: string) => void;
}

export function ConnectorAccountList({
  queryRef,
  onReload,
}: ConnectorAccountListProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const data = usePreloadedQuery<ConnectorAccountListQuery>(
    connectorAccountListQuery,
    queryRef,
  );

  if (data.connector?.__typename !== "Connector") {
    throw new NotFoundError(t("detailsPage.notFound"));
  }

  return (
    <ConnectorAccountListContent
      connectorKey={data.connector}
      onReload={onReload}
    />
  );
}

interface ConnectorAccountListContentProps {
  connectorKey: ConnectorAccountList_connector$key;
  onReload: (connectorId: string) => void;
}

function ConnectorAccountListContent({
  connectorKey,
  onReload,
}: ConnectorAccountListContentProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const connector = useFragment(connectorAccountListFragment, connectorKey);
  const {
    data: accountsData,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    ConnectorAccountListAccountsQuery,
    ConnectorAccountList_accounts$key
  >(connectorAccountListAccountsFragment, connector);
  const [selected, setSelected] = useState<ReadonlySet<string>>(() => new Set());
  const [enableAccounts, isEnabling] = useMutation<ConnectorAccountListEnableMutation>(
    enableConnectorAccountsMutation,
  );
  const [, startTransition] = useTransition();
  const signal = connectionSignalFrom({
    connectionStatus: connector.connectionStatus,
    canReconnect: connector.canReconnect,
    providerOrganizations: {
      status: connector.providerOrganizations.status,
    },
  });
  const presented = signal == null ? null : presentConnection(signal);
  const issues = connectionIssueKeys(presented == null ? [] : [presented]);
  const stored = accountsData.accounts.edges;
  const discoveryFailed = issues.length === 0
    && connector.accountDiscovery.status === "UNAVAILABLE";
  const pending = issues.length === 0 && !discoveryFailed
    ? connector.accountDiscovery.nodes.filter(account => !account.enabled)
    : [];
  const selectedAccounts = pending.filter(account => selected.has(account.externalAccountId));
  const allPendingSelected = pending.length > 0 && selectedAccounts.length === pending.length;
  const total = accountsData.accounts.totalCount + pending.length;
  const { heading, title, list, empty, actions } = connectorAccountsDrawer();

  function reload() {
    setSelected(new Set());
    startTransition(() => {
      onReload(connector.id);
    });
  }

  function toggleSelectAll() {
    setSelected(
      allPendingSelected
        ? new Set()
        : new Set(pending.map(account => account.externalAccountId)),
    );
  }

  async function enableSelected() {
    if (selectedAccounts.length === 0 || isEnabling) {
      return;
    }

    try {
      await enableAccounts({
        variables: {
          input: {
            connectorId: connector.id,
            accounts: selectedAccounts.map(account => ({
              externalAccountId: account.externalAccountId,
              name: account.name,
            })),
          },
        },
      }, { errorToast: t("detailsPage.errors.enable") });
    } catch {
      return;
    }

    reload();
  }

  return (
    <>
      <DrawerHeader>
        <div className={heading()}>
          <div className={title()}>
            <DrawerTitle>{t("detailsPage.accounts.title")}</DrawerTitle>
            <Text size={2} color="faint">{total}</Text>
          </div>
          <DrawerDescription>
            {t("detailsPage.accounts.subtitle")}
          </DrawerDescription>
        </div>
        <DrawerClose
          render={(
            <IconButton
              variant="ghost"
              color="neutral"
              size={2}
              aria-label={t("detailsPage.actions.close")}
            >
              <XIcon />
            </IconButton>
          )}
        />
      </DrawerHeader>
      <DrawerBody className={issues.length === 0 && connector.canEnable && pending.length > 0 ? "pb-16" : undefined}>
        {issues.length === 0 && !discoveryFailed && pending.length === 0 && stored.length === 0
          ? (
              <Card variant="soft" size={2}>
                <div className={empty()}>
                  <Text size={2} color="faint">
                    {t("detailsPage.accounts.empty")}
                  </Text>
                </div>
              </Card>
            )
          : (
              <div className={list()}>
                {(issues.length > 0 || discoveryFailed) && (
                  <ConnectorProbeError
                    connectorKey={connector}
                    issues={issues}
                    messages={discoveryFailed
                      ? [t("detailsPage.accounts.discoveryFailed")]
                      : undefined}
                  />
                )}
                {stored.map(({ node }) => (
                  <ConnectorAccountListItem
                    key={node.id}
                    accountKey={node}
                    connectorKey={connector}
                    unknown={issues.length > 0}
                    onDisconnected={reload}
                  />
                ))}
                {pending.map(account => (
                  <DiscoveredConnectorAccountListItem
                    key={account.externalAccountId}
                    accountKey={account}
                    connectorKey={connector}
                    selected={selected.has(account.externalAccountId)}
                    selectable={connector.canEnable}
                    onSelectedChange={(checked) => {
                      setSelected((current) => {
                        const next = new Set(current);
                        if (checked) {
                          next.add(account.externalAccountId);
                        } else {
                          next.delete(account.externalAccountId);
                        }
                        return next;
                      });
                    }}
                  />
                ))}
              </div>
            )}
        {hasNext && (
          <Button
            variant="ghost"
            color="neutral"
            loading={isLoadingNext}
            onClick={() => loadNext(PAGE_SIZE)}
            className="self-start"
          >
            {t("detailsPage.accounts.loadMore")}
          </Button>
        )}
      </DrawerBody>
      {issues.length === 0 && connector.canEnable && pending.length > 0 && (
        <DrawerFooter className="absolute inset-x-4 bottom-4">
          <div className={actions()}>
            <Button
              variant="ghost"
              color="neutral"
              disabled={isEnabling}
              onClick={toggleSelectAll}
            >
              {t("detailsPage.accounts.selectAll")}
            </Button>
            <Button
              variant="solid"
              loading={isEnabling}
              disabled={selectedAccounts.length === 0}
              onClick={() => {
                void enableSelected();
              }}
            >
              {t("detailsPage.accounts.enable")}
            </Button>
          </div>
        </DrawerFooter>
      )}
    </>
  );
}
