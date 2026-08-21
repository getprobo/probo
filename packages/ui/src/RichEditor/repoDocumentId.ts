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

// automerge-repo document ids, mirrored byte-for-byte from the Go implementation
// in pkg/automerge/collaboration/documentid.go so a browser tab and a Go agent
// compute the same id for a document version without coordinating. This is what
// lets them share sync and, crucially, ephemeral gossip (presence and cursors):
// a peer drops an ephemeral frame whose document id it does not recognise.
//
// The format is base58check (base58 of the payload followed by the first four
// bytes of its double SHA-256), the same encoding @automerge/automerge-repo uses
// for a 16-byte document id.

/** The automerge: URL scheme automerge-repo puts in front of a document id. */
export const AUTOMERGE_URL_PREFIX = "automerge:";

/** The length of the binary document id automerge-repo base58check-encodes. */
export const DOCUMENT_ID_BYTE_LENGTH = 16;

// The Bitcoin base58 alphabet automerge-repo's bs58check uses.
const BASE58_ALPHABET =
  "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";
const BASE58_RADIX = 58n;

/** Encodes a 16-byte identifier as an automerge-repo document id. */
export async function encodeDocumentId(id: Uint8Array): Promise<string> {
  if (id.length !== DOCUMENT_ID_BYTE_LENGTH) {
    throw new Error(
      `automerge document id must be ${DOCUMENT_ID_BYTE_LENGTH} bytes, got ${id.length}`,
    );
  }

  return base58CheckEncode(id);
}

/** Decodes an automerge-repo document id, rejecting a bad checksum or length. */
export async function decodeDocumentId(
  documentId: string,
): Promise<Uint8Array> {
  const payload = await base58CheckDecode(documentId);
  if (payload.length !== DOCUMENT_ID_BYTE_LENGTH) {
    throw new Error(
      `automerge document id ${JSON.stringify(documentId)} decodes to ${payload.length} bytes, want ${DOCUMENT_ID_BYTE_LENGTH}`,
    );
  }

  return payload;
}

/** Reports whether documentId is a well-formed automerge-repo document id. */
export async function validDocumentId(documentId: string): Promise<boolean> {
  try {
    await decodeDocumentId(documentId);
    return true;
  } catch {
    return false;
  }
}

/**
 * Derives a stable automerge-repo document id from a seed string, such as a
 * Probo document-version GID, by hashing the seed and taking the first 16 bytes.
 * Every peer that knows the seed computes the same id.
 */
export async function deriveDocumentId(seed: string): Promise<string> {
  const digest = await sha256(new TextEncoder().encode(seed));

  return base58CheckEncode(digest.slice(0, DOCUMENT_ID_BYTE_LENGTH));
}

/** Wraps a document id in the automerge: scheme. */
export function automergeUrl(documentId: string): string {
  return AUTOMERGE_URL_PREFIX + documentId;
}

/** Derives a stable automerge: URL from a seed string. */
export async function deriveAutomergeUrl(seed: string): Promise<string> {
  return automergeUrl(await deriveDocumentId(seed));
}

/** Extracts and validates the document id from an automerge: URL. */
export async function parseAutomergeUrl(url: string): Promise<string> {
  if (!url.startsWith(AUTOMERGE_URL_PREFIX)) {
    throw new Error(
      `automerge url ${JSON.stringify(url)} is missing the ${JSON.stringify(AUTOMERGE_URL_PREFIX)} scheme`,
    );
  }

  const documentId = url.slice(AUTOMERGE_URL_PREFIX.length);
  await decodeDocumentId(documentId);

  return documentId;
}

async function base58CheckEncode(payload: Uint8Array): Promise<string> {
  const check = await checksum(payload);
  const combined = new Uint8Array(payload.length + check.length);
  combined.set(payload);
  combined.set(check, payload.length);

  return base58Encode(combined);
}

async function base58CheckDecode(encoded: string): Promise<Uint8Array> {
  const decoded = base58Decode(encoded);
  if (decoded.length < 4) {
    throw new Error("base58check value is too short to contain a checksum");
  }

  const payload = decoded.slice(0, decoded.length - 4);
  const want = decoded.slice(decoded.length - 4);
  const got = await checksum(payload);

  for (let i = 0; i < 4; i++) {
    if (got[i] !== want[i]) {
      throw new Error("base58check checksum mismatch");
    }
  }

  return payload;
}

async function checksum(payload: Uint8Array): Promise<Uint8Array> {
  const first = await sha256(payload);
  const second = await sha256(first);

  return second.slice(0, 4);
}

async function sha256(data: Uint8Array): Promise<Uint8Array> {
  // Copy into a fresh ArrayBuffer-backed view so the argument is a plain
  // BufferSource regardless of the input's backing store (SharedArrayBuffer).
  const bytes = new Uint8Array(data.length);
  bytes.set(data);
  const digest = await crypto.subtle.digest("SHA-256", bytes.buffer);

  return new Uint8Array(digest);
}

function base58Encode(input: Uint8Array): string {
  let value = 0n;
  for (const byte of input) {
    value = value * 256n + BigInt(byte);
  }

  let encoded = "";
  while (value > 0n) {
    const remainder = value % BASE58_RADIX;
    value = value / BASE58_RADIX;
    encoded = BASE58_ALPHABET[Number(remainder)] + encoded;
  }

  // Each leading zero byte is encoded as the alphabet's first character.
  for (const byte of input) {
    if (byte !== 0) {
      break;
    }

    encoded = BASE58_ALPHABET[0] + encoded;
  }

  return encoded;
}

function base58Decode(encoded: string): Uint8Array {
  let value = 0n;
  for (const character of encoded) {
    const index = BASE58_ALPHABET.indexOf(character);
    if (index < 0) {
      throw new Error(`invalid base58 character ${JSON.stringify(character)}`);
    }

    value = value * BASE58_RADIX + BigInt(index);
  }

  const digits: number[] = [];
  while (value > 0n) {
    digits.unshift(Number(value % 256n));
    value = value / 256n;
  }

  // Restore the leading zero bytes the encoder wrote as leading '1's.
  let zeros = 0;
  for (const character of encoded) {
    if (character !== BASE58_ALPHABET[0]) {
      break;
    }

    zeros++;
  }

  const result = new Uint8Array(zeros + digits.length);
  result.set(digits, zeros);

  return result;
}
