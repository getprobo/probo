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
  displayName?: string;
  avatarURL?: string;
  // False when the peer is still connected but idle or in a hidden tab — the
  // facepile dims them; the caret still draws so every open tab stays visible.
  active: boolean;
  anchorPosition: number;
  headPosition: number;
};

export type RichEditorCollaborationHandle = DocHandle<RichEditorAutomergeDocument> & {
  updatePresence?: (anchorPosition: number, headPosition: number) => void;
  onPresence?: (listener: (presences: RichEditorPresence[]) => void) => () => void;
  onReview?: (listener: () => void) => () => void;
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

export function decorationsForPresences(
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
    // One color per person, so every tab's caret matches their avatar ring.
    const seed = presence.identityID || presence.connectionID;

    if (from !== to) {
      decorations.push(
        Decoration.inline(
          from,
          to,
          {
            class: "presence-selection",
            style: `--presence-highlight: ${presenceHighlightColor(seed)}`,
          },
        ),
      );
    }
    decorations.push(
      Decoration.widget(
        head,
        () => {
          const cursor = document.createElement("span");
          const label = presence.displayName?.trim() || "Collaborator";
          cursor.className = "presence-cursor";
          cursor.setAttribute("aria-label", `${label} cursor`);
          cursor.style.setProperty("--presence-color", presenceColor(seed));

          if (presence.displayName?.trim()) {
            const flag = document.createElement("span");
            flag.className = "presence-cursor-flag";
            flag.textContent = presence.displayName.trim();
            cursor.appendChild(flag);
          }

          return cursor;
        },
        {
          // The caret is rebuilt only when it moves, which replays the name
          // flag's flash. A heartbeat in place keeps the same DOM.
          key: `${presence.connectionID}:${anchor}:${head}`,
          // Negative side keeps the caret before text inserted at this position,
          // matching the Automerge cursor's move: "before".
          side: -1,
        },
      ),
    );
  }

  return DecorationSet.create(pmDocument, decorations);
}

function presenceHue(seed: string): number {
  let hash = 0;
  for (const character of seed) {
    hash = ((hash << 5) - hash + character.charCodeAt(0)) | 0;
  }
  return Math.abs(hash) % 360;
}

export function presenceColor(seed: string): string {
  return `hsl(${presenceHue(seed)} 70% 45%)`;
}

// Translucent fill for a remote selection range. Must stay a valid CSS color —
// appending a hex alpha to hsl() is invalid and browsers drop the highlight.
export function presenceHighlightColor(seed: string): string {
  return `hsl(${presenceHue(seed)} 70% 45% / 0.28)`;
}

function clamp(value: number, minimum: number, maximum: number): number {
  return Math.min(Math.max(value, minimum), maximum);
}
