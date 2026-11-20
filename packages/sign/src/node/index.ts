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

// Node-only entry — do not import this in browser bundles.

import { createHmac, hkdfSync } from "node:crypto";

/** HKDF info string used to derive the JWT signing key. */
const HKDF_INFO = "probo-sign/identity/v1";

/**
 * Claims passed to signIdentity(). The function sets iat automatically.
 * Provide either expiresIn (a duration string) or exp (a Unix timestamp).
 */
export interface IdentityClaims {
  /** Customer's stable user identifier. Maps to external_user_id on the data plane. */
  sub: string;
  /** User's email address. */
  email: string;
  /** User's display name (optional). */
  name?: string;
  /** Customer's account/organization identifier (optional, for B2B). */
  account_id?: string;
  /**
   * Publishable key this JWT is bound to (aud claim).
   * Optional; include when you want to scope the JWT to a specific config.
   */
  aud?: string;
  /**
   * Token lifetime as a duration string ("10m", "1h") or seconds as a number.
   * Maximum: 3600 (1 hour). Defaults to 600 (10 minutes).
   */
  expiresIn?: string | number;
  /** Explicit exp Unix timestamp. Takes priority over expiresIn when both are set. */
  exp?: number;
  /**
   * Override iat (issued-at) Unix timestamp.
   * Leave unset in production; useful only for test vector verification.
   */
  iat?: number;
}

function base64url(data: Buffer): string {
  return data.toString("base64").replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}

function parseDuration(d: string | number): number {
  if (typeof d === "number") return d;
  const m = /^(\d+)(s|m|h)$/.exec(d);
  if (!m) throw new Error(`invalid expiresIn: ${d}. Use e.g. "10m", "1h", or a number of seconds.`);
  const n = parseInt(m[1], 10);
  const unit = m[2];
  if (unit === "s") return n;
  if (unit === "m") return n * 60;
  return n * 3600;
}

/**
 * Derive a 32-byte signing key from the organization secret using HKDF-SHA256.
 * The derived key is what HMAC-SHA256 operates on when signing JWTs.
 */
export function deriveSigningKey(orgSecret: string): Buffer {
  const ikm = Buffer.from(orgSecret, "utf8");
  const salt = Buffer.alloc(32, 0);
  const info = Buffer.from(HKDF_INFO, "utf8");
  return Buffer.from(hkdfSync("sha256", ikm, salt, info, 32));
}

/**
 * Sign an identity JWT using HS256 with an HKDF-derived key.
 *
 * Call this in your backend signup handler; pass the returned token to the
 * browser where it is provided to ProboDocuments({ identityToken }).
 *
 * Max token lifetime is 1 hour; the default is 10 minutes.
 */
export function signIdentity(orgSecret: string, claims: IdentityClaims): string {
  const iat = claims.iat ?? Math.floor(Date.now() / 1000);

  let exp: number;
  if (claims.exp !== undefined) {
    exp = claims.exp;
  } else {
    const ttl = parseDuration(claims.expiresIn ?? 600);
    if (ttl > 3600) throw new Error("identity token max lifetime is 1 hour (3600 seconds)");
    exp = iat + ttl;
  }

  // Payload field order is part of the JWT byte sequence; keep it stable.
  const payload: Record<string, unknown> = {
    sub: claims.sub,
    email: claims.email,
    iat,
    exp,
  };
  if (claims.name !== undefined) payload["name"] = claims.name;
  if (claims.account_id !== undefined) payload["account_id"] = claims.account_id;
  if (claims.aud !== undefined) payload["aud"] = claims.aud;

  const headerB64 = base64url(Buffer.from(JSON.stringify({ alg: "HS256", typ: "JWT" })));
  const payloadB64 = base64url(Buffer.from(JSON.stringify(payload)));
  const signingInput = `${headerB64}.${payloadB64}`;

  const key = deriveSigningKey(orgSecret);
  const sig = createHmac("sha256", key).update(signingInput).digest();

  return `${signingInput}.${base64url(sig)}`;
}
