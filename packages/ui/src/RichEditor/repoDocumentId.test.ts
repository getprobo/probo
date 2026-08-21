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

import { describe, expect, it } from "vitest";

import {
  automergeUrl,
  decodeDocumentId,
  deriveAutomergeUrl,
  deriveDocumentId,
  encodeDocumentId,
  parseAutomergeUrl,
  validDocumentId,
} from "./repoDocumentId";

// A genuine @automerge/automerge-repo document id (the one the Go interop client
// uses). Decoding it validates our base58check against real upstream output.
const REAL_REPO_ID = "34YWzjYt5gPJpq5RfXAkPfPcUj1r";

// Canonical outputs of the Go DeriveDocumentID (pkg/automerge/collaboration).
// Equality here proves the browser and the Go agent derive the same id for a
// version, which is what makes their sync and presence line up.
const GO_DERIVED: Record<string, string> = {
  document_version_2Abc123: "2m7mUm61HVK58xqK8aYTGYLQ4atr",
  "gid:probo:document_version:01J000000000000000000000":
    "3r6ksouJrcqDzaHWPGCEzddsAEhW",
  "hello world": "3ajNHtb2g2k3cYHyXLUJFwQuyr42",
  "": "4AyyyhobrQ6KECw6yZaZ2Ss2eVuL",
};

describe("repo document id", () => {
  it("decodes a real automerge-repo id and round-trips it", async () => {
    const id = await decodeDocumentId(REAL_REPO_ID);
    expect(id).toHaveLength(16);
    expect(await encodeDocumentId(id)).toBe(REAL_REPO_ID);
    expect(await validDocumentId(REAL_REPO_ID)).toBe(true);
  });

  it("derives ids that match the Go implementation exactly", async () => {
    for (const [seed, expected] of Object.entries(GO_DERIVED)) {
      expect(await deriveDocumentId(seed)).toBe(expected);
    }
  });

  it("derives deterministically and per-seed", async () => {
    const seed = "document_version_9";
    expect(await deriveDocumentId(seed)).toBe(await deriveDocumentId(seed));
    expect(await deriveDocumentId(seed)).not.toBe(
      await deriveDocumentId(seed + "x"),
    );
  });

  it("round-trips arbitrary 16-byte ids, including leading zeros", async () => {
    const cases: Uint8Array[] = [
      new Uint8Array(16),
      new Uint8Array([0, 0, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 0, 0, 0]),
      new Uint8Array(16).fill(255),
      new Uint8Array([
        0xde, 0xad, 0xbe, 0xef, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
        0x88, 0x99, 0xaa, 0xbb,
      ]),
    ];

    for (const id of cases) {
      const encoded = await encodeDocumentId(id);
      expect(await validDocumentId(encoded)).toBe(true);
      expect(await decodeDocumentId(encoded)).toEqual(id);
    }
  });

  it("rejects a corrupted id", async () => {
    const flipped = REAL_REPO_ID.endsWith("r")
      ? REAL_REPO_ID.slice(0, -1) + "s"
      : REAL_REPO_ID.slice(0, -1) + "r";
    await expect(decodeDocumentId(flipped)).rejects.toThrow();

    await expect(decodeDocumentId("0OIl")).rejects.toThrow();
    expect(await validDocumentId("")).toBe(false);
  });

  it("wraps and parses the automerge: scheme", async () => {
    const url = await deriveAutomergeUrl("document_version_9");
    const documentId = await parseAutomergeUrl(url);
    expect(automergeUrl(documentId)).toBe(url);
    expect(await validDocumentId(documentId)).toBe(true);

    await expect(parseAutomergeUrl(REAL_REPO_ID)).rejects.toThrow();
    await expect(
      parseAutomergeUrl("automerge:not-a-valid-id!!"),
    ).rejects.toThrow();
  });
});
