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

import type { RichEditorPresence } from "@probo/ui";
import { describe, expect, it } from "vitest";

import {
  groupPresencesByIdentity,
  type PresenceEntry,
  pruneStalePresences,
} from "./documentPresence";

function presence(overrides: Partial<RichEditorPresence> = {}): RichEditorPresence {
  return {
    connectionID: "conn-1",
    identityID: "idnt_ada",
    displayName: "Ada",
    active: true,
    anchorPosition: 1,
    headPosition: 1,
    ...overrides,
  };
}

describe("pruneStalePresences", () => {
  it("keeps an inactive peer within the TTL and drops a peer past it", () => {
    const now = 100_000;
    const entries: PresenceEntry[] = [
      {
        at: now - 10_000,
        presence: presence({
          connectionID: "idle",
          identityID: "idnt_idle",
          displayName: "Idle",
          active: false,
        }),
      },
      {
        at: now - 45_000,
        presence: presence({
          connectionID: "gone",
          identityID: "idnt_gone",
          displayName: "Gone",
          active: true,
        }),
      },
    ];

    const kept = pruneStalePresences(entries, now, 30_000);

    expect(kept).toHaveLength(1);
    expect(kept[0]?.identityID).toBe("idnt_idle");
    expect(kept[0]?.active).toBe(false);
  });
});

describe("groupPresencesByIdentity", () => {
  it("collapses connections for one person and prefers active state", () => {
    const people = groupPresencesByIdentity([
      presence({
        connectionID: "tab-a",
        identityID: "idnt_ada",
        displayName: "Ada",
        active: false,
        avatarURL: "https://example.com/a.png",
      }),
      presence({
        connectionID: "tab-b",
        identityID: "idnt_ada",
        displayName: "Ada Lovelace",
        active: true,
      }),
      presence({
        connectionID: "tab-c",
        identityID: "idnt_bob",
        displayName: "Bob",
        active: false,
      }),
    ]);

    expect(people).toHaveLength(2);
    expect(people[0]).toMatchObject({
      identityID: "idnt_ada",
      displayName: "Ada Lovelace",
      avatarURL: "https://example.com/a.png",
      active: true,
    });
    expect(people[1]).toMatchObject({
      identityID: "idnt_bob",
      active: false,
    });
  });
});
