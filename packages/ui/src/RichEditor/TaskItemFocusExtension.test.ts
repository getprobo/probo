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

import { syncTaskCheckboxFocusability } from "./TaskItemFocusExtension";

const taskCheckboxSelector = "ul[data-type=\"taskList\"] input[type=\"checkbox\"]";

function fakeInput(tabIndex = 0) {
  return { tabIndex } as unknown as HTMLInputElement;
}

function fakeEditorRoot(taskCheckboxCount: number) {
  const taskCheckboxes = Array.from(
    { length: taskCheckboxCount },
    () => fakeInput(),
  );
  const otherInput = fakeInput();

  return {
    root: {
      querySelectorAll: (selector: string) => {
        expect(selector).toBe(taskCheckboxSelector);
        return taskCheckboxes as unknown as NodeListOf<HTMLInputElement>;
      },
    },
    taskCheckboxes,
    otherInput,
  };
}

describe("syncTaskCheckboxFocusability", () => {
  it("removes task checkboxes from the tab order when read-only", () => {
    const { root, taskCheckboxes, otherInput } = fakeEditorRoot(2);

    syncTaskCheckboxFocusability(root, false);

    expect(taskCheckboxes.map(checkbox => checkbox.tabIndex)).toEqual([-1, -1]);
    expect(otherInput.tabIndex).toBe(0);
  });

  it("restores the natural tab order when editable", () => {
    const { root, taskCheckboxes, otherInput } = fakeEditorRoot(2);
    taskCheckboxes.forEach((checkbox) => {
      checkbox.tabIndex = -1;
    });
    otherInput.tabIndex = -1;

    syncTaskCheckboxFocusability(root, true);

    expect(taskCheckboxes.map(checkbox => checkbox.tabIndex)).toEqual([0, 0]);
    expect(otherInput.tabIndex).toBe(-1);
  });

  it("does nothing when there are no task checkboxes", () => {
    const { root, otherInput } = fakeEditorRoot(0);

    syncTaskCheckboxFocusability(root, false);

    expect(otherInput.tabIndex).toBe(0);
  });
});
