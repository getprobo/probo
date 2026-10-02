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

import { type ReactNode, useEffect, useState } from "react";
import { graphql, usePaginationFragment } from "react-relay";

import type { AddableConnectorAccounts_connector$key } from "#/__generated__/core/AddableConnectorAccounts_connector.graphql";
import type { AddableConnectorAccountsPaginationQuery } from "#/__generated__/core/AddableConnectorAccountsPaginationQuery.graphql";
import type { AddableConnectorListItem_account$key } from "#/__generated__/core/AddableConnectorListItem_account.graphql";

// The account fragment is defined next to the card that reads it. This import
// keeps the spread above colocated with that module for the Relay lint rule.
import "./AddableConnectorListItem";

const PAGE_SIZE = 50;

const fragment = graphql`
  fragment AddableConnectorAccounts_connector on Connector
  @refetchable(queryName: "AddableConnectorAccountsPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    id
    accounts(
      first: $first
      after: $after
      orderBy: { direction: DESC, field: CREATED_AT }
    ) @connection(key: "AddableConnectorAccounts_accounts", filters: []) {
      edges {
        node {
          ...AddableConnectorListItem_account
        }
      }
    }
  }
`;

export interface LoadedConnectorAccounts {
  id: string;
  hasNext: boolean;
  accounts: AddableConnectorListItem_account$key;
}

export function ConnectorAccountPages({
  connectorKeys,
  children,
}: {
  connectorKeys: readonly AddableConnectorAccounts_connector$key[];
  children: (pages: readonly LoadedConnectorAccounts[]) => ReactNode;
}) {
  return (
    <AccountPageLevel connectorKeys={connectorKeys} index={0} pages={[]}>
      {children}
    </AccountPageLevel>
  );
}

function AccountPageLevel({
  connectorKeys,
  index,
  pages,
  children,
}: {
  connectorKeys: readonly AddableConnectorAccounts_connector$key[];
  index: number;
  pages: readonly LoadedConnectorAccounts[];
  children: (pages: readonly LoadedConnectorAccounts[]) => ReactNode;
}) {
  const connectorKey = connectorKeys[index];
  if (connectorKey == null) {
    return children(pages);
  }

  return (
    <ConnectorAccountPage connectorKey={connectorKey}>
      {page => (
        <AccountPageLevel
          connectorKeys={connectorKeys}
          index={index + 1}
          pages={[...pages, page]}
        >
          {children}
        </AccountPageLevel>
      )}
    </ConnectorAccountPage>
  );
}

function ConnectorAccountPage({
  connectorKey,
  children,
}: {
  connectorKey: AddableConnectorAccounts_connector$key;
  children: (page: LoadedConnectorAccounts) => ReactNode;
}) {
  const {
    data,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    AddableConnectorAccountsPaginationQuery,
    AddableConnectorAccounts_connector$key
  >(fragment, connectorKey);
  // hasNext stays true when a page fails, so a failed load is latched
  // to stop the effect from requesting that page again.
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (failed || !hasNext || isLoadingNext) {
      return;
    }

    loadNext(PAGE_SIZE, {
      onComplete: error => setFailed(error != null),
    });
  }, [failed, hasNext, isLoadingNext, loadNext]);

  return children({
    id: data.id,
    hasNext: hasNext && !failed,
    accounts: data.accounts.edges.map(({ node }) => node),
  });
}
