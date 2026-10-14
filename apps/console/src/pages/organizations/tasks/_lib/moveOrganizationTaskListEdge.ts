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

import { type TaskPriority, type TaskState, taskStates } from "@probo/tasks";
import { ConnectionHandler } from "react-relay";
import type { Environment, RecordSourceSelectorProxy } from "relay-runtime";

import { placeEdgeSorted } from "./taskConnectionOrder";
import {
  organizationTaskListConnectionFilters,
  type OrganizationTaskListFilter,
  organizationTaskListFilterVariants,
  taskListConnectionKeys,
} from "./taskPath";

type TaskRecord = NonNullable<ReturnType<RecordSourceSelectorProxy["get"]>>;

export interface MoveOrganizationTaskListEdgeOptions {
  fromState: TaskState;
  toState: TaskState;
  organizationId: string;
  filters: readonly OrganizationTaskListFilter[];
}

export function isTaskState(value: unknown): value is TaskState {
  return typeof value === "string"
    && (taskStates as readonly string[]).includes(value);
}

export function readStoredTaskState(environment: Environment, taskId: string): TaskState | null {
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

function getOrganization(store: RecordSourceSelectorProxy, organizationId: string) {
  return store.get(organizationId);
}

export function moveOrganizationTaskListEdge(
  store: RecordSourceSelectorProxy,
  node: TaskRecord,
  {
    fromState,
    toState,
    organizationId,
    filters,
  }: MoveOrganizationTaskListEdgeOptions,
) {
  if (fromState === toState) {
    return;
  }

  const organization = getOrganization(store, organizationId);
  if (organization == null) {
    return;
  }

  const nodeId = node.getDataID();
  for (const filter of filters) {
    const fromConnection = ConnectionHandler.getConnection(
      organization,
      taskListConnectionKeys[fromState],
      organizationTaskListConnectionFilters(filter, fromState),
    );
    const toConnection = ConnectionHandler.getConnection(
      organization,
      taskListConnectionKeys[toState],
      organizationTaskListConnectionFilters(filter, toState),
    );

    if (toConnection != null && findEdge(toConnection, nodeId) != null) {
      if (fromConnection != null && findEdge(fromConnection, nodeId) != null) {
        ConnectionHandler.deleteNode(fromConnection, nodeId);
        adjustTotalCount(fromConnection, -1);
      }
      continue;
    }

    if (fromConnection != null && findEdge(fromConnection, nodeId) != null) {
      ConnectionHandler.deleteNode(fromConnection, nodeId);
      adjustTotalCount(fromConnection, -1);
    }

    if (toConnection == null) {
      continue;
    }

    placeEdgeSorted(
      toConnection,
      ConnectionHandler.createEdge(store, toConnection, node, "TaskEdge"),
      node,
    );
    adjustTotalCount(toConnection, 1);
  }
}

function readRank(node: TaskRecord) {
  const rank = node.getValue("rank");
  return typeof rank === "number" ? rank : null;
}

// UpdateRank shifts every other task in the same state and priority whose
// rank sits between the old and new values. The mutation payload only returns
// the dragged task, so the list has to apply that shift itself.
function shiftSiblingRanks(
  connection: TaskRecord,
  nodeId: string,
  priority: string,
  previousRank: number,
  nextRank: number,
) {
  if (previousRank === nextRank) {
    return;
  }

  const low = Math.min(previousRank, nextRank);
  const high = Math.max(previousRank, nextRank);
  const delta = nextRank < previousRank ? 1 : -1;
  for (const edge of connection.getLinkedRecords("edges") ?? []) {
    const other = edge?.getLinkedRecord("node");
    if (other == null || other.getDataID() === nodeId) {
      continue;
    }
    if (other.getValue("priority") !== priority) {
      continue;
    }
    const rank = readRank(other);
    if (rank == null || rank < low || rank > high) {
      continue;
    }
    other.setValue(rank + delta, "rank");
  }
}

export function applyOrganizationTaskRankDrop(
  store: RecordSourceSelectorProxy,
  node: TaskRecord,
  organizationId: string,
  state: TaskState,
  filters: readonly OrganizationTaskListFilter[],
  {
    previousRank,
    nextRank,
    nextPriority,
  }: {
    previousRank: number;
    nextRank: number;
    nextPriority?: TaskPriority;
  },
) {
  const previousPriority = node.getValue("priority");
  node.setValue(nextRank, "rank");
  if (nextPriority != null) {
    node.setValue(nextPriority, "priority");
  }

  const organization = getOrganization(store, organizationId);
  if (organization == null) {
    return;
  }

  const nodeId = node.getDataID();
  const priority = node.getValue("priority");
  for (const filter of filters) {
    const connection = ConnectionHandler.getConnection(
      organization,
      taskListConnectionKeys[state],
      organizationTaskListConnectionFilters(filter, state),
    );
    if (connection == null) {
      continue;
    }
    if (
      nextPriority == null
      && typeof priority === "string"
      && previousPriority === priority
    ) {
      shiftSiblingRanks(connection, nodeId, priority, previousRank, nextRank);
    }
    const edge = findEdge(connection, nodeId);
    if (edge != null) {
      placeEdgeSorted(connection, edge, node);
    }
  }
}

export function resortOrganizationTaskListEdge(
  store: RecordSourceSelectorProxy,
  node: TaskRecord,
  organizationId: string,
  state: TaskState,
  filters: readonly OrganizationTaskListFilter[],
) {
  const organization = getOrganization(store, organizationId);
  if (organization == null) {
    return;
  }

  const nodeId = node.getDataID();
  for (const filter of filters) {
    const connection = ConnectionHandler.getConnection(
      organization,
      taskListConnectionKeys[state],
      organizationTaskListConnectionFilters(filter, state),
    );
    if (connection == null) {
      continue;
    }
    const edge = findEdge(connection, nodeId);
    if (edge != null) {
      placeEdgeSorted(connection, edge, node);
    }
  }
}

export function taskListStateMoveUpdaters(
  environment: Environment,
  {
    taskId,
    nextState,
    organizationId,
    filter,
  }: {
    taskId: string;
    nextState?: TaskState | null;
    organizationId: string;
    filter: OrganizationTaskListFilter;
  },
) {
  const fromState = nextState == null ? null : readStoredTaskState(environment, taskId);
  const filters = organizationTaskListFilterVariants(filter);

  function apply(store: RecordSourceSelectorProxy) {
    if (fromState == null || nextState == null || fromState === nextState) {
      return;
    }

    const node = store.get(taskId);
    if (node == null) {
      return;
    }

    moveOrganizationTaskListEdge(store, node, {
      fromState,
      toState: nextState,
      organizationId,
      filters,
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
