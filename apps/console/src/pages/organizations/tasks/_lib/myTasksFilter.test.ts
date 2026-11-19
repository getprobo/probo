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

import { myTasksSearchParams } from "./myTasksFilter";

const profileId = "profile-1";

describe("myTasksSearchParams", () => {
  it("assigns the current user when the URL has no filters", () => {
    const next = myTasksSearchParams(new URLSearchParams(), profileId);

    expect(next?.toString()).toBe(new URLSearchParams({ assignee: profileId }).toString());
  });

  it("keeps unrelated params while adding the current user", () => {
    const next = myTasksSearchParams(new URLSearchParams("from=nav"), profileId);

    expect(next?.get("assignee")).toBe(profileId);
    expect(next?.get("from")).toBe("nav");
  });

  it("leaves a URL that already sets a filter", () => {
    const urls = [
      "q=audit",
      "status=done",
      "assignee=someone-else",
      "due=overdue",
      "internalControl=control-1",
    ];

    for (const url of urls) {
      expect(myTasksSearchParams(new URLSearchParams(url), profileId)).toBeNull();
    }
  });

  it("does nothing without a profile", () => {
    expect(myTasksSearchParams(new URLSearchParams(), "")).toBeNull();
  });
});
