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

// Generates byte-exact automerge-repo protocol fixtures for the Go collaboration
// codec, using the pinned packages' own CBOR encoder so the bytes match what the
// JavaScript client puts on the wire. Run via
// `make generate-automerge-collaboration-fixtures`.

import { cbor } from "@automerge/automerge-repo";
import { Buffer } from "node:buffer";
import process from "node:process";

// The presence envelope marker key (PRESENCE_MESSAGE_MARKER in the upstream
// source). It is a stable protocol constant; the package does not export the
// constants module, so it is inlined here and documented in PROTOCOL.md.
const PRESENCE_MESSAGE_MARKER = "__presence";
import { mkdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const outputDir = join(
  dirname(fileURLToPath(import.meta.url)),
  "..",
  "..",
  "pkg",
  "automerge",
  "collaboration",
  "testdata",
);

mkdirSync(outputDir, { recursive: true });

const base64 = value => Buffer.from(value).toString("base64");

// Each presence message, wrapped in the __presence envelope exactly as
// Presence.broadcast does before DocHandle.broadcast CBOR-encodes it.
const presenceMessages = {
  update: { type: "update", channel: "cursor", value: { anchor: "a", head: "b" } },
  snapshot: { type: "snapshot", state: { cursor: { anchor: "a", head: "b" } } },
  heartbeat: { type: "heartbeat" },
  goodbye: { type: "goodbye" },
};

const fixtures = {};

for (const [name, message] of Object.entries(presenceMessages)) {
  const envelope = { [PRESENCE_MESSAGE_MARKER]: message };
  const data = cbor.encode(envelope);

  fixtures[`presence-${name}`] = {
    description: `Presence ${name} envelope CBOR-encoded as an ephemeral payload`,
    marker: PRESENCE_MESSAGE_MARKER,
    envelope,
    cborBase64: base64(data),
  };

  // A full ephemeral repo message carrying this presence payload.
  const ephemeral = {
    type: "ephemeral",
    senderId: "peer-a",
    targetId: "peer-b",
    documentId: "4NMNnkMhL2wRfvHYuG1uxN",
    sessionId: "session-a",
    count: 1,
    data,
  };

  fixtures[`ephemeral-${name}`] = {
    description: `Ephemeral repo message wrapping a presence ${name}`,
    message: { ...ephemeral, data: base64(ephemeral.data) },
    payloadCborBase64: base64(data),
  };
}

// A round-trip guard: CBOR that decodes back to the envelope it came from.
fixtures["presence-roundtrip"] = (() => {
  const envelope = {
    [PRESENCE_MESSAGE_MARKER]: {
      type: "update",
      channel: "cursor",
      value: { anchor: "AAEC", head: "AwQF" },
    },
  };
  const data = cbor.encode(envelope);
  const decoded = cbor.decode(data);

  return {
    description: "CBOR round-trip of a presence update envelope",
    envelope,
    cborBase64: base64(data),
    decodesEqual: JSON.stringify(decoded) === JSON.stringify(envelope),
  };
})();

// Transport-layer frames, encoded with the same repo CBOR helper the WebSocket
// adapter uses for every frame. Each frame is one CBOR-encoded message.
const peerMetadata = { isEphemeral: false };

const wireFrames = {
  "wire-join": {
    description: "Client join handshake frame",
    message: {
      type: "join",
      senderId: "peer-a",
      peerMetadata,
      supportedProtocolVersions: ["1"],
    },
  },
  "wire-peer": {
    description: "Server peer handshake reply frame",
    message: {
      type: "peer",
      senderId: "server",
      targetId: "peer-a",
      peerMetadata,
      selectedProtocolVersion: "1",
    },
  },
  "wire-error": {
    description: "Server error frame before closing the socket",
    message: {
      type: "error",
      senderId: "server",
      targetId: "peer-a",
      message: "unauthorized",
    },
  },
  "wire-sync": {
    description: "Framed sync message carrying opaque Automerge sync bytes",
    message: {
      type: "sync",
      senderId: "peer-a",
      targetId: "server",
      documentId: "4NMNnkMhL2wRfvHYuG1uxN",
      data: new Uint8Array([0, 1, 2, 3]),
    },
  },
  "wire-ephemeral": {
    description: "Framed ephemeral message carrying a presence heartbeat payload",
    message: {
      type: "ephemeral",
      senderId: "peer-a",
      targetId: "server",
      documentId: "4NMNnkMhL2wRfvHYuG1uxN",
      sessionId: "session-a",
      count: 1,
      data: cbor.encode({ [PRESENCE_MESSAGE_MARKER]: { type: "heartbeat" } }),
    },
  },
};

for (const [name, frame] of Object.entries(wireFrames)) {
  const encoded = cbor.encode(frame.message);
  const message = { ...frame.message };
  if (message.data instanceof Uint8Array) {
    message.data = base64(message.data);
  }

  fixtures[name] = {
    description: frame.description,
    message,
    frameCborBase64: base64(encoded),
  };
}

for (const [name, fixture] of Object.entries(fixtures)) {
  const path = join(outputDir, `${name}.json`);
  writeFileSync(path, `${JSON.stringify(fixture, null, 2)}\n`);
  process.stdout.write(`wrote ${path}\n`);
}
