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

import {
  type TaskDeadlineBounds,
  taskMatchesDeadline,
  taskPriorities,
  type TaskPriority,
  type TaskState,
  taskStates,
} from "@probo/tasks";
import { ConnectionHandler } from "react-relay";
import type {
  Environment,
  RecordSourceSelectorProxy,
} from "relay-runtime";

import {
  taskListConnectionFilters,
  taskListConnectionKeys,
} from "./taskPath";

type TaskRecord = NonNullable<ReturnType<RecordSourceSelectorProxy["get"]>>;

export interface MoveTaskListEdgeOptions {
  fromState: TaskState;
  toState: TaskState;
  organizationId: string;
  queries: readonly (string | null)[];
  deadline: TaskDeadlineBounds;
}

export interface TaskListStateMoveOptions {
  taskId: string;
  nextState?: TaskState | null;
  organizationId: string;
  query: string | null;
  deadline: TaskDeadlineBounds;
}

function isTaskState(value: unknown): value is TaskState {
  return typeof value === "string"
    && (taskStates as readonly string[]).includes(value);
}

function readStoredTaskState(environment: Environment, taskId: string): TaskState | null {
  const record = environment.getStore().getSource().get(taskId) as
    | { readonly state?: unknown }
    | null
    | undefined;
  if (record == null || !isTaskState(record.state)) {
    return null;
  }

  return record.state;
}

function findEdge(connection: TaskRecord, nodeId: string) {
  return (connection.getLinkedRecords("edges") ?? [])
    .find(existing => existing?.getLinkedRecord("node")?.getDataID() === nodeId);
}

export function adjustTotalCount(connection: TaskRecord, delta: number) {
  const current = connection.getValue("totalCount");
  if (typeof current !== "number") {
    return;
  }

  connection.setValue(Math.max(0, current + delta), "totalCount");
}

function priorityRank(node: TaskRecord) {
  const priorityIndex = taskPriorities.indexOf(node.getValue("priority") as TaskPriority);
  const rank = node.getValue("rank");
  return (
    (priorityIndex === -1 ? taskPriorities.length : priorityIndex) * 1_000_000
    + (typeof rank === "number" ? rank : 0)
  );
}

function priorityRankCursorValue(node: TaskRecord) {
  const priorityIndex = taskPriorities.indexOf(node.getValue("priority") as TaskPriority);
  const rank = node.getValue("rank");
  if (priorityIndex === -1 || typeof rank !== "number") {
    return null;
  }

  return (priorityIndex + 1) * 1_000_000 + rank;
}

function setPriorityRankCursor(edge: TaskRecord, node: TaskRecord) {
  const value = priorityRankCursorValue(node);
  if (value == null) {
    return;
  }

  const json = JSON.stringify([node.getDataID(), value]);
  edge.setValue(
    btoa(json).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/u, ""),
    "cursor",
  );
}

function sortsAfter(
  existing: TaskRecord,
  incomingRank: number,
  incomingId: string,
) {
  const existingRank = priorityRank(existing);
  if (existingRank !== incomingRank) {
    return existingRank > incomingRank;
  }

  return existing.getDataID() > incomingId;
}

function placeEdgeByPriority(
  connection: TaskRecord,
  edge: TaskRecord,
  node: TaskRecord,
) {
  setPriorityRankCursor(edge, node);
  const nodeId = node.getDataID();
  const incomingRank = priorityRank(node);
  const edges = (connection.getLinkedRecords("edges") ?? [])
    .filter(existing => existing?.getLinkedRecord("node")?.getDataID() !== nodeId);
  const insertAt = edges.findIndex((existing) => {
    const existingNode = existing?.getLinkedRecord("node");
    return existingNode != null && sortsAfter(existingNode, incomingRank, nodeId);
  });
  connection.setLinkedRecords(
    insertAt === -1
      ? [...edges, edge]
      : [...edges.slice(0, insertAt), edge, ...edges.slice(insertAt)],
    "edges",
  );
}

function getViewer(store: RecordSourceSelectorProxy) {
  return store.getRoot().getLinkedRecord("viewer");
}

function uniqueQueries(queries: readonly (string | null)[]) {
  return [...new Set(queries)];
}

export function moveTaskListEdge(
  store: RecordSourceSelectorProxy,
  node: TaskRecord,
  {
    fromState,
    toState,
    organizationId,
    queries,
    deadline,
  }: MoveTaskListEdgeOptions,
) {
  if (fromState === toState) {
    return;
  }

  const viewer = getViewer(store);
  if (viewer == null) {
    return;
  }

  const nodeId = node.getDataID();
  const deadlineValue = node.getValue("deadline");
  const storedDeadline = typeof deadlineValue === "string" ? deadlineValue : null;
  for (const query of uniqueQueries(queries)) {
    const fromConnection = ConnectionHandler.getConnection(
      viewer,
      taskListConnectionKeys[fromState],
      taskListConnectionFilters(organizationId, query, fromState, deadline),
    );
    const toConnection = ConnectionHandler.getConnection(
      viewer,
      taskListConnectionKeys[toState],
      taskListConnectionFilters(organizationId, query, toState, deadline),
    );

    if (toConnection != null && findEdge(toConnection, nodeId) != null) {
      if (fromConnection != null && findEdge(fromConnection, nodeId) != null) {
        ConnectionHandler.deleteNode(fromConnection, nodeId);
        adjustTotalCount(fromConnection, -1);
      }
      continue;
    }

    const hadFrom = fromConnection != null && findEdge(fromConnection, nodeId) != null;
    if (hadFrom && fromConnection != null) {
      ConnectionHandler.deleteNode(fromConnection, nodeId);
      adjustTotalCount(fromConnection, -1);
    }

    if (toConnection == null) {
      continue;
    }

    const matches = taskMatchesDeadline(storedDeadline, deadline);
    if (matches === false || (matches == null && !hadFrom)) {
      continue;
    }

    placeEdgeByPriority(
      toConnection,
      ConnectionHandler.createEdge(store, toConnection, node, "TaskEdge"),
      node,
    );
    adjustTotalCount(toConnection, 1);
  }
}

export function resortTaskListEdge(
  store: RecordSourceSelectorProxy,
  node: TaskRecord,
  {
    state,
    organizationId,
    queries,
    deadline,
  }: {
    state: TaskState;
    organizationId: string;
    queries: readonly (string | null)[];
    deadline: TaskDeadlineBounds;
  },
) {
  const viewer = getViewer(store);
  if (viewer == null) {
    return;
  }

  const nodeId = node.getDataID();
  for (const query of uniqueQueries(queries)) {
    const connection = ConnectionHandler.getConnection(
      viewer,
      taskListConnectionKeys[state],
      taskListConnectionFilters(organizationId, query, state, deadline),
    );
    if (connection == null) {
      continue;
    }
    const edge = findEdge(connection, nodeId);
    if (edge != null) {
      placeEdgeByPriority(connection, edge, node);
    }
  }
}

export function taskListStateMoveUpdaters(
  environment: Environment,
  {
    taskId,
    nextState,
    organizationId,
    query,
    deadline,
  }: TaskListStateMoveOptions,
) {
  const fromState = nextState == null ? null : readStoredTaskState(environment, taskId);
  const queries = query == null ? [null] : [query, null];

  function apply(store: RecordSourceSelectorProxy) {
    if (fromState == null || nextState == null || fromState === nextState) {
      return;
    }

    const node = store.get(taskId);
    if (node == null) {
      return;
    }

    moveTaskListEdge(store, node, {
      fromState,
      toState: nextState,
      organizationId,
      queries,
      deadline,
    });
  }

  return {
    optimisticUpdater(store: RecordSourceSelectorProxy) {
      const node = store.get(taskId);
      if (node != null && nextState != null) {
        node.setValue(nextState, "state");
      }
      apply(store);
    },
    updater(store: RecordSourceSelectorProxy) {
      apply(store);
    },
  };
}
