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
import { DrawerHeader } from "@probo/ui/src/v2/Drawer/DrawerHeader";
import { DrawerTitle } from "@probo/ui/src/v2/Drawer/DrawerTitle";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, usePaginationFragment } from "react-relay";

import type { ConnectorAccountListContent_accounts$key } from "#/__generated__/core/ConnectorAccountListContent_accounts.graphql";
import type { ConnectorAccountListContent_connector$key } from "#/__generated__/core/ConnectorAccountListContent_connector.graphql";
import type { ConnectorAccountListContentAccountsQuery } from "#/__generated__/core/ConnectorAccountListContentAccountsQuery.graphql";
import {
  connectionIssueKeys,
  connectionSignalFrom,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";

import { connectorAccountsDrawer } from "../variants";

import { ConnectorAccountListItem } from "./ConnectorAccountListItem";
import { ConnectorProbeError } from "./ConnectorProbeError";

const PAGE_SIZE = 50;

const connectorAccountListContentFragment = graphql`
  fragment ConnectorAccountListContent_connector on Connector {
    id
    canReconnect
    connectionStatus
    providerOrganizations {
      status
    }
    ...ConnectorAccountListContent_accounts
    ...ConnectorProbeError_connector
    ...ConnectorAccountListItem_connector
  }
`;

const connectorAccountListContentAccountsFragment = graphql`
  fragment ConnectorAccountListContent_accounts on Connector
  @refetchable(queryName: "ConnectorAccountListContentAccountsQuery")
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

interface ConnectorAccountListContentProps {
  connectorKey: ConnectorAccountListContent_connector$key;
  onReload: (connectorId: string) => void;
}

export function ConnectorAccountListContent({
  connectorKey,
  onReload,
}: ConnectorAccountListContentProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const connector = useFragment(connectorAccountListContentFragment, connectorKey);
  const {
    data: accountsData,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    ConnectorAccountListContentAccountsQuery,
    ConnectorAccountListContent_accounts$key
  >(connectorAccountListContentAccountsFragment, connector);
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
  const { heading, title, list, empty } = connectorAccountsDrawer();

  function reload() {
    startTransition(() => {
      onReload(connector.id);
    });
  }

  return (
    <>
      <DrawerHeader>
        <div className={heading()}>
          <div className={title()}>
            <DrawerTitle>{t("detailsPage.accounts.title")}</DrawerTitle>
            <Text size={2} color="faint">{accountsData.accounts.totalCount}</Text>
          </div>
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
      <DrawerBody>
        {issues.length === 0 && stored.length === 0
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
    </>
  );
}
