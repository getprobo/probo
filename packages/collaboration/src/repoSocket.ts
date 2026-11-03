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
  cbor,
  type Message,
  NetworkAdapter,
  type PeerId,
  type PeerMetadata,
} from "@automerge/automerge-repo/slim";
import { ProtocolV1 } from "@automerge/automerge-repo-network-websocket";

// The stock adapter gives up after 1s with no peer and the document is
// reported unavailable. A far handshake plus a large snapshot needs longer.
export const repoHandshakeTimeoutMs = 30_000;

const retryIntervalMs = 5_000;

type ServerFrame = {
  type?: string;
  senderId?: PeerId;
  targetId?: PeerId;
  message?: string;
  peerMetadata?: PeerMetadata;
  selectedProtocolVersion?: string;
  documentId?: string;
  data?: Uint8Array;
};

function assert(value: unknown, message = "Assertion failed"): asserts value {
  if (value === false || value === null || value === undefined) {
    throw new Error(message);
  }
}

// A websocket view of a Uint8Array must not include bytes outside that view.
function toArrayBuffer(bytes: Uint8Array): ArrayBuffer {
  const { buffer, byteOffset, byteLength } = bytes;
  return buffer.slice(byteOffset, byteOffset + byteLength) as ArrayBuffer;
}

// RepoWebSocketAdapter is the browser side of one collaboration socket.
// It speaks the automerge-repo websocket protocol and waits
// repoHandshakeTimeoutMs for the server's peer reply before the network is
// treated as ready with nobody on it.
export class RepoWebSocketAdapter extends NetworkAdapter {
  socket?: WebSocket;

  #ready = false;
  #readyResolver?: () => void;
  #readyPromise: Promise<void> = new Promise((resolve) => {
    this.#readyResolver = resolve;
  });

  #retryIntervalId?: ReturnType<typeof setInterval>;
  #reconnectTimeoutId?: ReturnType<typeof setTimeout>;
  #handshakeTimeoutId?: ReturnType<typeof setTimeout>;
  #disconnected = false;

  remotePeerId?: PeerId;

  readonly url: string;

  constructor(url: string) {
    super();
    this.url = url;
  }

  isReady() {
    return this.#ready;
  }

  whenReady() {
    return this.#readyPromise;
  }

  #forceReady() {
    clearTimeout(this.#handshakeTimeoutId);
    this.#handshakeTimeoutId = undefined;
    if (this.#ready) return;

    this.#ready = true;
    this.#readyResolver?.();
  }

  connect(peerId: PeerId, peerMetadata?: PeerMetadata) {
    this.#disconnected = false;
    if (!this.socket || !this.peerId) {
      this.peerId = peerId;
      this.peerMetadata = peerMetadata ?? {};
    } else {
      assert(peerId === this.peerId);
      this.socket.removeEventListener("open", this.onOpen);
      this.socket.removeEventListener("close", this.onClose);
      this.socket.removeEventListener("message", this.onMessage);
      this.socket.removeEventListener("error", this.onError);
      // A replaced socket that opens later would sit on the server waiting
      // for a join nobody sends.
      this.socket.close();
    }

    if (!this.#retryIntervalId) {
      this.#retryIntervalId = setInterval(() => {
        this.connect(peerId, peerMetadata);
      }, retryIntervalMs);
    }

    this.socket = new WebSocket(this.url);
    this.socket.binaryType = "arraybuffer";
    this.socket.addEventListener("open", this.onOpen);
    this.socket.addEventListener("close", this.onClose);
    this.socket.addEventListener("message", this.onMessage);
    this.socket.addEventListener("error", this.onError);

    // Retries run every few seconds; re-arming here would postpone the
    // deadline forever while the server is unreachable.
    if (!this.#ready && this.#handshakeTimeoutId === undefined) {
      this.#handshakeTimeoutId = setTimeout(() => {
        this.#handshakeTimeoutId = undefined;
        this.#forceReady();
      }, repoHandshakeTimeoutMs);
    }
    this.join();
  }

  onOpen = () => {
    clearInterval(this.#retryIntervalId);
    this.#retryIntervalId = undefined;
    this.join();
  };

  onClose = () => {
    if (this.remotePeerId) {
      this.emit("peer-disconnected", { peerId: this.remotePeerId });
    }

    if (retryIntervalMs > 0 && !this.#retryIntervalId) {
      this.#reconnectTimeoutId = setTimeout(() => {
        if (this.#disconnected) return;
        assert(this.peerId);
        this.connect(this.peerId, this.peerMetadata);
      }, retryIntervalMs);
    }
  };

  onMessage = (event: MessageEvent<ArrayBuffer | Uint8Array>) => {
    const data = event.data;
    if (data instanceof ArrayBuffer) {
      this.receiveMessage(new Uint8Array(data));
      return;
    }
    if (data instanceof Uint8Array) {
      this.receiveMessage(data);
    }
  };

  onError = () => {};

  join() {
    assert(this.peerId);
    assert(this.socket);
    if (this.socket.readyState === WebSocket.OPEN) {
      this.#sendFrame({
        type: "join",
        senderId: this.peerId,
        peerMetadata: this.peerMetadata ?? {},
        supportedProtocolVersions: [ProtocolV1],
      });
    }
  }

  disconnect() {
    if (this.#disconnected) return;
    this.#disconnected = true;
    const socket = this.socket;
    if (socket) {
      socket.removeEventListener("open", this.onOpen);
      socket.removeEventListener("close", this.onClose);
      socket.removeEventListener("message", this.onMessage);
      socket.removeEventListener("error", this.onError);
      socket.close();
    }
    clearInterval(this.#retryIntervalId);
    this.#retryIntervalId = undefined;
    clearTimeout(this.#reconnectTimeoutId);
    this.#reconnectTimeoutId = undefined;
    clearTimeout(this.#handshakeTimeoutId);
    this.#handshakeTimeoutId = undefined;
    if (this.remotePeerId) {
      this.emit("peer-disconnected", { peerId: this.remotePeerId });
    }
    this.socket = undefined;
  }

  send(message: Message) {
    if ("data" in message && message.data?.byteLength === 0) {
      throw new Error("Tried to send a zero-length message");
    }
    assert(this.peerId);
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) return;

    this.#sendFrame(message);
  }

  #sendFrame(message: object) {
    assert(this.socket);
    if (this.socket.readyState !== WebSocket.OPEN) return;

    const encoded = cbor.encode(message);
    this.socket.send(toArrayBuffer(encoded));
  }

  receiveMessage(messageBytes: Uint8Array) {
    if (messageBytes.byteLength === 0) {
      throw new Error("received a zero-length message");
    }

    let message: ServerFrame;
    try {
      message = cbor.decode<ServerFrame>(messageBytes);
    } catch {
      return;
    }

    assert(this.socket);
    if (message.type === "peer" && message.senderId) {
      this.#forceReady();
      this.remotePeerId = message.senderId;
      this.emit("peer-candidate", {
        peerId: message.senderId,
        peerMetadata: message.peerMetadata ?? {},
      });
      return;
    }

    if (message.type === "error") return;

    this.emit("message", message as Message);
  }
}
