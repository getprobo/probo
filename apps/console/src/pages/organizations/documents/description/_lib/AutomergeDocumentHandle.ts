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

import * as Automerge from "@automerge/automerge";
import type { DocHandle } from "@automerge/prosemirror";
import {
  collaborationDebug,
  type RichEditorAutomergeDocument,
  type RichEditorPresence,
  summarizeAutomergeSpans,
} from "@probo/ui";

const collaborationProtocol = "automerge-sync-v1";
const maxGeneratedMessages = 100;
const initializationTimeoutMs = 35_000;
const reconnectMaxDelayMs = 10_000;
const presenceHeartbeatMs = 5_000;
const presenceUpdateThrottleMs = 50;

type CollaborationHandshake = {
  type: "ready";
  version: 1;
  revision: number;
  needsSeed: boolean;
  seedContent?: string;
  connectionId: string;
};

type ChangeListener = (
  payload: DocumentHandleChangePayload,
) => void;

type DocumentHandleChangePayload = {
  handle: DocHandle<RichEditorAutomergeDocument>;
  doc: Automerge.Doc<RichEditorAutomergeDocument>;
  patches: Automerge.Patch[];
  patchInfo: Automerge.PatchInfo<RichEditorAutomergeDocument>;
};

type PresenceSnapshot = {
  type: "presence";
  presences: Array<{
    connectionId: string;
    identityId: string;
    anchorPosition: number;
    headPosition: number;
  }>;
};

export class AutomergeDocumentHandle implements DocHandle<RichEditorAutomergeDocument> {
  readonly #endpoint: URL;
  #socket: WebSocket;
  #document: Automerge.Doc<RichEditorAutomergeDocument>;
  #syncState: Automerge.SyncState;
  #listeners = new Set<ChangeListener>();
  #presenceListeners = new Set<(presences: RichEditorPresence[]) => void>();
  #onConnectionState?: (connected: boolean) => void;
  #closed = false;
  #ready = false;
  #reconnectAttempt = 0;
  #reconnectTimer?: number;
  #presenceTimer?: number;
  #presenceHeartbeat?: number;
  #pendingPresence?: { anchorPosition: number; headPosition: number };

  constructor(
    endpoint: URL,
    socket: WebSocket,
    document: Automerge.Doc<RichEditorAutomergeDocument>,
    onConnectionState?: (connected: boolean) => void,
  ) {
    this.#endpoint = endpoint;
    this.#socket = socket;
    this.#document = document;
    this.#syncState = Automerge.initSyncState();
    this.#onConnectionState = onConnectionState;

    this.#attachSocket(socket);
    this.#presenceHeartbeat = window.setInterval(
      () => this.#sendPresence(),
      presenceHeartbeatMs,
    );
    this.#sendAvailableMessages();
  }

  doc(): Automerge.Doc<RichEditorAutomergeDocument> {
    return this.#document;
  }

  change(fn: (document: RichEditorAutomergeDocument) => void): void {
    let patches: Automerge.Patch[] = [];
    let patchInfo: Automerge.PatchInfo<RichEditorAutomergeDocument> | undefined;
    const headsBefore = Automerge.getHeads(this.#document);

    this.#document = Automerge.change(
      this.#document,
      {
        patchCallback: (nextPatches, nextPatchInfo) => {
          patches = nextPatches;
          patchInfo = nextPatchInfo;
        },
      },
      fn,
    );
    collaborationDebug("handle-local-change", {
      headsBefore,
      headsAfter: Automerge.getHeads(this.#document),
      patches: summarizePatches(patches),
      spans: summarizeAutomergeSpans(this.#document),
    });
    if (patchInfo) {
      this.#emit(patches, patchInfo);
    }
    this.#sendAvailableMessages();
  }

  on(event: "change", callback: ChangeListener): void {
    if (event === "change") this.#listeners.add(callback);
  }

  off(event: "change", callback: ChangeListener): void {
    if (event === "change") this.#listeners.delete(callback);
  }

  updatePresence(anchorPosition: number, headPosition: number): void {
    this.#pendingPresence = { anchorPosition, headPosition };
    if (this.#presenceTimer !== undefined) return;

    this.#presenceTimer = window.setTimeout(() => {
      this.#presenceTimer = undefined;
      this.#sendPresence();
    }, presenceUpdateThrottleMs);
  }

  onPresence(listener: (presences: RichEditorPresence[]) => void): () => void {
    this.#presenceListeners.add(listener);
    return () => this.#presenceListeners.delete(listener);
  }

  close(): void {
    this.#closed = true;
    if (this.#reconnectTimer !== undefined) {
      window.clearTimeout(this.#reconnectTimer);
    }
    if (this.#presenceTimer !== undefined) {
      window.clearTimeout(this.#presenceTimer);
    }
    if (this.#presenceHeartbeat !== undefined) {
      window.clearInterval(this.#presenceHeartbeat);
    }
    this.#detachSocket(this.#socket);
    this.#socket.close(1000);
    this.#listeners.clear();
    this.#presenceListeners.clear();
  }

  waitUntilReady(): Promise<void> {
    if (typeof this.#document.body === "string") {
      this.#ready = true;
      return Promise.resolve();
    }

    return new Promise((resolve, reject) => {
      const timeout = window.setTimeout(() => {
        cleanup();
        reject(new Error("Document collaboration initialization timed out"));
      }, initializationTimeoutMs);

      const handleChange: ChangeListener = ({ doc }) => {
        if (typeof doc.body !== "string") return;
        cleanup();
        this.#ready = true;
        resolve();
      };
      const handleClose = () => {
        cleanup();
        reject(new Error("Document collaboration closed during initialization"));
      };
      const cleanup = () => {
        window.clearTimeout(timeout);
        this.off("change", handleChange);
        this.#socket.removeEventListener("close", handleClose);
      };

      this.on("change", handleChange);
      this.#socket.addEventListener("close", handleClose);
    });
  }

  #handleMessage = (event: MessageEvent) => {
    if (typeof event.data === "string") {
      let value: unknown;
      try {
        value = JSON.parse(event.data);
      } catch {
        return;
      }
      if (!isPresenceSnapshot(value)) return;
      const presences = value.presences.map(presence => ({
        connectionID: presence.connectionId,
        identityID: presence.identityId,
        anchorPosition: presence.anchorPosition,
        headPosition: presence.headPosition,
      }));
      for (const listener of this.#presenceListeners) {
        listener(presences);
      }
      return;
    }
    if (!(event.data instanceof ArrayBuffer)) return;

    let patches: Automerge.Patch[] = [];
    let patchInfo: Automerge.PatchInfo<RichEditorAutomergeDocument> | undefined;
    const headsBefore = Automerge.getHeads(this.#document);
    [this.#document, this.#syncState] = Automerge.receiveSyncMessage(
      this.#document,
      this.#syncState,
      new Uint8Array(event.data),
      {
        patchCallback: (nextPatches, nextPatchInfo) => {
          patches = nextPatches;
          patchInfo = nextPatchInfo;
        },
      },
    );
    collaborationDebug("sync-receive", {
      bytes: event.data.byteLength,
      headsBefore,
      headsAfter: Automerge.getHeads(this.#document),
      patches: summarizePatches(patches),
      spans: summarizeAutomergeSpans(this.#document),
    });
    if (patchInfo) {
      this.#emit(patches, patchInfo);
    }
    this.#sendAvailableMessages();
  };

  #handleDisconnect = () => {
    if (this.#closed || !this.#ready || this.#reconnectTimer !== undefined) return;

    collaborationDebug("socket-disconnected", {
      ready: this.#ready,
      attempt: this.#reconnectAttempt,
      heads: Automerge.getHeads(this.#document),
    });
    this.#detachSocket(this.#socket);
    this.#onConnectionState?.(false);
    const delay = Math.min(
      500 * 2 ** this.#reconnectAttempt,
      reconnectMaxDelayMs,
    );
    this.#reconnectAttempt++;
    this.#reconnectTimer = window.setTimeout(() => {
      this.#reconnectTimer = undefined;
      void this.#reconnect();
    }, delay);
  };

  async #reconnect(): Promise<void> {
    if (this.#closed) return;

    const socket = createSocket(this.#endpoint);
    try {
      await waitForHandshake(socket);
      if (this.#closed) {
        socket.close(1000);
        return;
      }

      this.#socket = socket;
      this.#syncState = Automerge.initSyncState();
      this.#reconnectAttempt = 0;
      this.#attachSocket(socket);
      this.#sendAvailableMessages();
      this.#sendPresence();
      this.#onConnectionState?.(true);
      collaborationDebug("socket-reconnected", {
        heads: Automerge.getHeads(this.#document),
      });
    } catch {
      socket.close();
      if (this.#closed) return;
      this.#reconnectTimer = window.setTimeout(() => {
        this.#reconnectTimer = undefined;
        void this.#reconnect();
      }, reconnectMaxDelayMs);
    }
  }

  #attachSocket(socket: WebSocket): void {
    socket.addEventListener("message", this.#handleMessage);
    socket.addEventListener("close", this.#handleDisconnect);
    socket.addEventListener("error", this.#handleDisconnect);
  }

  #detachSocket(socket: WebSocket): void {
    socket.removeEventListener("message", this.#handleMessage);
    socket.removeEventListener("close", this.#handleDisconnect);
    socket.removeEventListener("error", this.#handleDisconnect);
  }

  #sendAvailableMessages(): void {
    if (this.#socket.readyState !== WebSocket.OPEN) return;

    for (let i = 0; i < maxGeneratedMessages; i++) {
      let message: Automerge.SyncMessage | null;
      [this.#syncState, message] = Automerge.generateSyncMessage(
        this.#document,
        this.#syncState,
      );
      if (!message) return;
      const data = new ArrayBuffer(message.byteLength);
      new Uint8Array(data).set(message);
      collaborationDebug("sync-send", {
        sequence: i,
        bytes: message.byteLength,
        heads: Automerge.getHeads(this.#document),
      });
      this.#socket.send(data);
    }

    throw new Error("Automerge sync protocol did not quiesce");
  }

  #sendPresence(): void {
    if (
      !this.#pendingPresence
      || this.#socket.readyState !== WebSocket.OPEN
    ) {
      return;
    }
    this.#socket.send(
      JSON.stringify({
        type: "presence",
        ...this.#pendingPresence,
      }),
    );
  }

  #emit(
    patches: Automerge.Patch[],
    patchInfo: Automerge.PatchInfo<RichEditorAutomergeDocument>,
  ): void {
    const payload: DocumentHandleChangePayload = {
      handle: this,
      doc: this.#document,
      patches,
      patchInfo,
    };
    for (const listener of this.#listeners) {
      listener(payload);
    }
  }
}

export async function connectAutomergeDocument(
  documentVersionID: string,
  createSeed: (content: string) => Automerge.Doc<RichEditorAutomergeDocument>,
  onConnectionState?: (connected: boolean) => void,
): Promise<AutomergeDocumentHandle> {
  const endpoint = new URL(window.location.origin);
  endpoint.protocol = endpoint.protocol === "https:" ? "wss:" : "ws:";
  endpoint.pathname = [
    "api",
    "console",
    "v1",
    "document-versions",
    encodeURIComponent(documentVersionID),
    "sync",
  ].join("/");

  const socket = createSocket(endpoint);

  const handshake = await waitForHandshake(socket);
  collaborationDebug("handshake", {
    revision: handshake.revision,
    needsSeed: handshake.needsSeed,
    connectionId: handshake.connectionId,
  });
  const document = handshake.needsSeed
    ? createSeed(handshake.seedContent ?? "")
    : Automerge.init<RichEditorAutomergeDocument>();

  const handle = new AutomergeDocumentHandle(
    endpoint,
    socket,
    document,
    onConnectionState,
  );
  try {
    await handle.waitUntilReady();
    return handle;
  } catch (error) {
    handle.close();
    throw error;
  }
}

function createSocket(endpoint: URL): WebSocket {
  const socket = new WebSocket(endpoint, collaborationProtocol);
  socket.binaryType = "arraybuffer";
  return socket;
}

function summarizePatches(
  patches: Automerge.Patch[],
): Array<Record<string, unknown>> {
  return patches.map(patch => ({
    action: patch.action,
    path: patch.path,
  }));
}

function waitForHandshake(socket: WebSocket): Promise<CollaborationHandshake> {
  return new Promise((resolve, reject) => {
    function cleanup() {
      socket.removeEventListener("message", handleMessage);
      socket.removeEventListener("error", handleError);
      socket.removeEventListener("close", handleClose);
    }

    function handleMessage(event: MessageEvent) {
      if (typeof event.data !== "string") {
        cleanup();
        reject(new Error("Collaboration server sent binary data before handshake"));
        return;
      }

      let value: unknown;
      try {
        value = JSON.parse(event.data);
      } catch {
        cleanup();
        reject(new Error("Collaboration server returned an invalid handshake"));
        return;
      }
      if (!isCollaborationHandshake(value)) {
        cleanup();
        reject(new Error("Collaboration server returned an unsupported handshake"));
        return;
      }

      cleanup();
      resolve(value);
    }

    function handleError() {
      cleanup();
      reject(new Error("Cannot connect to document collaboration server"));
    }

    function handleClose() {
      cleanup();
      reject(new Error("Document collaboration connection closed before handshake"));
    }

    socket.addEventListener("message", handleMessage);
    socket.addEventListener("error", handleError);
    socket.addEventListener("close", handleClose);
  });
}

function isCollaborationHandshake(value: unknown): value is CollaborationHandshake {
  if (!value || typeof value !== "object") return false;
  const handshake = value as Record<string, unknown>;
  return handshake.type === "ready"
    && handshake.version === 1
    && typeof handshake.revision === "number"
    && typeof handshake.needsSeed === "boolean"
    && typeof handshake.connectionId === "string"
    && (
      handshake.seedContent === undefined
      || typeof handshake.seedContent === "string"
    );
}

function isPresenceSnapshot(value: unknown): value is PresenceSnapshot {
  if (!value || typeof value !== "object") return false;
  const snapshot = value as Record<string, unknown>;
  if (snapshot.type !== "presence" || !Array.isArray(snapshot.presences)) {
    return false;
  }
  return snapshot.presences.every((presence) => {
    if (!presence || typeof presence !== "object") return false;
    const item = presence as Record<string, unknown>;
    return typeof item.connectionId === "string"
      && typeof item.identityId === "string"
      && typeof item.anchorPosition === "number"
      && typeof item.headPosition === "number";
  });
}
