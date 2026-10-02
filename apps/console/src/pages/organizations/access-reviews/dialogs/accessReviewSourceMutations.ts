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

export const createAccessReviewSourcesMutation = graphql`
  mutation accessReviewSourceMutationsCreateMutation(
    $input: CreateAccessReviewSourcesInput!
  ) {
    createAccessReviewSources(input: $input) {
      results {
        created
        accessReviewSourceEdge {
          node {
            id
            name
            connectorId
            createdAt
            ...AccessReviewSourceListItem_source @arguments(deferConnectionStatus: false)
          }
        }
      }
    }
  }
`;

// prependCreatedSourceEdges inserts each created edge at the top of the
// sources connection. A result with created=false is an existing source,
// and a node already present in the connection is never duplicated.
export function prependCreatedSourceEdges(
  store: RecordSourceSelectorProxy,
  connectionId: string,
) {
  const payload = store.getRootField("createAccessReviewSources");
  const connection = store.get(connectionId);
  if (!payload || !connection) return;

  const results = payload.getLinkedRecords("results") ?? [];
  const edges = connection.getLinkedRecords("edges") ?? [];
  const present = new Set(
    edges.map(edge => edge?.getLinkedRecord("node")?.getDataID()),
  );

  for (const result of results) {
    if (result?.getValue("created") !== true) continue;

    const edge = result.getLinkedRecord("accessReviewSourceEdge");
    const nodeId = edge?.getLinkedRecord("node")?.getDataID();
    if (!edge || nodeId == null || present.has(nodeId)) continue;

    ConnectionHandler.insertEdgeBefore(connection, edge);
    present.add(nodeId);
  }
}
