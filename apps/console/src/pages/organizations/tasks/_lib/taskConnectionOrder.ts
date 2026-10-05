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

import { ConnectionHandler } from "react-relay";
import type { RecordSourceSelectorProxy } from "relay-runtime";

import { taskInternalControlIdsConnectionKey } from "./taskPath";
import { taskPriorities, type TaskPriority } from "./taskState";

type TaskRecord = NonNullable<ReturnType<RecordSourceSelectorProxy["get"]>>;

function priorityRank(node: TaskRecord) {
  const priorityIndex = taskPriorities.indexOf(
    node.getValue("priority") as TaskPriority,
  );
  return (
    (priorityIndex === -1 ? taskPriorities.length : priorityIndex) * 1_000_000
    + (typeof node.getValue("rank") === "number"
      ? (node.getValue("rank") as number)
      : 0)
  );
}

function findEdge(connection: TaskRecord, nodeId: string) {
  return (connection.getLinkedRecords("edges") ?? [])
    .find(existing => existing?.getLinkedRecord("node")?.getDataID() === nodeId);
}

function priorityRankCursorValue(node: TaskRecord) {
  const priorityIndex = taskPriorities.indexOf(
    node.getValue("priority") as TaskPriority,
  );
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

function placeEdgeSorted(
  connection: TaskRecord,
  edge: TaskRecord,
  node: TaskRecord,
) {
  setPriorityRankCursor(edge, node);
  const nodeId = node.getDataID();
  const newRank = priorityRank(node);
  const edges = (connection.getLinkedRecords("edges") ?? [])
    .filter(existing => existing?.getLinkedRecord("node")?.getDataID() !== nodeId);
  const insertAt = edges.findIndex((existing) => {
    const existingNode = existing?.getLinkedRecord("node");
    return existingNode != null && sortsAfter(existingNode, newRank, nodeId);
  });
  connection.setLinkedRecords(
    insertAt === -1
      ? [...edges, edge]
      : [...edges.slice(0, insertAt), edge, ...edges.slice(insertAt)],
    "edges",
  );
}

export function insertTaskEdgeSorted(
  store: RecordSourceSelectorProxy,
  connectionIds: readonly string[],
) {
  const edge = store.getRootField("createTask")?.getLinkedRecord("taskEdge");
  if (edge) {
    insertExistingTaskEdgeSorted(store, edge, connectionIds);
  }
}

export function insertExistingTaskEdgeSorted(
  store: RecordSourceSelectorProxy,
  edge: TaskRecord,
  connectionIds: readonly string[],
) {
  const node = edge.getLinkedRecord("node");
  if (!node) {
    return;
  }

  for (const connectionId of connectionIds) {
    const connection = store.get(connectionId);
    if (connection) {
      placeEdgeSorted(connection, edge, node);
    }
  }
}

export function moveTaskNodeSorted(
  store: RecordSourceSelectorProxy,
  node: TaskRecord,
  connections: {
    organizationConnectionId: string;
    previousInternalControlConnectionIds?: readonly string[];
    nextInternalControlConnectionIds?: readonly string[];
    createIfMissing?: boolean;
  },
) {
  const nodeId = node.getDataID();
  const organizationConnection = store.get(connections.organizationConnectionId);
  if (organizationConnection) {
    const edge = findEdge(organizationConnection, nodeId);
    if (edge) {
      placeEdgeSorted(organizationConnection, edge, node);
    }
  }

  const nextIds = new Set(connections.nextInternalControlConnectionIds ?? []);
  for (const previousId of connections.previousInternalControlConnectionIds ?? []) {
    if (nextIds.has(previousId)) {
      continue;
    }

    const previous = store.get(previousId);
    if (previous) {
      ConnectionHandler.deleteNode(previous, nodeId);
    }
  }

  for (const nextId of nextIds) {
    const next = store.get(nextId);
    if (!next) {
      continue;
    }

    const edge = findEdge(next, nodeId);
    if (!edge && !connections.createIfMissing) {
      continue;
    }

    placeEdgeSorted(
      next,
      edge ?? ConnectionHandler.createEdge(store, next, node, "TaskEdge"),
      node,
    );
  }
}

export function syncTaskInternalControlIds(
  store: RecordSourceSelectorProxy,
  task: TaskRecord,
  previousIds: readonly string[],
  nextIds: readonly string[],
) {
  const connection = ConnectionHandler.getConnection(
    task,
    taskInternalControlIdsConnectionKey,
  );
  if (!connection) {
    return;
  }

  const next = new Set(nextIds);
  for (const id of previousIds) {
    if (!next.has(id)) {
      ConnectionHandler.deleteNode(connection, id);
    }
  }

  const previous = new Set(previousIds);
  for (const id of nextIds) {
    if (previous.has(id) || connectionHasNode(connection, id)) {
      continue;
    }

    const control = store.get(id) ?? store.create(id, "InternalControl");
    if (control.getValue("id") == null) {
      control.setValue(id, "id");
    }
    ConnectionHandler.insertEdgeBefore(
      connection,
      ConnectionHandler.createEdge(store, connection, control, "InternalControlEdge"),
    );
  }
}

function connectionHasNode(connection: TaskRecord, nodeId: string) {
  return (connection.getLinkedRecords("edges") ?? [])
    .some(edge => edge?.getLinkedRecord("node")?.getDataID() === nodeId);
}
