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

import { Mention, type MentionNodeAttrs, type MentionOptions } from "@tiptap/extension-mention";
import { PluginKey } from "@tiptap/pm/state";
import type { Editor } from "@tiptap/react";
import type { SuggestionProps } from "@tiptap/suggestion";

export type MentionItem = {
  id: string;
  label: string;
  description?: string | null;
  avatarUrl?: string | null;
};

export type MentionSource = (query: string, signal: AbortSignal) => Promise<MentionItem[]>;

export type MentionSuggestionState = {
  active: boolean;
  loading: boolean;
  query: string;
  from: number;
  items: MentionItem[];
  select: ((item: MentionItem) => void) | null;
};

export type MentionStorage = {
  source: MentionSource | null;
  state: MentionSuggestionState;
  listeners: Set<() => void>;
  keyDown: ((event: KeyboardEvent) => boolean) | null;
};

export const mentionSuggestionKey = new PluginKey("mentionSuggestion");

const inactiveState: MentionSuggestionState = {
  active: false,
  loading: false,
  query: "",
  from: 0,
  items: [],
  select: null,
};

export function getMentionStorage(editor: Editor): MentionStorage | undefined {
  return (editor.storage as unknown as Record<string, unknown>).mention as
    | MentionStorage
    | undefined;
}

function setMentionState(editor: Editor, state: MentionSuggestionState) {
  const storage = getMentionStorage(editor);
  if (!storage) {
    return;
  }

  storage.state = state;
  storage.listeners.forEach(listener => listener());
}

function mentionText(attrs: MentionNodeAttrs): string {
  return `@${attrs.label ?? attrs.id ?? ""}`;
}

function activeState(props: SuggestionProps<MentionItem, MentionNodeAttrs>): MentionSuggestionState {
  return {
    active: true,
    loading: props.loading,
    query: props.query,
    from: props.range.from,
    items: props.items,
    select: item => props.command({ id: item.id, label: item.label }),
  };
}

export const MentionExtension = Mention.extend<
  MentionOptions<MentionItem, MentionNodeAttrs>,
  MentionStorage
>({
  addStorage() {
    return {
      source: null,
      state: inactiveState,
      listeners: new Set(),
      keyDown: null,
    };
  },
}).configure({
  HTMLAttributes: {
    class: "mention",
  },
  deleteTriggerWithBackspace: true,
  renderText: ({ node }) => mentionText(node.attrs as MentionNodeAttrs),
  renderHTML: ({ options, node }) => {
    const attrs = node.attrs as MentionNodeAttrs;

    return [
      "span",
      {
        ...options.HTMLAttributes,
        "data-type": "mention",
        "data-id": attrs.id,
        "data-label": attrs.label,
      },
      mentionText(attrs),
    ];
  },
  suggestion: {
    char: "@",
    pluginKey: mentionSuggestionKey,
    decorationClass: "mention-suggestion",
    debounce: 200,
    allowSpaces: true,
    allow: ({ editor }) => editor.isEditable && getMentionStorage(editor)?.source != null,
    items: async ({ editor, query, signal }) => {
      const source = getMentionStorage(editor)?.source;
      if (!source) {
        return [];
      }

      try {
        return await source(query, signal);
      } catch (error) {
        if (signal.aborted) {
          return [];
        }

        console.error("Cannot load mention suggestions", error);

        return [];
      }
    },
    render: () => {
      let current: Editor | null = null;

      return {
        onStart: (props) => {
          current = props.editor;
          setMentionState(props.editor, activeState(props));
        },
        onUpdate: (props) => {
          setMentionState(props.editor, activeState(props));
        },
        onExit: (props) => {
          current = null;
          setMentionState(props.editor, inactiveState);
        },
        onKeyDown: ({ event }) => {
          if (!current || event.key === "Escape") {
            return false;
          }

          return getMentionStorage(current)?.keyDown?.(event) ?? false;
        },
      };
    },
  },
});
