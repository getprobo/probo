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

export type PresenceEntry = {
  presence: RichEditorPresence;
  at: number;
};

export type DocumentPresencePerson = {
  identityID: string;
  displayName?: string;
  avatarURL?: string;
  active: boolean;
};

// Drop peers that have gone quiet past the TTL. A heartbeat refreshes `at`, so
// an idle-but-connected person stays listed until they leave.
export function pruneStalePresences(
  entries: Iterable<PresenceEntry>,
  now: number,
  timeToLiveMs: number,
): RichEditorPresence[] {
  const presences: RichEditorPresence[] = [];
  for (const entry of entries) {
    if (now - entry.at > timeToLiveMs) continue;
    presences.push(entry.presence);
  }
  return presences;
}

// Collapse one-or-more connections for the same person into a single facepile
// entry. Any active connection keeps the person active; name and avatar prefer
// an active connection when both exist.
export function groupPresencesByIdentity(
  presences: RichEditorPresence[],
): DocumentPresencePerson[] {
  const byIdentity = new Map<string, DocumentPresencePerson>();

  for (const presence of presences) {
    const existing = byIdentity.get(presence.identityID);
    if (!existing) {
      byIdentity.set(presence.identityID, {
        identityID: presence.identityID,
        displayName: presence.displayName,
        avatarURL: presence.avatarURL,
        active: presence.active,
      });
      continue;
    }

    if (presence.active) {
      existing.active = true;
    }

    if (
      presence.displayName
      && (presence.active || !existing.displayName)
    ) {
      existing.displayName = presence.displayName;
    }

    if (
      presence.avatarURL
      && (presence.active || !existing.avatarURL)
    ) {
      existing.avatarURL = presence.avatarURL;
    }
  }

  return [...byIdentity.values()].sort((left, right) => {
    if (left.active !== right.active) return left.active ? -1 : 1;
    return presenceSortName(left).localeCompare(presenceSortName(right));
  });
}

function presenceSortName(person: DocumentPresencePerson): string {
  return (person.displayName?.trim() || person.identityID).toLocaleLowerCase();
}
