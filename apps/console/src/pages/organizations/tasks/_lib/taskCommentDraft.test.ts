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

import { describe, expect, it } from "vitest";

import {
  createTaskCommentDraftSession,
  readTaskCommentDraft,
  type TaskCommentDraftClock,
  taskCommentDraftDelayMs,
  taskCommentDraftKey,
  type TaskCommentDraftState,
  type TaskCommentDraftStorage,
  writeTaskCommentDraft,
} from "./taskCommentDraft";

const taskId = "task-1";

function memoryStorage(): TaskCommentDraftStorage & { items: Map<string, string> } {
  const items = new Map<string, string>();

  return {
    items,
    getItem(key) {
      return items.get(key) ?? null;
    },
    setItem(key, value) {
      items.set(key, value);
    },
    removeItem(key) {
      items.delete(key);
    },
  };
}

function manualClock() {
  const pending = new Map<number, { callback: () => void; delayMs: number }>();
  let nextId = 1;

  const clock: TaskCommentDraftClock = {
    setTimeout(callback, delayMs) {
      const id = nextId;
      nextId += 1;
      pending.set(id, { callback, delayMs });
      return id;
    },
    clearTimeout(id) {
      pending.delete(id);
    },
  };

  return {
    clock,
    pending,
    fireNext() {
      const next = pending.entries().next();
      if (next.done) {
        return false;
      }

      const [id, timer] = next.value;
      pending.delete(id);
      timer.callback();
      return true;
    },
  };
}

function textDoc(text: string, extras: Record<string, unknown>[] = []) {
  return JSON.stringify({
    type: "doc",
    content: [
      { type: "paragraph", content: [{ type: "text", text }] },
      ...extras,
    ],
  });
}

const evidence = textDoc("See the attached evidence", [
  {
    type: "file",
    attrs: {
      fileId: "file-1",
      fileName: "audit.pdf",
      mimeType: "application/pdf",
      size: 2048,
    },
  },
  {
    type: "image",
    attrs: { fileId: "file-2", alt: "screenshot", src: null },
  },
  {
    type: "paragraph",
    content: [{
      type: "mention",
      attrs: { id: "profile-1", label: "Ada" },
    }],
  },
]);

describe("task comment drafts", () => {
  it("keeps text, mentions, and uploaded evidence for one task", () => {
    const storage = memoryStorage();

    writeTaskCommentDraft(storage, taskId, evidence);

    expect(taskCommentDraftKey(taskId)).toBe(
      "probo.console.task-comment-draft.v1.task-1",
    );
    expect(JSON.parse(readTaskCommentDraft(storage, taskId))).toEqual(JSON.parse(evidence));
    expect(readTaskCommentDraft(storage, "task-2")).toBe("");
  });

  it("drops an in-progress upload and keeps the settled file", () => {
    const storage = memoryStorage();
    const content = textDoc("notes", [
      { type: "contentUpload", attrs: { id: "upload-1", fileName: "pending.pdf" } },
      {
        type: "file",
        attrs: { fileId: "file-9", fileName: "ready.pdf", mimeType: "application/pdf", size: 10 },
      },
    ]);

    writeTaskCommentDraft(storage, taskId, content);

    const restored = JSON.parse(readTaskCommentDraft(storage, taskId)) as {
      content: { type?: string; attrs?: { fileId?: string } }[];
    };
    const types = restored.content.map(node => node.type);

    expect(types).not.toContain("contentUpload");
    expect(restored.content.some(node => node.attrs?.fileId === "file-9")).toBe(true);
  });

  it("keeps a comment that is only attached evidence", () => {
    const storage = memoryStorage();
    const content = JSON.stringify({
      type: "doc",
      content: [{
        type: "image",
        attrs: { fileId: "file-3", alt: "photo", src: null },
      }],
    });

    writeTaskCommentDraft(storage, taskId, content);

    expect(JSON.parse(readTaskCommentDraft(storage, taskId))).toEqual(JSON.parse(content));
  });

  it("removes empty, invalid, and placeholder-only drafts", () => {
    const storage = memoryStorage();
    const key = taskCommentDraftKey(taskId);
    storage.items.set(key, "not-json");

    expect(readTaskCommentDraft(storage, taskId)).toBe("");
    expect(storage.items.has(key)).toBe(false);

    writeTaskCommentDraft(storage, taskId, evidence);
    writeTaskCommentDraft(storage, taskId, JSON.stringify({
      type: "doc",
      content: [{ type: "paragraph" }],
    }));
    expect(storage.items.has(key)).toBe(false);

    writeTaskCommentDraft(storage, taskId, JSON.stringify({
      type: "doc",
      content: [{ type: "contentUpload", attrs: { id: "upload-1", fileName: "a.pdf" } }],
    }));
    expect(readTaskCommentDraft(storage, taskId)).toBe("");

    writeTaskCommentDraft(storage, taskId, JSON.stringify({ type: "paragraph" }));
    expect(storage.items.has(key)).toBe(false);
  });

  it("leaves the previous draft when storage rejects the write", () => {
    const storage = memoryStorage();
    writeTaskCommentDraft(storage, taskId, evidence);
    storage.setItem = () => {
      throw new Error("quota");
    };

    expect(() => {
      writeTaskCommentDraft(storage, taskId, textDoc("replacement"));
    }).not.toThrow();
    expect(JSON.parse(readTaskCommentDraft(storage, taskId))).toEqual(JSON.parse(evidence));
  });

  it("ignores storage reads that throw", () => {
    const storage: TaskCommentDraftStorage = {
      getItem() {
        throw new Error("denied");
      },
      setItem() {
        throw new Error("denied");
      },
      removeItem() {
        throw new Error("denied");
      },
    };

    expect(readTaskCommentDraft(storage, taskId)).toBe("");
    expect(() => {
      writeTaskCommentDraft(storage, taskId, "");
    }).not.toThrow();
  });

  it("writes the latest text after a pause and drops it when cleared", () => {
    const storage = memoryStorage();
    const clock = manualClock();
    const updates: TaskCommentDraftState[] = [];
    const session = createTaskCommentDraftSession({
      taskId,
      storage,
      content: "",
      clock: clock.clock,
      onUpdate: (state) => {
        updates.push(state);
      },
    });

    session.setContent(textDoc("hel"));
    session.setContent(textDoc("hello"));

    expect(storage.items.size).toBe(0);
    expect(clock.pending.size).toBe(1);
    expect([...clock.pending.values()][0]?.delayMs).toBe(taskCommentDraftDelayMs);

    expect(clock.fireNext()).toBe(true);
    expect(JSON.parse(readTaskCommentDraft(storage, taskId))).toEqual(JSON.parse(textDoc("hello")));

    session.setContent(textDoc("hello!"));
    session.clear();

    expect(clock.pending.size).toBe(0);
    expect(readTaskCommentDraft(storage, taskId)).toBe("");
    expect(updates.at(-1)).toEqual({ content: "", editorKey: 1 });
    expect(clock.fireNext()).toBe(false);
  });

  it("flushes a pending draft before the timer and keeps evidence", () => {
    const storage = memoryStorage();
    const clock = manualClock();
    const session = createTaskCommentDraftSession({
      taskId,
      storage,
      content: "",
      clock: clock.clock,
      onUpdate: () => undefined,
    });

    session.setContent(evidence);
    session.flush();

    expect(clock.pending.size).toBe(0);
    expect(JSON.parse(readTaskCommentDraft(storage, taskId))).toEqual(JSON.parse(evidence));
    expect(clock.fireNext()).toBe(false);
  });

  it("removes a saved draft as soon as the editor is cleared", () => {
    const storage = memoryStorage();
    const clock = manualClock();
    writeTaskCommentDraft(storage, taskId, evidence);
    const session = createTaskCommentDraftSession({
      taskId,
      storage,
      content: evidence,
      clock: clock.clock,
      onUpdate: () => undefined,
    });

    session.setContent(JSON.stringify({
      type: "doc",
      content: [{ type: "paragraph" }],
    }));

    expect(clock.pending.size).toBe(0);
    expect(readTaskCommentDraft(storage, taskId)).toBe("");
  });
});
