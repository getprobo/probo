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

import type { RecordSourceSelectorProxy } from "relay-runtime";
import { ConnectionHandler, graphql } from "relay-runtime";

export const createAccessReviewSourceMutation = graphql`
  mutation accessReviewSourceMutationsCreateMutation(
    $input: CreateAccessReviewSourceInput!
  ) {
    createAccessReviewSource(input: $input) {
      created
      accessReviewSourceEdge {
        node {
          id
          name
          connectorId
          createdAt
          ...AccessReviewSourceListItem_source
        }
      }
    }
  }
`;

// prependCreatedSourceEdge inserts the mutation's edge at the top of the
// sources connection. Creation is idempotent per connector, so a call
// that resolved to an existing source (created=false) inserts nothing,
// and a node already present in the connection is never duplicated.
export function prependCreatedSourceEdge(
  store: RecordSourceSelectorProxy,
  connectionId: string,
) {
  const payload = store.getRootField("createAccessReviewSource");
  if (!payload || payload.getValue("created") !== true) return;

  const edge = payload.getLinkedRecord("accessReviewSourceEdge");
  const node = edge?.getLinkedRecord("node");
  const connection = store.get(connectionId);
  if (!edge || !node || !connection) return;

  const nodeId = node.getDataID();
  const edges = connection.getLinkedRecords("edges") ?? [];
  if (edges.some(e => e?.getLinkedRecord("node")?.getDataID() === nodeId)) {
    return;
  }

  ConnectionHandler.insertEdgeBefore(connection, edge);
}
