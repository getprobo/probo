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

// A collaboration handle for the editor, plus a close() that tears down the repo
// and its network connection.
export type RepoCollaborationHandle = RichEditorCollaborationHandle & {
  close: () => void;
};

// How long a remote collaborator's cursor lingers after their last update before
// it is pruned, so a caret does not stay behind when a peer goes away without a
// clean disconnect.
const presenceTimeToLiveMs = 30_000;

const initializationTimeoutMs = 35_000;

// The presence payload broadcast over repo ephemeral messages: a caret or
// selection expressed as stable Automerge cursors.
type SelectionPresence = {
  kind: "selection";
  selection: TextSelection;
};

function isSelectionPresence(value: unknown): value is SelectionPresence {
  if (!value || typeof value !== "object") return false;
  const message = value as Record<string, unknown>;
  return message.kind === "selection" && !!message.selection;
}

// connectRepoDocument connects to the automerge-repo collaboration endpoint for a
// document version, finds the (server-seeded) document, and returns a handle the
// editor can drive. Presence rides repo ephemeral messages carrying stable
// cursors, so remote carets stay anchored while other people type.
export async function connectRepoDocument(
  documentVersionID: string,
): Promise<RepoCollaborationHandle> {
  const endpoint = new URL(window.location.origin);
  endpoint.protocol = endpoint.protocol === "https:" ? "wss:" : "ws:";
  endpoint.pathname = [
    "api",
    "console",
    "v1",
    "document-versions",
    encodeURIComponent(documentVersionID),
    "repo",
  ].join("/");

  const network = new WebSocketClientAdapter(endpoint.toString());
  const repo = new Repo({ network: [network] });

  const documentURL = (await deriveAutomergeUrl(
    documentVersionID,
  )) as AutomergeUrl;

  let handle: DocHandle<RichEditorAutomergeDocument>;
  try {
    handle = await repo.find<RichEditorAutomergeDocument>(documentURL, {
      signal: AbortSignal.timeout(initializationTimeoutMs),
    });
  } catch (error) {
    void repo.shutdown();
    throw error;
  }

  const adapter = richEditorPresenceAdapter();
  const collaboration = handle as unknown as RepoCollaborationHandle;

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

    handle.broadcast({
      kind: "selection",
      selection,
    } satisfies SelectionPresence);
  };

  collaboration.onPresence = (listener) => {
    // Presence is decoration state: the editor replaces all remote cursors on
    // each update, so accumulate the latest selection per peer and emit the full
    // set, pruning peers that have gone quiet.
    const latest = new Map<
      string,
      { presence: RichEditorPresence; at: number }
    >();

    const handleEphemeral = (
      payload: DocHandleEphemeralMessagePayload<RichEditorAutomergeDocument>,
    ) => {
      if (!isSelectionPresence(payload.message)) return;

      const document = handle.doc();
      if (!document) return;

      const resolved = pmSelectionFromPresence(
        adapter,
        document,
        payload.message.selection,
      );
      const sender = String(payload.senderId);

      const now = Date.now();
      if (resolved) {
        latest.set(sender, {
          at: now,
          presence: {
            connectionID: sender,
            identityID: sender,
            anchorPosition: resolved.anchorPosition,
            headPosition: resolved.headPosition,
          },
        });
      }

      const presences: RichEditorPresence[] = [];
      for (const [peer, entry] of latest) {
        if (now - entry.at > presenceTimeToLiveMs) {
          latest.delete(peer);
          continue;
        }

        presences.push(entry.presence);
      }

      listener(presences);
    };

    handle.on("ephemeral-message", handleEphemeral);

    return () => handle.off("ephemeral-message", handleEphemeral);
  };

  collaboration.close = () => {
    void repo.shutdown();
  };

  return collaboration;
}
