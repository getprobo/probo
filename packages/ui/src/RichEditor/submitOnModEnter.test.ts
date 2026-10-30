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

import { submitOnModEnter } from "./submitOnModEnter";

function press(overrides: Partial<{
  key: string;
  repeat: boolean;
  altKey: boolean;
  shiftKey: boolean;
  metaKey: boolean;
  ctrlKey: boolean;
}> = {}) {
  let prevented = false;
  const event = {
    key: "Enter",
    repeat: false,
    altKey: false,
    shiftKey: false,
    metaKey: false,
    ctrlKey: false,
    preventDefault: () => {
      prevented = true;
    },
    ...overrides,
  };
  return {
    event,
    prevented: () => prevented,
  };
}

describe("submitOnModEnter", () => {
  it("submits on Command-Enter when the platform uses the meta key", () => {
    let submitted = 0;
    const { event, prevented } = press({ metaKey: true });

    const handled = submitOnModEnter(event, () => {
      submitted += 1;
    }, true);

    expect(handled).toBe(true);
    expect(submitted).toBe(1);
    expect(prevented()).toBe(true);
  });

  it("submits on Control-Enter when the platform uses the control key", () => {
    let submitted = 0;
    const { event, prevented } = press({ ctrlKey: true });

    const handled = submitOnModEnter(event, () => {
      submitted += 1;
    }, false);

    expect(handled).toBe(true);
    expect(submitted).toBe(1);
    expect(prevented()).toBe(true);
  });

  it("ignores the other modifier so the platform shortcut stays unambiguous", () => {
    let submitted = 0;
    const mac = press({ ctrlKey: true });
    const other = press({ metaKey: true });

    expect(submitOnModEnter(mac.event, () => {
      submitted += 1;
    }, true)).toBe(false);
    expect(submitOnModEnter(other.event, () => {
      submitted += 1;
    }, false)).toBe(false);
    expect(submitted).toBe(0);
    expect(mac.prevented()).toBe(false);
    expect(other.prevented()).toBe(false);
  });

  it("leaves Shift-Enter and plain Enter for the editor", () => {
    let submitted = 0;
    const shifted = press({ metaKey: true, shiftKey: true });
    const plain = press();
    const submit = () => {
      submitted += 1;
    };

    expect(submitOnModEnter(shifted.event, submit, true)).toBe(false);
    expect(submitOnModEnter(plain.event, submit, true)).toBe(false);
    expect(submitted).toBe(0);
    expect(shifted.prevented()).toBe(false);
    expect(plain.prevented()).toBe(false);
  });

  it("swallows a held Command-Enter without sending again", () => {
    let submitted = 0;
    const { event, prevented } = press({ metaKey: true, repeat: true });

    const handled = submitOnModEnter(event, () => {
      submitted += 1;
    }, true);

    expect(handled).toBe(true);
    expect(submitted).toBe(0);
    expect(prevented()).toBe(true);
  });

  it("does nothing when no composer is listening", () => {
    const { event, prevented } = press({ metaKey: true });

    expect(submitOnModEnter(event, undefined, true)).toBe(false);
    expect(prevented()).toBe(false);
  });
});
