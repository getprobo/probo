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

type JSONNode = {
  type?: string;
  text?: string;
  attrs?: Record<string, unknown>;
  content?: JSONNode[];
  marks?: unknown;
};

const taskMarker = /^\[([ xX])\]\s?/;

function isNode(value: unknown): value is JSONNode {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

// promoteCheckboxLists turns bullet and ordered items whose text still starts
// with GFM checkbox syntax into task lists. Content saved before task lists
// existed stored "- [ ]" as a plain bullet.
export function promoteCheckboxLists<T>(value: T): T {
  if (!isNode(value)) {
    return value;
  }

  const [promoted] = normalizeNode(value);
  return (promoted ?? value) as T;
}

function normalizeNode(node: JSONNode): JSONNode[] {
  if (!isNode(node)) {
    return [];
  }

  const content = node.content?.flatMap(child => normalizeNode(child));
  const next = content ? { ...node, content } : node;
  if (next.type === "bulletList" || next.type === "orderedList") {
    return splitCheckboxList(next);
  }

  return [next];
}

function splitCheckboxList(list: JSONNode): JSONNode[] {
  const ordered = list.type === "orderedList";
  const start = ordered ? listStart(list) : 1;
  const runs: { task: boolean; startIndex: number; items: JSONNode[] }[] = [];

  for (const [index, item] of (list.content ?? []).entries()) {
    const taskItem = checkboxItem(item);
    const task = taskItem != null;
    const last = runs[runs.length - 1];
    if (!last || last.task !== task) {
      runs.push({ task, startIndex: index, items: [] });
    }

    const current = runs[runs.length - 1];
    if (current) {
      current.items.push(taskItem ?? item);
    }
  }

  if (runs.length === 0 || (runs.length === 1 && !runs[0]?.task)) {
    return [list];
  }

  return runs.map((run) => {
    if (run.task) {
      return { type: "taskList", content: run.items };
    }

    if (!ordered) {
      return { type: "bulletList", content: run.items };
    }

    return {
      type: "orderedList",
      attrs: { ...list.attrs, start: start + run.startIndex },
      content: run.items,
    };
  });
}

function listStart(list: JSONNode): number {
  const start = list.attrs?.start;
  return typeof start === "number" && start >= 1 ? start : 1;
}

function checkboxItem(item: JSONNode): JSONNode | null {
  if (item.type !== "listItem") {
    return null;
  }

  const [first, ...rest] = item.content ?? [];
  if (!first || first.type !== "paragraph") {
    return null;
  }

  const text = first.content?.[0];
  if (!text || text.type !== "text" || !text.text) {
    return null;
  }

  const match = text.text.match(taskMarker);
  if (!match || match.index !== 0) {
    return null;
  }

  const checked = match[1] === "x" || match[1] === "X";
  const restText = text.text.slice(match[0].length);
  const inline = first.content ? [...first.content] : [];
  if (restText) {
    inline[0] = { ...text, text: restText };
  } else {
    inline.shift();
  }

  return {
    type: "taskItem",
    attrs: { checked },
    content: [
      { ...first, content: inline },
      ...rest,
    ],
  };
}
