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

// Adapted from @automerge/prosemirror's MIT-licensed sync plugin. The
// structural-leaf path deliberately bypasses its text-offset fast path.

import * as Automerge from "@automerge/automerge";
import type {
  DocHandle,
  SchemaAdapter,
} from "@automerge/prosemirror";
import { patchesToTr } from "@automerge/prosemirror/dist/patchesToTr.js";
import {
  pmDocFromSpans,
  pmNodeToSpans,
} from "@automerge/prosemirror/dist/traversal.js";
import type { Fragment } from "@tiptap/pm/model";
import {
  type EditorState,
  Plugin,
  PluginKey,
  type Transaction,
} from "@tiptap/pm/state";

import type { RichEditorAutomergeDocument } from "./collaboration";
import { applyDocumentSteps, reassignSuggestionBlocks } from "./documentSteps";
import {
  suggestionReassignmentFrom,
  suggestionSplitFrom,
} from "./suggestionSplit";

export const automergeRemoteTransaction = "automerge-remote";
import {
  collaborationDebug,
  summarizeAutomergeSpans,
  summarizeProseMirrorDocument,
} from "./collaborationDebug";

export const automergeSyncPluginKey = new PluginKey("automerge-sync");

export function createAutomergeSyncPlugin(
  adapter: SchemaAdapter,
  handle: DocHandle<RichEditorAutomergeDocument>,
  path: Automerge.Prop[],
): Plugin {
  let ignoreTransaction = false;

  return new Plugin({
    key: automergeSyncPluginKey,
    view: (view) => {
      const onPatch = ({
        doc,
        patches,
        patchInfo,
      }: {
        doc: Automerge.Doc<RichEditorAutomergeDocument>;
        patches: Automerge.Patch[];
        patchInfo: Automerge.PatchInfo<RichEditorAutomergeDocument>;
      }) => {
        if (ignoreTransaction) return;

        collaborationDebug("remote-patches", () => ({
          patchCount: patches.length,
          patches: patches.map(patch => ({
            action: patch.action,
            path: patch.path,
          })),
          heads: Automerge.getHeads(doc),
          spans: summarizeAutomergeSpans(doc),
          prosemirror: summarizeProseMirrorDocument(view.state.doc),
        }));

        // Full reconcile only when a horizontal rule appears or disappears.
        // Presence of an HR elsewhere must not force a rewrite on every patch.
        const beforeHadRule = spansHaveHorizontalRule(patchInfo.before, path);
        const afterHasRule = spansHaveHorizontalRule(doc, path);
        const transaction = beforeHadRule !== afterHasRule
          ? reconcileAutomergeDocument(adapter, path, doc, view.state)
          : patchesToTr({
              adapter,
              path,
              before: patchInfo.before,
              after: doc,
              patches,
              state: view.state,
            });
        transaction.setMeta(automergeRemoteTransaction, true);
        ignoreTransaction = true;
        try {
          view.dispatch(transaction);
        } finally {
          ignoreTransaction = false;
        }
        collaborationDebug("remote-applied", () => ({
          stepCount: transaction.steps.length,
          prosemirror: summarizeProseMirrorDocument(view.state.doc),
        }));
      };

      handle.on("change", onPatch);

      return {
        destroy() {
          handle.off("change", onPatch);
        },
      };
    },
    appendTransaction(transactions, _oldState, state) {
      if (ignoreTransaction) return undefined;

      const changedTransactions = transactions.filter(
        transaction => transaction.docChanged,
      );
      const reassignment = suggestionReassignmentFrom(transactions);
      if (changedTransactions.length === 0) {
        if (reassignment) {
          ignoreTransaction = true;
          try {
            handle.change((document) => {
              reassignSuggestionBlocks(
                document,
                path,
                reassignment.from,
                reassignment.to,
              );
            });
          } finally {
            ignoreTransaction = false;
          }
        }
        return undefined;
      }

      const touchesRule = changedTransactions.some(transactionTouchesHorizontalRule);

      collaborationDebug("local-before", () => ({
        transactionCount: changedTransactions.length,
        steps: changedTransactions.flatMap(transaction =>
          transaction.steps.map(step => step.constructor.name),
        ),
        heads: Automerge.getHeads(handle.doc()),
        spans: summarizeAutomergeSpans(handle.doc()),
        prosemirror: summarizeProseMirrorDocument(state.doc),
        structuralLeaf: touchesRule,
      }));

      ignoreTransaction = true;
      try {
        handle.change((document) => {
          // Only HR insert/delete/move needs a full span rewrite. Typing next
          // to an existing rule stays on the incremental path.
          if (touchesRule) {
            Automerge.updateSpans(
              document,
              path,
              pmNodeToSpans(adapter, state.doc),
              adapter.updateSpansConfig(),
            );
            return;
          }
          for (const transaction of changedTransactions) {
            const applied = applyDocumentSteps(
              adapter,
              document,
              path,
              transaction.docs[0],
              transaction.steps,
              suggestionSplitFrom(transaction),
            );
            if (!applied) {
              Automerge.updateSpans(
                document,
                path,
                pmNodeToSpans(adapter, state.doc),
                adapter.updateSpansConfig(),
              );
              return;
            }
          }
          if (reassignment) {
            reassignSuggestionBlocks(
              document,
              path,
              reassignment.from,
              reassignment.to,
            );
          }
        });
      } finally {
        ignoreTransaction = false;
      }

      collaborationDebug("local-after", () => ({
        heads: Automerge.getHeads(handle.doc()),
        spans: summarizeAutomergeSpans(handle.doc()),
        prosemirror: summarizeProseMirrorDocument(state.doc),
      }));

      return undefined;
    },
  });
}

export function reconcileAutomergeDocument(
  adapter: SchemaAdapter,
  path: Automerge.Prop[],
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  state: EditorState,
): Transaction {
  const nextDocument = pmDocFromSpans(
    adapter,
    Automerge.spans(document, path),
  );
  const transaction = state.tr;
  const change = findDocumentDiff(state.doc.content, nextDocument.content);
  if (!change) return transaction;

  transaction.replace(
    change.start,
    change.endBefore,
    nextDocument.slice(change.start, change.endAfter),
  );
  transaction.setMeta("addToHistory", false);
  return transaction;
}

function spansHaveHorizontalRule(
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  path: Automerge.Prop[],
): boolean {
  return Automerge.spans(document, path).some((span) => {
    if (span.type !== "block") return false;
    const type = span.value.type;
    if (typeof type === "string") return type === "horizontal-rule";
    if (Automerge.isImmutableString(type)) return type.val === "horizontal-rule";
    return false;
  });
}

function transactionTouchesHorizontalRule(transaction: {
  docs: Array<{
    content: { size: number };
    nodesBetween: (
      from: number,
      to: number,
      callback: (node: { type: { name: string } }) => void | boolean,
    ) => void;
  }>;
  steps: ReadonlyArray<{
    from?: number;
    to?: number;
    slice?: {
      content: {
        descendants: (
          callback: (node: { type: { name: string } }) => void | boolean,
        ) => void;
      };
    };
  }>;
}): boolean {
  const before = transaction.docs[0];
  if (!before) return false;

  for (const step of transaction.steps) {
    if (typeof step.from === "number" && typeof step.to === "number") {
      let found = false;
      const from = Math.max(0, Math.min(step.from, before.content.size));
      const to = Math.max(from, Math.min(step.to, before.content.size));
      before.nodesBetween(from, to, (node) => {
        if (node.type.name === "horizontalRule") found = true;
        return !found;
      });
      if (found) return true;
    }

    let inserted = false;
    step.slice?.content.descendants((node) => {
      if (node.type.name === "horizontalRule") inserted = true;
      return !inserted;
    });
    if (inserted) return true;
  }

  return false;
}

function findDocumentDiff(
  before: Fragment,
  after: Fragment,
): {
  start: number;
  endBefore: number;
  endAfter: number;
} | null {
  let start = before.findDiffStart(after);
  if (start === null) return null;

  const end = before.findDiffEnd(after);
  if (!end) return null;

  let endBefore = end.a;
  let endAfter = end.b;
  if (endBefore < start && before.size < after.size) {
    if (
      start > 0
      && start < after.size
      && isSurrogatePair(after.textBetween(start - 1, start + 1))
    ) {
      start--;
    }
    endAfter = start + endAfter - endBefore;
    endBefore = start;
  } else if (endAfter < start) {
    if (
      start > 0
      && start < before.size
      && isSurrogatePair(before.textBetween(start - 1, start + 1))
    ) {
      start--;
    }
    endBefore = start + endBefore - endAfter;
    endAfter = start;
  }

  return { start, endBefore, endAfter };
}

export function isSurrogatePair(value: string): boolean {
  if (value.length !== 2) return false;
  const first = value.charCodeAt(0);
  const second = value.charCodeAt(1);
  return first >= 0xd800
    && first <= 0xdbff
    && second >= 0xdc00
    && second <= 0xdfff;
}
