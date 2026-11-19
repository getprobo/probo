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

import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router";

import {
  createTaskCommentDraftSession,
  readTaskCommentDraft,
  type TaskCommentDraftStorage,
} from "./taskCommentDraft";

function browserLocalStorage(): TaskCommentDraftStorage | null {
  try {
    if (typeof window === "undefined") {
      return null;
    }

    return window.localStorage;
  } catch {
    return null;
  }
}

function initialContent(
  storage: TaskCommentDraftStorage | null,
  taskId: string | undefined,
) {
  if (storage == null || taskId == null || taskId === "") {
    return "";
  }

  return readTaskCommentDraft(storage, taskId);
}

export function useTaskCommentDraft() {
  const { taskId } = useParams<{ taskId: string }>();
  const storage = useMemo(() => browserLocalStorage(), []);
  const [snapshot, setSnapshot] = useState(() => ({
    content: initialContent(storage, taskId),
    editorKey: 0,
  }));
  const [session] = useState(() => createTaskCommentDraftSession({
    taskId,
    storage,
    content: snapshot.content,
    onUpdate: setSnapshot,
  }));

  useEffect(() => {
    function flush() {
      session.flush();
    }

    function flushWhenHidden() {
      if (document.visibilityState === "hidden") {
        flush();
      }
    }

    // Tab close does not wait for the debounce timer.
    window.addEventListener("pagehide", flush);
    document.addEventListener("visibilitychange", flushWhenHidden);
    return () => {
      window.removeEventListener("pagehide", flush);
      document.removeEventListener("visibilitychange", flushWhenHidden);
      flush();
    };
  }, [session]);

  return {
    content: snapshot.content,
    editorKey: snapshot.editorKey,
    setContent: session.setContent,
    clear: session.clear,
  };
}
