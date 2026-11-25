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

import { documentVersionState } from "./documentVersionState";

describe("documentVersionState", () => {
  it("is stable while the published and pending versions stay the same", () => {
    expect(documentVersionState("v1", null)).toBe(documentVersionState("v1", null));
  });

  it("changes when a pending version is approved", () => {
    expect(documentVersionState(null, "pending"))
      .not.toBe(documentVersionState("pending", null));
  });

  it("changes when a rejected version is republished under the same number", () => {
    expect(documentVersionState("v1", "rejected"))
      .not.toBe(documentVersionState("v1", "republished"));
    expect(documentVersionState("v1", null))
      .not.toBe(documentVersionState("v1", "republished"));
  });

  it("treats missing versions as empty", () => {
    expect(documentVersionState(undefined, undefined)).toBe(documentVersionState(null, null));
  });
});
