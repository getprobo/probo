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

for (const [name, fixture] of Object.entries(fixtures)) {
  const path = join(outputDir, `${name}.json`);
  writeFileSync(path, `${JSON.stringify(fixture, null, 2)}\n`);
  process.stdout.write(`wrote ${path}\n`);
}
