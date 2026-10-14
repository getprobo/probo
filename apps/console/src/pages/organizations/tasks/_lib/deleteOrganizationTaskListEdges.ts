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

import { taskStates } from "@probo/tasks";
import { ConnectionHandler } from "react-relay";
import type { RecordSourceSelectorProxy } from "relay-runtime";

import { adjustTotalCount } from "./moveOrganizationTaskListEdge";
import {
  organizationTaskListConnectionFilters,
  type OrganizationTaskListFilter,
  organizationTaskListFilterVariants,
  taskListConnectionKeys,
} from "./taskPath";

export function deleteOrganizationTaskListEdges(
  store: RecordSourceSelectorProxy,
  {
    taskId,
    organizationId,
    filter,
  }: {
    taskId: string;
    organizationId: string;
    filter: OrganizationTaskListFilter;
  },
) {
  const organization = store.get(organizationId);
  if (organization == null) {
    return;
  }

  for (const variant of organizationTaskListFilterVariants(filter)) {
    for (const state of taskStates) {
      const connection = ConnectionHandler.getConnection(
        organization,
        taskListConnectionKeys[state],
        organizationTaskListConnectionFilters(variant, state),
      );
      if (connection == null) {
        continue;
      }
      const hadEdge = (connection.getLinkedRecords("edges") ?? []).some(
        edge => edge?.getLinkedRecord("node")?.getDataID() === taskId,
      );
      if (!hadEdge) {
        continue;
      }
      ConnectionHandler.deleteNode(connection, taskId);
      adjustTotalCount(connection, -1);
    }
  }
}
