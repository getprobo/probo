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

import { type TaskDeadlineBounds, taskMatchesDeadline, taskStates } from "@probo/tasks";
import { ConnectionHandler } from "react-relay";
import type { RecordSourceSelectorProxy } from "relay-runtime";

import { adjustTotalCount } from "./moveTaskListEdge";
import {
  rememberedTaskListDeadlines,
  taskListConnectionFilters,
  taskListConnectionKeys,
} from "./taskPath";

type TaskRecord = NonNullable<ReturnType<RecordSourceSelectorProxy["get"]>>;

function readString(record: TaskRecord, field: string) {
  const value = record.getValue(field);
  return typeof value === "string" ? value : null;
}

export function deleteTaskListEdges(
  store: RecordSourceSelectorProxy,
  {
    taskId,
    organizationId,
    query,
    deadline,
  }: {
    taskId: string;
    organizationId: string;
    query: string | null;
    deadline: TaskDeadlineBounds;
  },
) {
  const viewer = store.getRoot().getLinkedRecord("viewer");
  const task = store.get(taskId);
  const taskState = task == null ? null : readString(task, "state");
  if (viewer == null) {
    return;
  }

  const queries = query == null ? [null] : [query, null];
  const deadlines = [deadline].concat(rememberedTaskListDeadlines());
  const seenDeadlines = new Set<string>();
  for (const state of taskStates) {
    for (const filterQuery of queries) {
      for (const filterDeadline of deadlines) {
        const deadlineKey = `${filterDeadline.deadlineOnOrBefore ?? ""}\0${filterDeadline.hasDeadline ?? ""}`;
        if (seenDeadlines.has(`${state}\0${filterQuery ?? ""}\0${deadlineKey}`)) {
          continue;
        }
        seenDeadlines.add(`${state}\0${filterQuery ?? ""}\0${deadlineKey}`);
        const connection = ConnectionHandler.getConnection(
          viewer,
          taskListConnectionKeys[state],
          taskListConnectionFilters(organizationId, filterQuery, state, filterDeadline),
        );
        if (connection == null) {
          continue;
        }
        const hadEdge = (connection.getLinkedRecords("edges") ?? []).some(
          edge => edge?.getLinkedRecord("node")?.getDataID() === taskId,
        );
        if (!hadEdge) {
          if (task != null && taskState === state && matchesTaskQuery(task, filterQuery)) {
            const storedDeadline = readString(task, "deadline");
            if (taskMatchesDeadline(storedDeadline, filterDeadline) === true) {
              adjustTotalCount(connection, -1);
            }
          }
          continue;
        }
        ConnectionHandler.deleteNode(connection, taskId);
        adjustTotalCount(connection, -1);
      }
    }
  }
}

function matchesTaskQuery(
  task: NonNullable<ReturnType<RecordSourceSelectorProxy["get"]>>,
  query: string | null,
) {
  if (query == null || query === "") {
    return true;
  }
  const name = task.getValue("name");
  return typeof name === "string" && name.toLowerCase().includes(query.toLowerCase());
}
