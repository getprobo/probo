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

import { isRichEditorContentEmpty } from "#/pages/organizations/_lib/richEditorContent";

const taskCommentDraftKeyPrefix = "probo.console.task-comment-draft.v1.";

export const taskCommentDraftDelayMs = 300;

export type TaskCommentDraftStorage = Pick<Storage, "getItem" | "setItem" | "removeItem">;

export type TaskCommentDraftState = {
  content: string;
  editorKey: number;
};

export type TaskCommentDraftClock = {
  setTimeout: (callback: () => void, delayMs: number) => number;
  clearTimeout: (id: number) => void;
};

type JSONNode = Record<string, unknown> & {
  type?: string;
  content?: JSONNode[];
};

export function taskCommentDraftKey(taskId: string) {
  return taskCommentDraftKeyPrefix + taskId;
}

export function readTaskCommentDraft(
  storage: TaskCommentDraftStorage,
  taskId: string,
): string {
  if (taskId === "") {
    return "";
  }

  const key = taskCommentDraftKey(taskId);
  let raw: string | null;
  try {
    raw = storage.getItem(key);
  } catch {
    return "";
  }

  if (raw == null || raw === "") {
    return "";
  }

  const content = normalizeTaskCommentDraft(raw);
  if (content === "") {
    forgetTaskCommentDraft(storage, key);
  }

  return content;
}

export function writeTaskCommentDraft(
  storage: TaskCommentDraftStorage,
  taskId: string,
  content: string,
) {
  if (taskId === "") {
    return;
  }

  const key = taskCommentDraftKey(taskId);
  const next = normalizeTaskCommentDraft(content);
  try {
    if (next === "") {
      storage.removeItem(key);
      return;
    }

    storage.setItem(key, next);
  } catch {
    // Private mode and quota exhaustion are best effort.
  }
}

export function createTaskCommentDraftSession(options: {
  taskId: string | undefined;
  storage: TaskCommentDraftStorage | null;
  content: string;
  delayMs?: number;
  clock?: TaskCommentDraftClock;
  onUpdate: (state: TaskCommentDraftState) => void;
}) {
  const taskId = options.taskId;
  const storage = options.storage;
  const delayMs = options.delayMs ?? taskCommentDraftDelayMs;
  const clock = options.clock ?? browserClock();
  let content = options.content;
  let editorKey = 0;
  let timer: number | null = null;

  function cancelTimer() {
    if (timer == null) {
      return;
    }

    clock.clearTimeout(timer);
    timer = null;
  }

  function writeCurrent() {
    if (taskId == null || taskId === "" || storage == null) {
      return;
    }

    writeTaskCommentDraft(storage, taskId, content);
  }

  function writeNow() {
    cancelTimer();
    writeCurrent();
  }

  function publish() {
    options.onUpdate({ content, editorKey });
  }

  return {
    setContent(this: void, next: string) {
      content = next;
      publish();
      if (normalizeTaskCommentDraft(next) === "") {
        writeNow();
        return;
      }

      cancelTimer();
      timer = clock.setTimeout(() => {
        timer = null;
        writeCurrent();
      }, delayMs);
    },
    clear(this: void) {
      content = "";
      editorKey += 1;
      writeNow();
      publish();
    },
    flush(this: void) {
      writeNow();
    },
  };
}

function browserClock(): TaskCommentDraftClock {
  return {
    setTimeout(callback, delayMs) {
      return window.setTimeout(callback, delayMs);
    },
    clearTimeout(id) {
      window.clearTimeout(id);
    },
  };
}

function forgetTaskCommentDraft(storage: TaskCommentDraftStorage, key: string) {
  try {
    storage.removeItem(key);
  } catch {
    // The bad value is ignored on the next read as well.
  }
}

function normalizeTaskCommentDraft(content: string): string {
  let parsed: unknown;
  try {
    parsed = JSON.parse(content);
  } catch {
    return "";
  }

  if (!isJSONNode(parsed) || parsed.type !== "doc") {
    return "";
  }

  const next = JSON.stringify(withoutContentUploads(parsed));
  if (isRichEditorContentEmpty(next)) {
    return "";
  }

  return next;
}

function withoutContentUploads(node: JSONNode): JSONNode {
  if (!Array.isArray(node.content)) {
    return node;
  }

  const content = node.content
    .filter(child => child.type !== "contentUpload")
    .map(child => withoutContentUploads(child));

  return { ...node, content };
}

function isJSONNode(value: unknown): value is JSONNode {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    return false;
  }

  if (!("content" in value) || value.content === undefined) {
    return true;
  }

  return Array.isArray(value.content) && value.content.every(isJSONNode);
}
