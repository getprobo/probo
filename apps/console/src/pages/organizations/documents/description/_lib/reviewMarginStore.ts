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

import { ConnectionHandler, type RecordSourceSelectorProxy } from "relay-runtime";

const openThreadConnection = "DocumentReviewMargin_commentThreads";
const openSuggestionConnection = "DocumentReviewMargin_suggestions";
const commentConnection = "DocumentReviewMargin_comments";

function openThreads(store: RecordSourceSelectorProxy, documentId: string) {
  return store.get(ConnectionHandler.getConnectionID(
    documentId,
    openThreadConnection,
    { filter: { resolved: false } },
  ));
}

export function insertThreadEdge(store: RecordSourceSelectorProxy, documentId: string) {
  const edge = store.getRootField("createDocumentCommentThread")?.getLinkedRecord("documentCommentThreadEdge");
  const connection = openThreads(store, documentId);
  if (edge && connection) ConnectionHandler.insertEdgeAfter(connection, edge);
}

export function insertCommentEdge(store: RecordSourceSelectorProxy, threadId: string) {
  const edge = store.getRootField("createDocumentComment")?.getLinkedRecord("documentCommentEdge");
  const connection = store.get(ConnectionHandler.getConnectionID(threadId, commentConnection));
  if (edge && connection) ConnectionHandler.insertEdgeAfter(connection, edge);
}

export function deleteCommentEdge(store: RecordSourceSelectorProxy, threadId: string, commentId: string) {
  const connection = store.get(ConnectionHandler.getConnectionID(threadId, commentConnection));
  if (connection) ConnectionHandler.deleteNode(connection, commentId);
}

export function deleteOpenThread(store: RecordSourceSelectorProxy, documentId: string, threadId: string) {
  const connection = openThreads(store, documentId);
  if (connection) ConnectionHandler.deleteNode(connection, threadId);
}

export function insertSuggestionEdge(store: RecordSourceSelectorProxy, documentId: string) {
  const edge = store.getRootField("startDocumentSuggestion")?.getLinkedRecord("documentSuggestionEdge");
  const connection = store.get(ConnectionHandler.getConnectionID(
    documentId,
    openSuggestionConnection,
    { filter: { status: "OPEN" }, orderBy: { direction: "DESC", field: "CREATED_AT" } },
  ));
  if (edge && connection) ConnectionHandler.insertEdgeBefore(connection, edge);
}

export function closeSuggestion(store: RecordSourceSelectorProxy, documentId: string, suggestionId: string) {
  const suggestions = store.get(ConnectionHandler.getConnectionID(
    documentId,
    openSuggestionConnection,
    { filter: { status: "OPEN" }, orderBy: { direction: "DESC", field: "CREATED_AT" } },
  ));
  if (suggestions) ConnectionHandler.deleteNode(suggestions, suggestionId);
  const threads = openThreads(store, documentId);
  for (const edge of threads?.getLinkedRecords("edges") ?? []) {
    const node = edge?.getLinkedRecord("node");
    if (node?.getValue("suggestionId") === suggestionId && threads) {
      ConnectionHandler.deleteNode(threads, node.getDataID());
    }
  }
}

export function reopenThreadEdge(store: RecordSourceSelectorProxy, documentId: string, threadId: string) {
  const connection = openThreads(store, documentId);
  const node = store.get(threadId);
  if (!connection || !node) return;
  const edge = ConnectionHandler.createEdge(store, connection, node, "DocumentCommentThreadEdge");
  ConnectionHandler.insertEdgeAfter(connection, edge);
}
