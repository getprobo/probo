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

import type { fetchDocumentQueuePageApprovalsQuery } from "#/__generated__/core/fetchDocumentQueuePageApprovalsQuery.graphql";
import type { fetchDocumentQueuePageSignaturesQuery } from "#/__generated__/core/fetchDocumentQueuePageSignaturesQuery.graphql";
import { coreEnvironment } from "#/lib/relay/environment";

import {
  DOCUMENT_QUEUE_ID_PAGE_SIZE,
  type DocumentQueueKind,
  type DocumentQueuePage,
} from "./documentQueue";

const fetchDocumentQueuePageSignaturesQuery = graphql`
  query fetchDocumentQueuePageSignaturesQuery(
    $organizationId: ID!
    $first: Int!
    $after: CursorKey
  ) @throwOnFieldError {
    viewer @required(action: THROW) {
      pendingQueue: signableDocuments(
        organizationId: $organizationId
        first: $first
        after: $after
        filter: { signed: false }
        orderBy: { field: UPDATED_AT, direction: DESC }
      ) {
        totalCount
        pageInfo {
          hasNextPage
          endCursor
        }
        edges {
          node {
            id
          }
        }
      }
    }
  }
`;

const fetchDocumentQueuePageApprovalsQuery = graphql`
  query fetchDocumentQueuePageApprovalsQuery(
    $organizationId: ID!
    $first: Int!
    $after: CursorKey
  ) @throwOnFieldError {
    viewer @required(action: THROW) {
      pendingQueue: approvableDocuments(
        organizationId: $organizationId
        first: $first
        after: $after
        filter: { approvalStates: [PENDING] }
        orderBy: { field: UPDATED_AT, direction: DESC }
      ) {
        totalCount
        pageInfo {
          hasNextPage
          endCursor
        }
        edges {
          node {
            id
          }
        }
      }
    }
  }
`;

type FetchDocumentQueuePageOptions = {
  kind: DocumentQueueKind;
  organizationId: string;
  after: string;
};

function toPage(connection: {
  totalCount: number;
  pageInfo: { hasNextPage: boolean; endCursor: string | null | undefined };
  edges: ReadonlyArray<{ node: { id: string } }>;
}): DocumentQueuePage {
  return {
    ids: connection.edges.map(({ node }) => node.id),
    totalCount: connection.totalCount,
    endCursor: connection.pageInfo.endCursor ?? null,
    hasNextPage: connection.pageInfo.hasNextPage,
  };
}

// Loads the next pending-ID page for the frozen queue. Uses the core
// environment because DocumentQueueProvider sits above RelayProvider.
export async function fetchDocumentQueuePage(
  options: FetchDocumentQueuePageOptions,
): Promise<DocumentQueuePage> {
  const variables = {
    organizationId: options.organizationId,
    first: DOCUMENT_QUEUE_ID_PAGE_SIZE,
    after: options.after,
  };

  if (options.kind === "signatures") {
    const data = await fetchQuery<fetchDocumentQueuePageSignaturesQuery>(
      coreEnvironment,
      fetchDocumentQueuePageSignaturesQuery,
      variables,
    ).toPromise();
    if (data == null) {
      throw new Error("document queue page is empty");
    }
    return toPage(data.viewer.pendingQueue);
  }

  const data = await fetchQuery<fetchDocumentQueuePageApprovalsQuery>(
    coreEnvironment,
    fetchDocumentQueuePageApprovalsQuery,
    variables,
  ).toPromise();
  if (data == null) {
    throw new Error("document queue page is empty");
  }
  return toPage(data.viewer.pendingQueue);
}
