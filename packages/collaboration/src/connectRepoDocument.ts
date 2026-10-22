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

import {
  type AutomergeUrl,
  type DocHandle,
  type DocHandleEphemeralMessagePayload,
  type DocumentProgress,
  Repo,
} from "@automerge/automerge-repo";
import { WebSocketClientAdapter } from "@automerge/automerge-repo-network-websocket";
import {
  deriveAutomergeUrl,
  pmSelectionFromPresence,
  presenceFromPmSelection,
  type RichEditorAutomergeDocument,
  type RichEditorCollaborationHandle,
  type RichEditorPresence,
  richEditorPresenceAdapter,
  type TextSelection,
} from "@probo/ui";

import { type PresenceEntry, pruneStalePresences } from "./presence";

// A collaboration handle for the editor, plus a close() that tears down the repo
// and its network connection.
export type RepoCollaborationHandle = RichEditorCollaborationHandle & {
  close: () => void;
};

// How long a remote collaborator lingers after their last ephemeral before they
// are pruned (three missed 10s heartbeats).
export const presenceTimeToLiveMs = 30_000;

// Heartbeat interval so idle peers stay listed while the tab is open.
const presenceHeartbeatIntervalMs = 10_000;

// Idle threshold: no caret/selection activity and the person is treated as
// inactive (still connected, no caret).
const presenceIdleMs = 60_000;

const initializationTimeoutMs = 35_000;

// The presence payload broadcast over repo ephemeral messages: a caret or
// selection expressed as stable Automerge cursors. displayName, identityID, and
// avatarURL are stamped by the gateway from the authenticated connection.
type SelectionPresence = {
  kind: "selection";
  selection: TextSelection;
  active: boolean;
  displayName?: string;
  identityID?: string;
  avatarURL?: string;
};

function isCursor(value: unknown): value is string {
  return typeof value === "string" && value !== "";
}

function isTextSelection(value: unknown): value is TextSelection {
  if (!value || typeof value !== "object") return false;
  const selection = value as Record<string, unknown>;
  return typeof selection.field === "string"
    && selection.field !== ""
    && isCursor(selection.anchor)
    && isCursor(selection.head);
}

function isSelectionPresence(value: unknown): value is SelectionPresence {
  if (!value || typeof value !== "object") return false;
  const message = value as Record<string, unknown>;
  return message.kind === "selection"
    && typeof message.active === "boolean"
    && isTextSelection(message.selection);
}

function isDeparturePresence(value: unknown): boolean {
  if (!value || typeof value !== "object") return false;
  return (value as { kind?: unknown }).kind === "departure";
}

// For a plain automerge: URL (no heads or path), findWithProgress returns the
// DocumentQuery itself, which exposes the handle before the document is ready.
// The public DocumentProgress type omits it.
type DocumentProgressWithHandle<T> = DocumentProgress<T> & {
  handle?: DocHandle<T>;
};

// connectRepoDocument connects to the automerge-repo collaboration endpoint for a
// document, finds the (server-seeded) document, and returns a handle the
// editor can drive. Presence rides repo ephemeral messages carrying stable
// cursors, so remote carets stay anchored while other people type.
export async function connectRepoDocument(
  documentId: string,
): Promise<RepoCollaborationHandle> {
  const endpoint = new URL(window.location.origin);
  endpoint.protocol = endpoint.protocol === "https:" ? "wss:" : "ws:";
  endpoint.pathname = [
    "api",
    "console",
    "v1",
    "documents",
    encodeURIComponent(documentId),
    "collaboration",
  ].join("/");

  const network = new WebSocketClientAdapter(endpoint.toString());
  const repo = new Repo({ network: [network] });

  const documentURL = (await deriveAutomergeUrl(documentId)) as AutomergeUrl;

  const progress = repo.findWithProgress<RichEditorAutomergeDocument>(
    documentURL,
  ) as DocumentProgressWithHandle<RichEditorAutomergeDocument>;
  const handle = progress.handle;
  if (!handle) {
    void repo.shutdown();
    throw new Error("automerge-repo did not expose the document handle before sync");
  }

  const adapter = richEditorPresenceAdapter();
  const collaboration = handle as unknown as RepoCollaborationHandle;

  let lastSelection: TextSelection | undefined;
  let lastActivityAt = Date.now();
  let lastBroadcastActive: boolean | undefined;
  let documentReady = false;
  const pendingEphemerals: DocHandleEphemeralMessagePayload<
    RichEditorAutomergeDocument
  >[] = [];
  const latest = new Map<string, PresenceEntry>();
  const listeners = new Set<(presences: RichEditorPresence[]) => void>();

  const localActive = (): boolean => {
    if (typeof document !== "undefined" && document.visibilityState === "hidden") {
      return false;
    }
    return Date.now() - lastActivityAt < presenceIdleMs;
  };

  const broadcastSelection = (selection: TextSelection, active: boolean) => {
    handle.broadcast({
      kind: "selection",
      selection,
      active,
    } satisfies SelectionPresence);
    lastBroadcastActive = active;
  };

  const publishPresence = () => {
    if (!lastSelection) return;
    const active = localActive();
    broadcastSelection(lastSelection, active);
  };

  collaboration.updatePresence = (anchorPosition, headPosition) => {
    const document = handle.doc();
    if (!document) return;

    const selection = presenceFromPmSelection(
      adapter,
      document,
      anchorPosition,
      headPosition,
    );
    if (!selection) return;

    lastSelection = selection;
    lastActivityAt = Date.now();
    publishPresence();
  };

  const emitPresences = () => {
    const now = Date.now();
    for (const [peer, entry] of latest) {
      if (now - entry.at > presenceTimeToLiveMs) {
        latest.delete(peer);
      }
    }
    const presences = pruneStalePresences(
      latest.values(),
      now,
      presenceTimeToLiveMs,
    );
    for (const listener of listeners) {
      listener(presences);
    }
  };

  const applyEphemeral = (
    payload: DocHandleEphemeralMessagePayload<RichEditorAutomergeDocument>,
  ) => {
    const sender = String(payload.senderId);
    if (isDeparturePresence(payload.message)) {
      if (!latest.delete(sender)) return;
      emitPresences();
      return;
    }

    if (!isSelectionPresence(payload.message)) return;

    const document = handle.doc();
    if (!document) return;

    // A peer cursor can throw inside Automerge. Keep the ephemeral handler alive.
    let resolved: ReturnType<typeof pmSelectionFromPresence>;
    try {
      resolved = pmSelectionFromPresence(
        adapter,
        document,
        payload.message.selection,
      );
    } catch {
      resolved = null;
    }
    const now = Date.now();
    const displayName = typeof payload.message.displayName === "string"
      ? payload.message.displayName.trim()
      : undefined;
    const identityID = typeof payload.message.identityID === "string"
      && payload.message.identityID !== ""
      ? payload.message.identityID
      : sender;
    const avatarURL = typeof payload.message.avatarURL === "string"
      ? payload.message.avatarURL.trim()
      : undefined;
    const active = payload.message.active;

    const previous = latest.get(sender);
    const isNewSender = !previous;
    if (resolved) {
      latest.set(sender, {
        at: now,
        presence: {
          connectionID: sender,
          identityID,
          displayName: displayName || undefined,
          avatarURL: avatarURL || undefined,
          active,
          anchorPosition: resolved.anchorPosition,
          headPosition: resolved.headPosition,
        },
      });
    } else if (previous) {
      // Heartbeat / selection that no longer resolves: keep the last caret but
      // refresh liveness and identity fields so the peer is not pruned.
      latest.set(sender, {
        at: now,
        presence: {
          ...previous.presence,
          identityID,
          displayName: displayName || previous.presence.displayName,
          avatarURL: avatarURL || previous.presence.avatarURL,
          active,
        },
      });
    } else {
      return;
    }

    emitPresences();

    // Answer a peer we have not seen yet so they do not wait for the idle
    // heartbeat — including people connected through another server process.
    if (isNewSender) {
      publishPresence();
    }
  };

  const handleEphemeral = (
    payload: DocHandleEphemeralMessagePayload<RichEditorAutomergeDocument>,
  ) => {
    if (!isSelectionPresence(payload.message) && !isDeparturePresence(payload.message)) {
      return;
    }

    // Join snapshots can arrive while the CRDT is still syncing; hold them
    // until offsets can be resolved against a ready document.
    if (!documentReady) {
      pendingEphemerals.push(payload);
      return;
    }

    applyEphemeral(payload);
  };

  handle.on("ephemeral-message", handleEphemeral);

  try {
    await progress.whenReady({
      signal: AbortSignal.timeout(initializationTimeoutMs),
    });
  } catch (error) {
    handle.off("ephemeral-message", handleEphemeral);
    void repo.shutdown();
    throw error;
  }

  documentReady = true;
  for (const payload of pendingEphemerals) {
    applyEphemeral(payload);
  }
  pendingEphemerals.length = 0;

  collaboration.onPresence = (listener) => {
    listeners.add(listener);
    listener(
      pruneStalePresences(latest.values(), Date.now(), presenceTimeToLiveMs),
    );
    return () => {
      listeners.delete(listener);
    };
  };

  const heartbeat = window.setInterval(() => {
    if (!lastSelection) return;
    const active = localActive();
    broadcastSelection(lastSelection, active);
  }, presenceHeartbeatIntervalMs);

  // Drop gone peers even when nobody else is broadcasting.
  const prune = window.setInterval(() => {
    if (latest.size === 0) return;
    emitPresences();
  }, presenceHeartbeatIntervalMs);

  const onVisibilityChange = () => {
    if (!lastSelection) return;
    const active = localActive();
    if (active === lastBroadcastActive) return;
    broadcastSelection(lastSelection, active);
  };
  document.addEventListener("visibilitychange", onVisibilityChange);

  collaboration.close = () => {
    window.clearInterval(heartbeat);
    window.clearInterval(prune);
    document.removeEventListener("visibilitychange", onVisibilityChange);
    handle.off("ephemeral-message", handleEphemeral);
    listeners.clear();
    latest.clear();
    pendingEphemerals.length = 0;
    void repo.shutdown();
  };

  return collaboration;
}
