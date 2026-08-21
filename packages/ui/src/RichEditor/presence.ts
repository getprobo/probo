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

import type { DocHandle } from "@automerge/prosemirror";
import { Extension } from "@tiptap/core";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";

import type { RichEditorAutomergeDocument } from "./collaboration";

export type RichEditorPresence = {
  connectionID: string;
  identityID: string;
  anchorPosition: number;
  headPosition: number;
};

export type RichEditorCollaborationHandle = DocHandle<RichEditorAutomergeDocument> & {
  updatePresence?: (anchorPosition: number, headPosition: number) => void;
  onPresence?: (listener: (presences: RichEditorPresence[]) => void) => () => void;
};

const presencePluginKey = new PluginKey<DecorationSet>("automerge-presence");

export function createRichEditorPresenceExtension(
  handle: RichEditorCollaborationHandle,
): Extension {
  return Extension.create({
    name: "automergePresence",

    addProseMirrorPlugins() {
      return [
        new Plugin({
          key: presencePluginKey,
          state: {
            init: () => DecorationSet.empty,
            apply(transaction, decorations) {
              const presences = transaction.getMeta(presencePluginKey) as
                | RichEditorPresence[]
                | undefined;
              if (!presences) return decorations.map(transaction.mapping, transaction.doc);
              return decorationsForPresences(transaction.doc, presences);
            },
          },
          props: {
            decorations(state) {
              return presencePluginKey.getState(state);
            },
          },
          view: (view) => {
            const unsubscribe = handle.onPresence?.((presences) => {
              view.dispatch(view.state.tr.setMeta(presencePluginKey, presences));
            });
            return {
              destroy() {
                unsubscribe?.();
              },
            };
          },
        }),
      ];
    },
  });
}

function decorationsForPresences(
  pmDocument: ProseMirrorNode,
  presences: RichEditorPresence[],
): DecorationSet {
  const decorations: Decoration[] = [];
  for (const presence of presences) {
    const documentSize = pmDocument.content.size;
    const anchor = clamp(presence.anchorPosition, 0, documentSize);
    const head = clamp(presence.headPosition, 0, documentSize);
    const from = Math.min(anchor, head);
    const to = Math.max(anchor, head);
    const color = presenceColor(presence.identityID);

    if (from !== to) {
      decorations.push(
        Decoration.inline(
          from,
          to,
          {
            style: `background-color: ${color}33`,
          },
        ),
      );
    }
    decorations.push(
      Decoration.widget(
        head,
        () => {
          const cursor = document.createElement("span");
          cursor.setAttribute("aria-label", "Collaborator cursor");
          cursor.setAttribute("title", "Collaborator");
          cursor.style.borderLeft = `2px solid ${color}`;
          cursor.style.height = "1.2em";
          cursor.style.marginLeft = "-1px";
          cursor.style.pointerEvents = "none";
          return cursor;
        },
        {
          key: presence.connectionID,
          side: 1,
        },
      ),
    );
  }

  return DecorationSet.create(pmDocument, decorations);
}

function presenceColor(identityID: string): string {
  let hash = 0;
  for (const character of identityID) {
    hash = ((hash << 5) - hash + character.charCodeAt(0)) | 0;
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue} 70% 45%)`;
}

function clamp(value: number, minimum: number, maximum: number): number {
  return Math.min(Math.max(value, minimum), maximum);
}
