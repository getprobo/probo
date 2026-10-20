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

import { taskMatchesDeadline, type TaskPriority, type TaskState, taskStates } from "@probo/tasks";
import { ConnectionHandler } from "react-relay";
import type { Environment, RecordSourceSelectorProxy } from "relay-runtime";

import { placeEdgeSorted } from "./taskConnectionOrder";
import {
  organizationTaskListConnectionFilters,
  type OrganizationTaskListFilter,
  organizationTaskListFilterVariants,
  taskInternalControlIdsConnectionKey,
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

export function readStoredDeadline(environment: Environment, taskId: string): string | null {
  const record = environment.getStore().getSource().get(taskId) as
    | { readonly deadline?: unknown }
    | null
    | undefined;
  return typeof record?.deadline === "string" ? record.deadline : null;
}

export function readStoredAssignedToId(environment: Environment, taskId: string): string | null {
  const record = environment.getStore().getSource().get(taskId) as
    | { readonly assignedTo?: { readonly __ref?: unknown } | null }
    | null
    | undefined;
  const ref = record?.assignedTo?.__ref;
  return typeof ref === "string" ? ref : null;
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

interface TaskListMembership {
  assignedToId: string | null;
  internalControlIds: readonly string[] | null;
  name: string | null;
  deadline: string | null;
}

function readDeadline(node: TaskRecord) {
  const deadline = node.getValue("deadline");
  return typeof deadline === "string" ? deadline : null;
}

function readName(node: TaskRecord) {
  const name = node.getValue("name");
  return typeof name === "string" ? name : null;
}

function readAssignedToId(node: TaskRecord) {
  return node.getLinkedRecord("assignedTo")?.getDataID() ?? null;
}

function readInternalControlIds(node: TaskRecord) {
  const connection = ConnectionHandler.getConnection(
    node,
    taskInternalControlIdsConnectionKey,
  );
  if (connection == null) {
    return null;
  }

  return (connection.getLinkedRecords("edges") ?? []).flatMap((edge) => {
    const id = edge?.getLinkedRecord("node")?.getDataID();
    return id == null ? [] : [id];
  });
}

function readMembership(
  node: TaskRecord,
  internalControlIds?: readonly string[] | null,
): TaskListMembership {
  return {
    assignedToId: readAssignedToId(node),
    internalControlIds: internalControlIds === undefined
      ? readInternalControlIds(node)
      : internalControlIds,
    name: readName(node),
    deadline: readDeadline(node),
  };
}

// null means the control link is not loaded, so membership cannot be proved.
function matchesOrganizationListFilter(
  membership: TaskListMembership,
  filter: OrganizationTaskListFilter,
): boolean | null {
  if (filter.assignedToId != null && membership.assignedToId !== filter.assignedToId) {
    return false;
  }
  if (filter.hasAssignee === false && membership.assignedToId != null) {
    return false;
  }
  if (filter.query != null && filter.query !== "") {
    const name = membership.name ?? "";
    if (!name.toLowerCase().includes(filter.query.toLowerCase())) {
      return false;
    }
  }
  if (filter.internalControlId != null) {
    if (membership.internalControlIds == null) {
      return null;
    }
    if (!membership.internalControlIds.includes(filter.internalControlId)) {
      return false;
    }
  }
  const deadlineMatch = taskMatchesDeadline(membership.deadline, filter);
  if (deadlineMatch == null) {
    return null;
  }
  if (!deadlineMatch) {
    return false;
  }
  return true;
}

export function taskBelongsInOrganizationList(
  node: TaskRecord,
  filter: OrganizationTaskListFilter,
  internalControlIds?: readonly string[] | null,
) {
  return matchesOrganizationListFilter(readMembership(node, internalControlIds), filter);
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

    const hadFrom = fromConnection != null && findEdge(fromConnection, nodeId) != null;
    if (hadFrom && fromConnection != null) {
      ConnectionHandler.deleteNode(fromConnection, nodeId);
      adjustTotalCount(fromConnection, -1);
    }

    if (toConnection == null || findEdge(toConnection, nodeId) != null) {
      continue;
    }

    const membership = taskBelongsInOrganizationList(node, filter);
    if (membership === false || (membership == null && !hadFrom)) {
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
function shiftRanksAtOrAfter(
  connection: TaskRecord,
  nodeId: string,
  priority: TaskPriority,
  nextRank: number,
  shifted: Set<string>,
) {
  for (const edge of connection.getLinkedRecords("edges") ?? []) {
    const other = edge?.getLinkedRecord("node");
    if (other == null || other.getDataID() === nodeId || shifted.has(other.getDataID())) {
      continue;
    }
    shifted.add(other.getDataID());
    if (other.getValue("priority") !== priority) {
      continue;
    }
    const rank = readRank(other);
    if (rank == null || rank < nextRank) {
      continue;
    }
    other.setValue(rank + 1, "rank");
  }
}

function shiftSiblingRanks(
  connection: TaskRecord,
  nodeId: string,
  priority: string,
  previousRank: number,
  nextRank: number,
  shifted: Set<string>,
) {
  if (previousRank === nextRank) {
    return;
  }

  const low = Math.min(previousRank, nextRank);
  const high = Math.max(previousRank, nextRank);
  const delta = nextRank < previousRank ? 1 : -1;
  for (const edge of connection.getLinkedRecords("edges") ?? []) {
    const other = edge?.getLinkedRecord("node");
    if (other == null || other.getDataID() === nodeId || shifted.has(other.getDataID())) {
      continue;
    }
    shifted.add(other.getDataID());
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
  const shifted = new Set<string>();
  for (const filter of filters) {
    const connection = ConnectionHandler.getConnection(
      organization,
      taskListConnectionKeys[state],
      organizationTaskListConnectionFilters(filter, state),
    );
    if (connection == null) {
      continue;
    }
    if (nextPriority != null) {
      shiftRanksAtOrAfter(connection, nodeId, nextPriority, nextRank, shifted);
    } else if (
      typeof priority === "string"
      && previousPriority === priority
    ) {
      shiftSiblingRanks(connection, nodeId, priority, previousRank, nextRank, shifted);
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

export function reconcileOrganizationTaskListMembership(
  store: RecordSourceSelectorProxy,
  node: TaskRecord,
  {
    organizationId,
    state,
    filters,
    previousAssignedToId,
    previousInternalControlIds,
    nextInternalControlIds,
    previousDeadline,
  }: {
    organizationId: string;
    state: TaskState;
    filters: readonly OrganizationTaskListFilter[];
    previousAssignedToId: string | null;
    previousInternalControlIds: readonly string[] | null;
    nextInternalControlIds?: readonly string[] | null;
    previousDeadline?: string | null;
  },
) {
  const organization = getOrganization(store, organizationId);
  if (organization == null) {
    return;
  }

  const currentAssignedToId = readAssignedToId(node);
  const name = readName(node);
  const deadline = readDeadline(node);
  const before = matchesOrganizationListFilter.bind(null, {
    assignedToId: previousAssignedToId,
    internalControlIds: previousInternalControlIds,
    name,
    deadline: previousDeadline === undefined ? deadline : previousDeadline,
  });
  const now = matchesOrganizationListFilter.bind(null, {
    assignedToId: currentAssignedToId,
    internalControlIds: nextInternalControlIds === undefined
      ? previousInternalControlIds
      : nextInternalControlIds,
    name,
    deadline,
  });
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
    const wasMember = before(filter);
    const isMember = now(filter);
    if (edge != null && isMember === false && wasMember !== false) {
      ConnectionHandler.deleteNode(connection, nodeId);
      adjustTotalCount(connection, -1);
      continue;
    }
    if (edge == null && wasMember === false && isMember === true) {
      placeEdgeSorted(
        connection,
        ConnectionHandler.createEdge(store, connection, node, "TaskEdge"),
        node,
      );
      adjustTotalCount(connection, 1);
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
