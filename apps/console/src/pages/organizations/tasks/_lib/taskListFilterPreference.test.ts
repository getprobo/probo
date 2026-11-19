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

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  applyTaskListFilterPreference,
  defaultTaskListFilterPreference,
  readTaskListFilterPreference,
  restoredTaskListFilters,
  taskListFilterPreferenceFromParams,
  writeTaskListFilterPreference,
} from "./taskListFilterPreference";
import { unassignedAssignee } from "./taskPath";

const organizationId = "org-1";
const profileId = "profile-1";

function memoryStorage(): Storage {
  const values = new Map<string, string>();
  return {
    get length() {
      return values.size;
    },
    clear() {
      values.clear();
    },
    getItem(key: string) {
      return values.get(key) ?? null;
    },
    key(index: number) {
      return [...values.keys()][index] ?? null;
    },
    removeItem(key: string) {
      values.delete(key);
    },
    setItem(key: string, value: string) {
      values.set(key, value);
    },
  };
}

describe("defaultTaskListFilterPreference", () => {
  it("selects the current profile", () => {
    expect(defaultTaskListFilterPreference(profileId)).toEqual({
      status: null,
      assignee: profileId,
      due: null,
    });
  });

  it("returns null without a profile", () => {
    expect(defaultTaskListFilterPreference("")).toBeNull();
  });
});

describe("restoredTaskListFilters", () => {
  it("defaults a bare URL to the current user's tasks and asks to remember it", () => {
    const params = new URLSearchParams("q=audit");
    const restored = restoredTaskListFilters(params, null, profileId);

    expect(restored?.remember).toEqual({
      status: null,
      assignee: profileId,
      due: null,
    });
    expect(restored?.params.get("assignee")).toBe(profileId);
    expect(restored?.params.get("q")).toBe("audit");
  });

  it("leaves a URL that already names a filter", () => {
    const params = new URLSearchParams("status=done");

    expect(restoredTaskListFilters(params, null, profileId)).toBeNull();
  });

  it("applies a saved view without asking to remember it again", () => {
    const stored = {
      status: "in-progress",
      assignee: unassignedAssignee,
      due: "overdue",
    };
    const restored = restoredTaskListFilters(new URLSearchParams(), stored, profileId);

    expect(restored?.remember).toBeNull();
    expect(restored?.params.get("status")).toBe("in-progress");
    expect(restored?.params.get("assignee")).toBe(unassignedAssignee);
    expect(restored?.params.get("due")).toBe("overdue");
  });

  it("keeps a saved all-tasks view from adding the default assignee", () => {
    const stored = { status: null, assignee: null, due: null };
    const restored = restoredTaskListFilters(new URLSearchParams("q=audit"), stored, profileId);

    expect(restored?.params.toString()).toBe(new URLSearchParams("q=audit").toString());
    expect(restored?.remember).toBeNull();
  });

  it("returns null when nothing is saved and there is no profile", () => {
    expect(restoredTaskListFilters(new URLSearchParams(), null, "")).toBeNull();
  });
});

describe("task list filter storage", () => {
  beforeEach(() => {
    vi.stubGlobal("localStorage", memoryStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("round-trips a preference for one organization and profile", () => {
    const preference = {
      status: "todo",
      assignee: profileId,
      due: "1w",
    };

    writeTaskListFilterPreference(organizationId, profileId, preference);

    expect(readTaskListFilterPreference(organizationId, profileId)).toEqual(preference);
    expect(readTaskListFilterPreference(organizationId, "profile-2")).toBeNull();
    expect(readTaskListFilterPreference("org-2", profileId)).toBeNull();
  });

  it("keeps an explicit empty preference distinct from a missing one", () => {
    writeTaskListFilterPreference(organizationId, profileId, {
      status: null,
      assignee: null,
      due: null,
    });

    expect(readTaskListFilterPreference(organizationId, profileId)).toEqual({
      status: null,
      assignee: null,
      due: null,
    });
  });

  it("drops invalid saved fields and ignores corrupt JSON", () => {
    localStorage.setItem(
      `probo.console.tasks.filters.v1.${organizationId}.${profileId}`,
      JSON.stringify({ status: "nope", assignee: profileId, due: 3 }),
    );
    localStorage.setItem(
      `probo.console.tasks.filters.v1.${organizationId}.profile-2`,
      "{",
    );

    expect(readTaskListFilterPreference(organizationId, profileId)).toEqual({
      status: null,
      assignee: profileId,
      due: null,
    });
    expect(readTaskListFilterPreference(organizationId, "profile-2")).toBeNull();
  });

  it("ignores storage failures", () => {
    vi.stubGlobal("localStorage", {
      getItem: () => {
        throw new Error("blocked");
      },
      setItem: () => {
        throw new Error("blocked");
      },
    });

    expect(() => {
      writeTaskListFilterPreference(organizationId, profileId, {
        status: "todo",
        assignee: profileId,
        due: null,
      });
    }).not.toThrow();
    expect(readTaskListFilterPreference(organizationId, profileId)).toBeNull();
  });
});

describe("taskListFilterPreferenceFromParams", () => {
  it("reads the filter params and drops values the list does not understand", () => {
    const params = new URLSearchParams();
    params.set("status", "in-progress");
    params.set("assignee", profileId);
    params.set("due", "soon");
    params.set("q", "audit");

    expect(taskListFilterPreferenceFromParams(params)).toEqual({
      status: "in-progress",
      assignee: profileId,
      due: null,
    });
  });
});

describe("applyTaskListFilterPreference", () => {
  it("removes a cleared filter without touching the search query", () => {
    const params = new URLSearchParams("status=todo&q=audit");
    const next = applyTaskListFilterPreference(params, {
      status: null,
      assignee: profileId,
      due: null,
    });

    expect(next.get("status")).toBeNull();
    expect(next.get("assignee")).toBe(profileId);
    expect(next.get("q")).toBe("audit");
    expect(params.get("status")).toBe("todo");
  });
});
