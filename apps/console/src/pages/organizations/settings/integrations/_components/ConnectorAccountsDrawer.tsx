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
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";
import { Suspense, useEffect, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, graphql, useLazyLoadQuery, usePaginationFragment, useRelayEnvironment } from "react-relay";
import type { Environment } from "relay-runtime";

import type { ConnectorAccountsDrawer_accounts$key } from "#/__generated__/core/ConnectorAccountsDrawer_accounts.graphql";
import type { ConnectorAccountsDrawerAccountsQuery } from "#/__generated__/core/ConnectorAccountsDrawerAccountsQuery.graphql";
import type { ConnectorAccountsDrawerDiscoveryQuery } from "#/__generated__/core/ConnectorAccountsDrawerDiscoveryQuery.graphql";
import type { ConnectorAccountsDrawerEnableMutation } from "#/__generated__/core/ConnectorAccountsDrawerEnableMutation.graphql";
import type { ConnectorAccountsDrawerQuery } from "#/__generated__/core/ConnectorAccountsDrawerQuery.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";
import {
  connectionSignalFrom,
  type ConnectorConnectionStatus,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";

import {
  collectStoredAccountIds,
  type DiscoveredAccount,
  labelDiscoveredAccounts,
} from "../_lib/discoveredAccounts";
import { connectorAccountsDrawer, connectorCard } from "../variants";

import { ConnectorAccountListItem } from "./ConnectorAccountListItem";

const PAGE_SIZE = 50;

const connectorAccountsDrawerQuery = graphql`
  query ConnectorAccountsDrawerQuery($connectorId: ID!) {
    connector: node(id: $connectorId) {
      __typename
      ... on Connector {
        canDiscover: permission(action: "core:connector:discover")
        canEnable: permission(action: "core:connector:create")
        canDelete: permission(action: "core:connector:delete")
        initialAccountExternalId
        ...ConnectorAccountsDrawer_accounts
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
          externalAccountId
          ...ConnectorAccountListItem_account
        }
      }
    }
  }
`;

const connectorAccountsDrawerDiscoveryQuery = graphql`
  query ConnectorAccountsDrawerDiscoveryQuery($connectorId: ID!) {
    connector: node(id: $connectorId) {
      __typename
      ... on Connector {
        connectionStatus
        discoveredAccounts {
          externalAccountId
          name
        }
        accounts(first: 50, orderBy: { direction: ASC, field: CREATED_AT }) {
          pageInfo {
            hasNextPage
            endCursor
          }
          edges {
            node {
              externalAccountId
            }
          }
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
  connectorId: string | null;
  provider: string | null;
  providerName: string | null;
  open: boolean;
  fetchKey: number;
  preset: DiscoveredAccount[] | null;
  onOpenChange: (open: boolean) => void;
}

export function ConnectorAccountsDrawer({
  connectorId,
  provider,
  providerName,
  open,
  fetchKey,
  preset,
  onOpenChange,
}: ConnectorAccountsDrawerProps) {
  return (
    <Drawer
      open={open}
      onOpenChange={onOpenChange}
      swipeDirection="right"
    >
      <DrawerPopup side="right" size={3}>
        {connectorId != null && provider != null && providerName != null && (
          <SuspenseAccounts
            connectorId={connectorId}
            provider={provider}
            providerName={providerName}
            fetchKey={fetchKey}
            preset={preset}
          />
        )}
      </DrawerPopup>
    </Drawer>
  );
}

function SuspenseAccounts({
  connectorId,
  provider,
  providerName,
  fetchKey,
  preset,
}: {
  connectorId: string;
  provider: string;
  providerName: string;
  fetchKey: number;
  preset: DiscoveredAccount[] | null;
}) {
  return (
    <Suspense fallback={<ConnectorAccountsDrawerSkeleton />}>
      <ConnectorAccounts
        key={connectorId}
        connectorId={connectorId}
        provider={provider}
        providerName={providerName}
        fetchKey={fetchKey}
        preset={preset}
      />
    </Suspense>
  );
}

function ConnectorAccounts({
  connectorId,
  provider,
  providerName,
  fetchKey,
  preset,
}: {
  connectorId: string;
  provider: string;
  providerName: string;
  fetchKey: number;
  preset: DiscoveredAccount[] | null;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const environment = useRelayEnvironment();
  const data = useLazyLoadQuery<ConnectorAccountsDrawerQuery>(
    connectorAccountsDrawerQuery,
    { connectorId },
    {
      fetchKey,
      fetchPolicy: fetchKey > 0 ? "network-only" : "store-or-network",
    },
  );

  if (data.connector?.__typename !== "Connector") {
    throw new NotFoundError(t("detailsPage.notFound"));
  }

  const connector = data.connector;
  const {
    data: accountsData,
    loadNext,
    hasNext,
    isLoadingNext,
    refetch,
  } = usePaginationFragment<
    ConnectorAccountsDrawerAccountsQuery,
    ConnectorAccountsDrawer_accounts$key
  >(connectorAccountsDrawerAccountsFragment, connector);
  const [probe, setProbe] = useState<StoredProbe | null>(null);
  const [enabledIds, setEnabledIds] = useState<ReadonlySet<string>>(() => new Set());
  const [reopenedIds, setReopenedIds] = useState<ReadonlySet<string>>(() => new Set());
  const [selected, setSelected] = useState<ReadonlySet<string>>(() => new Set());
  const [enableAccounts, isEnabling] = useMutation<ConnectorAccountsDrawerEnableMutation>(
    enableConnectorAccountsMutation,
  );
  const [, startTransition] = useTransition();

  useEffect(() => {
    if (preset != null || !connector.canDiscover) {
      return;
    }

    let cancelled = false;
    const key = probeKey(connectorId, fetchKey);
    void loadDiscoveredAccounts(environment, connectorId).then(
      (next) => {
        if (cancelled) {
          return;
        }
        if (next.missing) {
          setProbe({ key, value: { kind: "missing" } });
          return;
        }
        setProbe({
          key,
          value: {
            kind: "ready",
            accounts: next.accounts,
            status: next.connectionStatus,
          },
        });
      },
      () => {
        if (!cancelled) {
          setProbe({ key, value: { kind: "failed" } });
        }
      },
    );
    return () => {
      cancelled = true;
    };
  }, [connector.canDiscover, connectorId, environment, fetchKey, preset]);

  const resolved = preset == null && connector.canDiscover
    ? (probe?.key === probeKey(connectorId, fetchKey) ? probe.value : null)
    : null;

  if (preset == null && connector.canDiscover && resolved == null) {
    return <ConnectorAccountsDrawerSkeleton />;
  }

  if (resolved?.kind === "missing") {
    throw new NotFoundError(t("detailsPage.notFound"));
  }

  const failure = resolved?.kind === "failed"
    ? t("detailsPage.accounts.discoveryFailed")
    : resolved?.kind === "ready"
      ? connectionFailure(resolved.status, t, providerName)
      : null;
  const discovered = preset != null
    ? preset
    : resolved?.kind === "ready" && failure == null
      ? resolved.accounts
      : [];
  const stored = accountsData.accounts.edges;
  const storedIds = new Set(stored.map(({ node }) => node.externalAccountId));
  const pending = discovered.filter(account => (
    (account.status === "PENDING" || reopenedIds.has(account.externalAccountId))
    && !storedIds.has(account.externalAccountId)
    && !enabledIds.has(account.externalAccountId)
  ));
  const selectedAccounts = pending.filter(account => selected.has(account.externalAccountId));
  const allPendingSelected = pending.length > 0 && selectedAccounts.length === pending.length;
  const total = accountsData.accounts.totalCount + pending.length;
  const { heading, title, list, empty, actions } = connectorAccountsDrawer();

  function markAccountPending(externalAccountId: string) {
    setEnabledIds((current) => {
      if (!current.has(externalAccountId)) {
        return current;
      }
      const next = new Set(current);
      next.delete(externalAccountId);
      return next;
    });
    setReopenedIds((current) => {
      const next = new Set(current);
      next.add(externalAccountId);
      return next;
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
            connectorId,
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

    const enabled = new Set(selectedAccounts.map(account => account.externalAccountId));
    setEnabledIds((current) => {
      const next = new Set(current);
      for (const id of enabled) {
        next.add(id);
      }
      return next;
    });
    setSelected(new Set());
    startTransition(() => {
      refetch({}, { fetchPolicy: "network-only" });
    });
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
      <DrawerBody className={failure == null && connector.canEnable && pending.length > 0 ? "pb-16" : undefined}>
        {failure != null && (
          <Card variant="soft" size={2}>
            <div className={empty()}>
              <Text size={2} color="faint">
                {failure}
              </Text>
            </div>
          </Card>
        )}
        {failure == null && pending.length === 0 && stored.length === 0
          ? (
              <Card variant="soft" size={2}>
                <div className={empty()}>
                  <Text size={2} color="faint">
                    {t("detailsPage.accounts.empty")}
                  </Text>
                </div>
              </Card>
            )
          : (pending.length > 0 || stored.length > 0) && (
              <div className={list()}>
                {stored.map(({ node }) => (
                  <ConnectorAccountListItem
                    key={node.id}
                    accountKey={node}
                    provider={provider}
                    connectorId={connectorId}
                    canDisconnect={
                      connector.canDelete
                      && node.externalAccountId !== connector.initialAccountExternalId
                    }
                    onDisconnected={markAccountPending}
                  />
                ))}
                {pending.map(account => (
                  <PendingAccountRow
                    key={account.externalAccountId}
                    account={account}
                    provider={provider}
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
      {failure == null && connector.canEnable && pending.length > 0 && (
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
  account,
  provider,
  selected,
  selectable,
  onSelectedChange,
}: {
  account: DiscoveredAccount;
  provider: string;
  selected: boolean;
  selectable: boolean;
  onSelectedChange: (checked: boolean) => void;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { identity, name, title } = connectorCard();

  return (
    <TonedCard
      tone="sand"
      size={2}
      icon={(
        <ThirdPartyLogo thirdParty={provider} />
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

type ResolvedProbe
  = { kind: "missing" }
    | { kind: "failed" }
    | {
      kind: "ready";
      accounts: DiscoveredAccount[];
      status: ConnectorConnectionStatus;
    };

interface StoredProbe {
  key: string;
  value: ResolvedProbe;
}

type DiscoveryListing
  = { missing: true }
    | {
      missing: false;
      connectionStatus: ConnectorConnectionStatus;
      accounts: DiscoveredAccount[];
    };

function probeKey(connectorId: string, fetchKey: number): string {
  return `${connectorId}:${fetchKey}`;
}

function connectionFailure(
  status: ConnectorConnectionStatus,
  t: (key: string, options: { provider: string }) => string,
  providerName: string,
): string | null {
  const signal = connectionSignalFrom({ connectionStatus: status });
  const issue = signal == null ? null : presentConnection(signal).issue;
  if (issue == null) {
    return null;
  }

  return t(`listPage.connectionIssues.${issue}`, { provider: providerName });
}

async function loadDiscoveredAccounts(
  environment: Environment,
  connectorId: string,
): Promise<DiscoveryListing> {
  const data = await fetchQuery<ConnectorAccountsDrawerDiscoveryQuery>(
    environment,
    connectorAccountsDrawerDiscoveryQuery,
    { connectorId },
    { fetchPolicy: "network-only" },
  ).toPromise();

  if (data?.connector?.__typename !== "Connector") {
    return { missing: true };
  }

  const connector = data.connector;
  const storedIds = await collectStoredAccountIds(
    environment,
    connectorId,
    {
      pageInfo: connector.accounts.pageInfo,
      edges: connector.accounts.edges.map(edge => ({
        node: { externalAccountId: edge.node.externalAccountId },
      })),
    },
  );

  return {
    missing: false,
    connectionStatus: connector.connectionStatus,
    accounts: labelDiscoveredAccounts(connector.discoveredAccounts, storedIds),
  };
}
