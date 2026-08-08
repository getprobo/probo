// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { Extension } from "@tiptap/core";
import { Plugin } from "@tiptap/pm/state";

import { collaborationIDAttribute } from "./automergeTable";

const tableNodeNames = new Set([
  "table",
  "tableRow",
  "tableCell",
  "tableHeader",
]);

export const AutomergeTableIDExtension = Extension.create({
  name: "automergeTableIDs",

  addGlobalAttributes() {
    return [{
      types: [...tableNodeNames],
      attributes: {
        [collaborationIDAttribute]: {
          default: null,
          parseHTML: element =>
            element.getAttribute("data-collaboration-id"),
          renderHTML: attributes => {
            const id = attributes[collaborationIDAttribute];
            return typeof id === "string" && id
              ? { "data-collaboration-id": id }
              : {};
          },
        },
      },
    }];
  },

  addProseMirrorPlugins() {
    return [
      new Plugin({
        appendTransaction(transactions, _oldState, state) {
          if (!transactions.some(transaction => transaction.docChanged)) {
            return;
          }

          const transaction = state.tr;
          state.doc.descendants((node, position) => {
            if (
              tableNodeNames.has(node.type.name)
              && !node.attrs[collaborationIDAttribute]
            ) {
              transaction.setNodeMarkup(position, undefined, {
                ...node.attrs,
                [collaborationIDAttribute]: crypto.randomUUID(),
              });
            }
          });
          return transaction.docChanged ? transaction : undefined;
        },
      }),
    ];
  },
});
