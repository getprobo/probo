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

import type { Editor } from "@tiptap/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { refocusEditor } from "./refocusEditor";

class FakeElement {
  isConnected: boolean;
  private readonly popup: FakeElement | null;

  constructor({ isConnected = true, popup = null }: { isConnected?: boolean; popup?: FakeElement | null } = {}) {
    this.isConnected = isConnected;
    this.popup = popup;
  }

  closest() {
    return this.popup;
  }
}

function editorWith({ isDestroyed = false, isEditable = true, hasFocus = false } = {}) {
  const focus = vi.fn();
  const editor = {
    isDestroyed,
    isEditable,
    view: { hasFocus: () => hasFocus },
    commands: { focus },
  } as unknown as Editor;
  return { editor, focus };
}

function focusOn(active: FakeElement | "body") {
  const body = new FakeElement();
  vi.stubGlobal("HTMLElement", FakeElement);
  vi.stubGlobal("document", { activeElement: active === "body" ? body : active, body });
}

describe("refocusEditor", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it.each([
    ["destroyed", { isDestroyed: true }],
    ["not editable", { isEditable: false }],
    ["already focused", { hasFocus: true }],
  ])("leaves an editor that is %s alone", (_name, options) => {
    focusOn("body");
    const { editor, focus } = editorWith(options);

    expect(refocusEditor(editor)()).toBe(false);
    expect(focus).not.toHaveBeenCalled();
  });

  it.each([
    ["focus fell to the body", "body" as const],
    ["focus stayed in a toolbar popup", new FakeElement({ popup: new FakeElement() })],
    ["the focused element was removed", new FakeElement({ isConnected: false })],
  ])("focuses the editor when %s", (_name, active) => {
    focusOn(active);
    const { editor, focus } = editorWith();

    expect(refocusEditor(editor)()).toBe(false);
    expect(focus).toHaveBeenCalledOnce();
  });

  it("keeps focus that moved outside the editor", () => {
    focusOn(new FakeElement());
    const { editor, focus } = editorWith();

    expect(refocusEditor(editor)()).toBe(false);
    expect(focus).not.toHaveBeenCalled();
  });
});
