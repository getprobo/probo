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
import type { RichEditorAutomergeDocument } from "@probo/ui";

const collaborationProtocol = "automerge-sync-v1";
const maxGeneratedMessages = 100;
const initializationTimeoutMs = 35_000;

type CollaborationHandshake = {
  type: "ready";
  version: 1;
  revision: number;
  needsSeed: boolean;
  seedContent?: string;
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

export class AutomergeDocumentHandle implements DocHandle<RichEditorAutomergeDocument> {
  readonly #socket: WebSocket;
  #document: Automerge.Doc<RichEditorAutomergeDocument>;
  #syncState: Automerge.SyncState;
  #listeners = new Set<ChangeListener>();
  #onDisconnect?: () => void;
  #closed = false;
  #disconnectNotified = false;

  constructor(
    socket: WebSocket,
    document: Automerge.Doc<RichEditorAutomergeDocument>,
    onDisconnect?: () => void,
  ) {
    this.#socket = socket;
    this.#document = document;
    this.#syncState = Automerge.initSyncState();
    this.#onDisconnect = onDisconnect;

    socket.addEventListener("message", this.#handleMessage);
    socket.addEventListener("close", this.#handleDisconnect);
    socket.addEventListener("error", this.#handleDisconnect);
    this.#sendAvailableMessages();
  }

  doc(): Automerge.Doc<RichEditorAutomergeDocument> {
    return this.#document;
  }

  change(fn: (document: RichEditorAutomergeDocument) => void): void {
    let patches: Automerge.Patch[] = [];
    let patchInfo: Automerge.PatchInfo<RichEditorAutomergeDocument> | undefined;

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

  close(): void {
    this.#closed = true;
    this.#socket.removeEventListener("message", this.#handleMessage);
    this.#socket.removeEventListener("close", this.#handleDisconnect);
    this.#socket.removeEventListener("error", this.#handleDisconnect);
    this.#socket.close(1000);
    this.#listeners.clear();
  }

  waitUntilReady(): Promise<void> {
    if (typeof this.#document.body === "string") return Promise.resolve();

    return new Promise((resolve, reject) => {
      const timeout = window.setTimeout(() => {
        cleanup();
        reject(new Error("Document collaboration initialization timed out"));
      }, initializationTimeoutMs);

      const handleChange: ChangeListener = ({ doc }) => {
        if (typeof doc.body !== "string") return;
        cleanup();
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
    if (!(event.data instanceof ArrayBuffer)) return;

    let patches: Automerge.Patch[] = [];
    let patchInfo: Automerge.PatchInfo<RichEditorAutomergeDocument> | undefined;
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
    if (patchInfo) {
      this.#emit(patches, patchInfo);
    }
    this.#sendAvailableMessages();
  };

  #handleDisconnect = () => {
    if (this.#closed || this.#disconnectNotified) return;
    this.#disconnectNotified = true;
    this.#onDisconnect?.();
  };

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
      this.#socket.send(data);
    }

    throw new Error("Automerge sync protocol did not quiesce");
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
  onDisconnect?: () => void,
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

  const socket = new WebSocket(endpoint, collaborationProtocol);
  socket.binaryType = "arraybuffer";

  const handshake = await waitForHandshake(socket);
  const document = handshake.needsSeed
    ? createSeed(handshake.seedContent ?? "")
    : Automerge.init<RichEditorAutomergeDocument>();

  let ready = false;
  const handle = new AutomergeDocumentHandle(socket, document, () => {
    if (ready) onDisconnect?.();
  });
  try {
    await handle.waitUntilReady();
    ready = true;
    return handle;
  } catch (error) {
    handle.close();
    throw error;
  }
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
    && (
      handshake.seedContent === undefined
      || typeof handshake.seedContent === "string"
    );
}
