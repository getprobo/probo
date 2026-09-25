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

import { fetchQuery, graphql } from "react-relay";
import type { Environment } from "relay-runtime";

import type { discoveredAccountsIdsQuery } from "#/__generated__/core/discoveredAccountsIdsQuery.graphql";

export type DiscoveredAccountStatus = "PENDING" | "CREATED";

export interface DiscoveredAccount {
  externalAccountId: string;
  name: string;
  status: DiscoveredAccountStatus;
}

export interface CreatedConnectorLocationState {
  connectorId: string;
  canEnable: boolean;
  discoveredAccounts: DiscoveredAccount[];
}

export interface StoredAccountIdPage {
  pageInfo: {
    hasNextPage: boolean;
    endCursor?: string | null;
  };
  edges: ReadonlyArray<{
    node: {
      externalAccountId: string;
    };
  }>;
}

const discoveredAccountIdsQuery = graphql`
  query discoveredAccountsIdsQuery($connectorId: ID!, $after: CursorKey) {
    connector: node(id: $connectorId) {
      __typename
      ... on Connector {
        accounts(
          first: 50
          after: $after
          orderBy: { direction: ASC, field: CREATED_AT }
        ) {
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

export function labelDiscoveredAccounts(
  discovered: ReadonlyArray<{ externalAccountId: string; name: string }>,
  storedIds: ReadonlySet<string>,
): DiscoveredAccount[] {
  return discovered.map(account => ({
    externalAccountId: account.externalAccountId,
    name: account.name,
    status: storedIds.has(account.externalAccountId) ? "CREATED" : "PENDING",
  }));
}

export function createdConnectorLocationState(state: unknown): CreatedConnectorLocationState | null {
  if (state == null || typeof state !== "object") {
    return null;
  }

  const value = state as Partial<CreatedConnectorLocationState>;
  if (typeof value.connectorId !== "string" || !Array.isArray(value.discoveredAccounts)) {
    return null;
  }

  const discoveredAccounts = value.discoveredAccounts.filter(isDiscoveredAccount);
  if (discoveredAccounts.length === 0) {
    return null;
  }

  return {
    connectorId: value.connectorId,
    canEnable: value.canEnable === true,
    discoveredAccounts,
  };
}

export async function collectStoredAccountIds(
  environment: Environment,
  connectorId: string,
  firstPage: StoredAccountIdPage,
): Promise<Set<string>> {
  const ids = new Set(firstPage.edges.map(edge => edge.node.externalAccountId));
  let hasNextPage = firstPage.pageInfo.hasNextPage;
  let after = firstPage.pageInfo.endCursor ?? null;

  while (hasNextPage && after != null) {
    const data = await fetchQuery<discoveredAccountsIdsQuery>(
      environment,
      discoveredAccountIdsQuery,
      { connectorId, after },
      { fetchPolicy: "network-only" },
    ).toPromise();
    const connection = data?.connector?.__typename === "Connector"
      ? data.connector.accounts
      : null;
    if (connection == null) {
      break;
    }

    for (const edge of connection.edges) {
      ids.add(edge.node.externalAccountId);
    }

    hasNextPage = connection.pageInfo.hasNextPage;
    after = connection.pageInfo.endCursor ?? null;
  }

  return ids;
}

function isDiscoveredAccount(value: unknown): value is DiscoveredAccount {
  if (value == null || typeof value !== "object") {
    return false;
  }

  const account = value as Partial<DiscoveredAccount>;

  return typeof account.externalAccountId === "string"
    && typeof account.name === "string"
    && (account.status === "PENDING" || account.status === "CREATED");
}
