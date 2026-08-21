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

// A real automerge-repo client used as the interop oracle for the Go gateway.
// It connects to a Go WebSocket server that speaks the repo protocol, finds a
// document by URL, and prints the materialized document as JSON. It is driven by
// the Go test in pkg/automerge/collaboration.
//
// Usage: node collaboration-interop-client.mjs <wsURL> <automergeUrl>

import { Repo } from "@automerge/automerge-repo";
import { WebSocketClientAdapter } from "@automerge/automerge-repo-network-websocket";
import process from "node:process";

const [wsURL, documentURL] = process.argv.slice(2);

if (!wsURL || !documentURL) {
  process.stderr.write("usage: collaboration-interop-client.mjs <wsURL> <automergeUrl>\n");
  process.exit(2);
}

const repo = new Repo({
  network: [new WebSocketClientAdapter(wsURL)],
});

try {
  const handle = await repo.find(documentURL, {
    signal: AbortSignal.timeout(10_000),
  });

  const doc = handle.doc();
  process.stdout.write(JSON.stringify(doc ?? null));
  process.exit(0);
} catch (error) {
  process.stderr.write(`interop client failed: ${error?.message ?? error}\n`);
  process.exit(1);
}
