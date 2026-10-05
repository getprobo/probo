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
import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { CardSkeleton } from "@probo/ui/src/v2/Card/CardSkeleton";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Drawer } from "@probo/ui/src/v2/Drawer/Drawer";
import { DrawerBody } from "@probo/ui/src/v2/Drawer/DrawerBody";
import { DrawerClose } from "@probo/ui/src/v2/Drawer/DrawerClose";
import { DrawerDescription } from "@probo/ui/src/v2/Drawer/DrawerDescription";
import { DrawerFooter } from "@probo/ui/src/v2/Drawer/DrawerFooter";
import { DrawerHeader } from "@probo/ui/src/v2/Drawer/DrawerHeader";
import { DrawerPopup } from "@probo/ui/src/v2/Drawer/DrawerPopup";
import { DrawerTitle } from "@probo/ui/src/v2/Drawer/DrawerTitle";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";
import { Suspense, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { ConnectorAccountsDrawer_accounts$key } from "#/__generated__/core/ConnectorAccountsDrawer_accounts.graphql";
import type { ConnectorAccountsDrawer_connector$key } from "#/__generated__/core/ConnectorAccountsDrawer_connector.graphql";
import type { ConnectorAccountsDrawerAccountsQuery } from "#/__generated__/core/ConnectorAccountsDrawerAccountsQuery.graphql";
import type { ConnectorAccountsDrawerEnableMutation } from "#/__generated__/core/ConnectorAccountsDrawerEnableMutation.graphql";
import type { ConnectorAccountsDrawerPendingAccount_account$key } from "#/__generated__/core/ConnectorAccountsDrawerPendingAccount_account.graphql";
import type { ConnectorAccountsDrawerPendingAccount_connector$key } from "#/__generated__/core/ConnectorAccountsDrawerPendingAccount_connector.graphql";
import type { ConnectorAccountsDrawerQuery } from "#/__generated__/core/ConnectorAccountsDrawerQuery.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";
import {
  connectionIssueKeys,
  connectionSignalFrom,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";

import { connectorAccountsDrawer, connectorCard } from "../variants";

import { ConnectorAccountListItem } from "./ConnectorAccountListItem";
import { ConnectorProbeError } from "./ConnectorProbeError";

const PAGE_SIZE = 50;

const connectorAccountsDrawerFragment = graphql`
  fragment ConnectorAccountsDrawer_connector on Connector {
    id
    canEnable: permission(action: "core:connector:create")
    canReconnect
    connectionStatus
    providerOrganizations {
      status
    }
    discoveredAccounts {
      externalAccountId
      name
      enabled
      ...ConnectorAccountsDrawerPendingAccount_account
    }
    ...ConnectorAccountsDrawer_accounts
    ...ConnectorAccountsDrawerPendingAccount_connector
    ...ConnectorProbeError_connector
    ...ConnectorAccountListItem_connector
  }
`;

const pendingAccountFragment = graphql`
  fragment ConnectorAccountsDrawerPendingAccount_account on DiscoveredConnectorAccount {
    externalAccountId
    name
  }
`;

const pendingAccountConnectorFragment = graphql`
  fragment ConnectorAccountsDrawerPendingAccount_connector on Connector {
    provider
  }
`;

export const connectorAccountsDrawerQuery = graphql`
  query ConnectorAccountsDrawerQuery($connectorId: ID!) {
    connector: node(id: $connectorId) {
      __typename
      ... on Connector {
        ...ConnectorAccountsDrawer_connector
      }
    }
  }
`;

const connectorAccountsDrawerAccountsFragment = graphql`
  fragment ConnectorAccountsDrawer_accounts on Connector
  @refetchable(queryName: "ConnectorAccountsDrawerAccountsQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    accounts(
      first: $first
      after: $after
      orderBy: { direction: ASC, field: CREATED_AT }
    ) @connection(key: "ConnectorAccountsDrawer_accounts", filters: []) {
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
  mutation ConnectorAccountsDrawerEnableMutation($input: EnableConnectorAccountsInput!) {
    enableConnectorAccounts(input: $input) {
      connectorAccounts {
        externalAccountId
      }
    }
  }
`;

interface ConnectorAccountsDrawerProps {
  handle: ReturnType<typeof Drawer.createHandle<string>>;
  queryRef: PreloadedQuery<ConnectorAccountsDrawerQuery> | null | undefined;
  onReload: (connectorId: string) => void;
  onOpenChange: (open: boolean) => void;
}

export function ConnectorAccountsDrawer({
  handle,
  queryRef,
  onReload,
  onOpenChange,
}: ConnectorAccountsDrawerProps) {
  return (
    <Drawer<string>
      handle={handle}
      onOpenChange={onOpenChange}
      swipeDirection="right"
    >
      {({ payload: connectorId }) => (
        <DrawerPopup side="right" size={3}>
          {connectorId != null && (
            <ErrorBoundary
              fallback={(error) => {
                if (error instanceof NotFoundError) {
                  throw error;
                }
                return <DrawerLoadFailed />;
              }}
            >
              {queryRef != null && queryRef.variables.connectorId === connectorId
                ? (
                    <Suspense fallback={<ConnectorAccountsDrawerSkeleton />}>
                      <Accounts queryRef={queryRef} onReload={onReload} />
                    </Suspense>
                  )
                : <ConnectorAccountsDrawerSkeleton />}
            </ErrorBoundary>
          )}
        </DrawerPopup>
      )}
    </Drawer>
  );
}

function Accounts({
  queryRef,
  onReload,
}: {
  queryRef: PreloadedQuery<ConnectorAccountsDrawerQuery>;
  onReload: (connectorId: string) => void;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const data = usePreloadedQuery<ConnectorAccountsDrawerQuery>(
    connectorAccountsDrawerQuery,
    queryRef,
  );

  if (data.connector?.__typename !== "Connector") {
    throw new NotFoundError(t("detailsPage.notFound"));
  }

  return <AccountList connectorKey={data.connector} onReload={onReload} />;
}

function AccountList({
  connectorKey,
  onReload,
}: {
  connectorKey: ConnectorAccountsDrawer_connector$key;
  onReload: (connectorId: string) => void;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const connector = useFragment(connectorAccountsDrawerFragment, connectorKey);
  const {
    data: accountsData,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    ConnectorAccountsDrawerAccountsQuery,
    ConnectorAccountsDrawer_accounts$key
  >(connectorAccountsDrawerAccountsFragment, connector);
  const [selected, setSelected] = useState<ReadonlySet<string>>(() => new Set());
  const [enableAccounts, isEnabling] = useMutation<ConnectorAccountsDrawerEnableMutation>(
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
  const pending = issues.length === 0
    ? connector.discoveredAccounts.filter(account => !account.enabled)
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
        {issues.length === 0 && pending.length === 0 && stored.length === 0
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
                {issues.length > 0 && (
                  <ConnectorProbeError
                    connectorKey={connector}
                    issues={issues}
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
                  <PendingAccountRow
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

function PendingAccountRow({
  accountKey,
  connectorKey,
  selected,
  selectable,
  onSelectedChange,
}: {
  accountKey: ConnectorAccountsDrawerPendingAccount_account$key;
  connectorKey: ConnectorAccountsDrawerPendingAccount_connector$key;
  selected: boolean;
  selectable: boolean;
  onSelectedChange: (checked: boolean) => void;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const account = useFragment(pendingAccountFragment, accountKey);
  const connector = useFragment(pendingAccountConnectorFragment, connectorKey);
  const { identity, name, title } = connectorCard();

  return (
    <TonedCard
      tone="sand"
      size={2}
      icon={(
        <ThirdPartyLogo thirdParty={connector.provider} />
      )}
      lead={(
        <div className={identity()}>
          <div className={name()}>
            <Heading level={2} size={3} weight="medium" highContrast className={title()}>
              {account.name}
            </Heading>
            <Text size={1} color="faint" className="shrink-0 font-mono">
              #
              {account.externalAccountId}
            </Text>
          </div>
          <Badge variant="soft" color="neutral" size={1}>
            {t("detailsPage.accounts.pending")}
          </Badge>
        </div>
      )}
      control={selectable
        ? (
            <Checkbox
              checked={selected}
              aria-label={account.name}
              onCheckedChange={onSelectedChange}
            />
          )
        : undefined}
    />
  );
}

function ConnectorAccountsDrawerSkeleton() {
  const { heading, list } = connectorAccountsDrawer();

  return (
    <>
      <DrawerHeader>
        <div className={heading()}>
          <HeadingSkeleton size={4} className="w-24" />
          <TextSkeleton size={2} className="w-full" />
          <TextSkeleton size={2} className="w-4/5" />
        </div>
      </DrawerHeader>
      <DrawerBody>
        <div className={list()} aria-hidden>
          <CardSkeleton size={4} />
          <CardSkeleton size={4} />
        </div>
      </DrawerBody>
    </>
  );
}

function DrawerLoadFailed() {
  const { t } = useTranslation("organizations/settings/integrations");
  const { heading, empty } = connectorAccountsDrawer();

  return (
    <>
      <DrawerHeader>
        <div className={heading()}>
          <DrawerTitle>{t("detailsPage.accounts.title")}</DrawerTitle>
          <DrawerDescription>
            {t("detailsPage.accounts.subtitle")}
          </DrawerDescription>
        </div>
      </DrawerHeader>
      <DrawerBody>
        <Card variant="soft" size={2}>
          <div className={empty()}>
            <Text size={2} color="faint">
              {t("detailsPage.accounts.discoveryFailed")}
            </Text>
          </div>
        </Card>
      </DrawerBody>
    </>
  );
}
