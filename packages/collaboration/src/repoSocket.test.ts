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

import { cbor, type PeerId } from "@automerge/automerge-repo/slim";
import { ProtocolV1 } from "@automerge/automerge-repo-network-websocket";
import { afterEach, describe, expect, it, vi } from "vitest";

import { repoHandshakeTimeoutMs, RepoWebSocketAdapter } from "./repoSocket";

const opened: FakeWebSocket[] = [];

class FakeWebSocket extends EventTarget {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
  static reachable = true;

  readyState = FakeWebSocket.CONNECTING;
  binaryType = "blob";
  sent: ArrayBuffer[] = [];
  url: string;

  constructor(url: string) {
    super();
    this.url = url;
    opened.push(this);
    if (!FakeWebSocket.reachable) return;

    queueMicrotask(() => {
      this.readyState = FakeWebSocket.OPEN;
      this.dispatchEvent(new Event("open"));
    });
  }

  send(data: ArrayBuffer) {
    this.sent.push(data);
  }

  close() {
    this.readyState = FakeWebSocket.CLOSED;
    this.dispatchEvent(new Event("close"));
  }
}

const OriginalWebSocket = globalThis.WebSocket;

afterEach(() => {
  globalThis.WebSocket = OriginalWebSocket;
  vi.useRealTimers();
  opened.length = 0;
  FakeWebSocket.reachable = true;
});

async function connectAdapter(): Promise<{
  adapter: RepoWebSocketAdapter;
  socket: FakeWebSocket;
}> {
  globalThis.WebSocket = FakeWebSocket as unknown as typeof WebSocket;
  vi.useFakeTimers();
  const adapter = new RepoWebSocketAdapter("ws://collaboration.test/doc");
  adapter.connect("client-far" as PeerId);
  const socket = opened[opened.length - 1];
  await Promise.resolve();
  expect(socket?.readyState).toBe(FakeWebSocket.OPEN);
  if (!socket) {
    throw new Error("fake socket was not created");
  }
  return { adapter, socket };
}

describe("RepoWebSocketAdapter", () => {
  it("sends the join frame when the socket opens", async () => {
    const { adapter, socket } = await connectAdapter();

    expect(socket.sent).toHaveLength(1);
    const join = cbor.decode<Record<string, unknown>>(
      new Uint8Array(socket.sent[0]),
    );
    expect(join).toMatchObject({
      type: "join",
      senderId: "client-far",
      supportedProtocolVersions: [ProtocolV1],
    });
    adapter.disconnect();
  });

  it("gives up on an unreachable server despite retries", async () => {
    globalThis.WebSocket = FakeWebSocket as unknown as typeof WebSocket;
    FakeWebSocket.reachable = false;
    vi.useFakeTimers();
    const adapter = new RepoWebSocketAdapter("ws://collaboration.test/doc");
    adapter.connect("client-far" as PeerId);

    await vi.advanceTimersByTimeAsync(repoHandshakeTimeoutMs - 1);
    expect(opened.length).toBeGreaterThan(1);
    expect(adapter.isReady()).toBe(false);

    await vi.advanceTimersByTimeAsync(1);
    expect(adapter.isReady()).toBe(true);
    adapter.disconnect();
  });

  it("closes the socket a retry replaces", async () => {
    globalThis.WebSocket = FakeWebSocket as unknown as typeof WebSocket;
    FakeWebSocket.reachable = false;
    vi.useFakeTimers();
    const adapter = new RepoWebSocketAdapter("ws://collaboration.test/doc");
    adapter.connect("client-far" as PeerId);

    await vi.advanceTimersByTimeAsync(5_000);

    expect(opened).toHaveLength(2);
    expect(opened[0].readyState).toBe(FakeWebSocket.CLOSED);
    adapter.disconnect();
  });

  it("stays pending through the stock one second silence", async () => {
    const { adapter } = await connectAdapter();

    await vi.advanceTimersByTimeAsync(1_000);

    expect(adapter.isReady()).toBe(false);
    adapter.disconnect();
  });

  it("becomes ready when the peer reply arrives after one second", async () => {
    const { adapter, socket } = await connectAdapter();
    await vi.advanceTimersByTimeAsync(1_000);

    const encoded = cbor.encode({
      type: "peer",
      senderId: "probo-gateway",
      targetId: "client-far",
      peerMetadata: {},
      selectedProtocolVersion: ProtocolV1,
    });
    socket.dispatchEvent(new MessageEvent("message", { data: encoded }));

    expect(adapter.isReady()).toBe(true);
    adapter.disconnect();
  });

  it("becomes ready once the handshake budget expires", async () => {
    const { adapter } = await connectAdapter();

    await vi.advanceTimersByTimeAsync(repoHandshakeTimeoutMs - 1);
    expect(adapter.isReady()).toBe(false);

    await vi.advanceTimersByTimeAsync(1);
    expect(adapter.isReady()).toBe(true);
    adapter.disconnect();
  });
});
