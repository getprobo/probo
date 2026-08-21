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
import pmToAm from "@automerge/prosemirror/dist/pmToAm.js";
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

        collaborationDebug("remote-patches", {
          patchCount: patches.length,
          patches: patches.map(patch => ({
            action: patch.action,
            path: patch.path,
          })),
          heads: Automerge.getHeads(doc),
          spans: summarizeAutomergeSpans(doc),
          prosemirror: summarizeProseMirrorDocument(view.state.doc),
        });

        const transaction = hasHorizontalRule(view.state.doc)
          ? reconcileAutomergeDocument(adapter, path, doc, view.state)
          : patchesToTr({
              adapter,
              path,
              before: patchInfo.before,
              after: doc,
              patches,
              state: view.state,
            });
        ignoreTransaction = true;
        view.dispatch(transaction);
        ignoreTransaction = false;
        collaborationDebug("remote-applied", {
          stepCount: transaction.steps.length,
          prosemirror: summarizeProseMirrorDocument(view.state.doc),
        });
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
      if (changedTransactions.length === 0) return undefined;

      collaborationDebug("local-before", {
        transactionCount: changedTransactions.length,
        steps: changedTransactions.flatMap(transaction =>
          transaction.steps.map(step => step.constructor.name),
        ),
        heads: Automerge.getHeads(handle.doc()),
        spans: summarizeAutomergeSpans(handle.doc()),
        prosemirror: summarizeProseMirrorDocument(state.doc),
        structuralLeaf: hasHorizontalRule(state.doc),
      });

      ignoreTransaction = true;
      handle.change((document) => {
        if (
          hasHorizontalRule(state.doc)
          || hasMarkStep(changedTransactions)
        ) {
          Automerge.updateSpans(
            document,
            path,
            pmNodeToSpans(adapter, state.doc),
            adapter.updateSpansConfig(),
          );
        } else {
          for (const transaction of changedTransactions) {
            const spans = Automerge.spans(document, path);
            pmToAm(
              adapter,
              spans,
              transaction.steps,
              document,
              transaction.docs[0],
              path,
            );
          }
        }
      });
      ignoreTransaction = false;

      collaborationDebug("local-after", {
        heads: Automerge.getHeads(handle.doc()),
        spans: summarizeAutomergeSpans(handle.doc()),
        prosemirror: summarizeProseMirrorDocument(state.doc),
      });

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

function hasHorizontalRule(document: {
  descendants: (
    callback: (node: { type: { name: string } }) => boolean,
  ) => void;
}): boolean {
  let found = false;
  document.descendants((node) => {
    if (node.type.name === "horizontalRule") {
      found = true;
      return false;
    }

    return true;
  });
  return found;
}

function hasMarkStep(transactions: Transaction[]): boolean {
  return transactions.some(transaction =>
    transaction.steps.some((step) => {
      const stepType = step.constructor.name;
      return stepType === "AddMarkStep" || stepType === "RemoveMarkStep";
    }),
  );
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

function isSurrogatePair(value: string): boolean {
  if (value.length !== 2) return false;
  const first = value.charCodeAt(0);
  const second = value.charCodeAt(1);
  return first >= 0xdc00
    && first <= 0xdfff
    && second >= 0xd800
    && second <= 0xdbff;
}
