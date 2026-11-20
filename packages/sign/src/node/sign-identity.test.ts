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

import { createHmac } from "node:crypto";
import { describe, expect, it } from "vitest";
import { deriveSigningKey, signIdentity } from "./index";

function decodeJwt(token: string): {
  header: Record<string, unknown>;
  payload: Record<string, unknown>;
  valid: boolean;
  signingInput: string;
  signature: Buffer;
} {
  const [h, p, s] = token.split(".");
  const header = JSON.parse(Buffer.from(h, "base64url").toString()) as Record<string, unknown>;
  const payload = JSON.parse(Buffer.from(p, "base64url").toString()) as Record<string, unknown>;
  const signature = Buffer.from(s, "base64url");
  return { header, payload, signingInput: `${h}.${p}`, signature, valid: true };
}

describe("deriveSigningKey", () => {
  it("produces a 32-byte Buffer", () => {
    const key = deriveSigningKey("sk_live_testsecret");
    expect(key).toBeInstanceOf(Buffer);
    expect(key.byteLength).toBe(32);
  });

  it("is deterministic", () => {
    const k1 = deriveSigningKey("sk_live_abc");
    const k2 = deriveSigningKey("sk_live_abc");
    expect(k1.toString("hex")).toBe(k2.toString("hex"));
  });

  it("differs for different secrets", () => {
    const k1 = deriveSigningKey("sk_live_abc");
    const k2 = deriveSigningKey("sk_live_xyz");
    expect(k1.toString("hex")).not.toBe(k2.toString("hex"));
  });
});

describe("signIdentity", () => {
  const secret = "sk_live_test_secret_for_unit_tests";

  it("produces a valid three-part JWT", () => {
    const token = signIdentity(secret, {
      sub: "u_1234",
      email: "alice@example.com",
      exp: Math.floor(Date.now() / 1000) + 600,
    });
    expect(token.split(".")).toHaveLength(3);
  });

  it("sets alg=HS256 and typ=JWT in header", () => {
    const token = signIdentity(secret, {
      sub: "u_1234",
      email: "alice@example.com",
      exp: Math.floor(Date.now() / 1000) + 600,
    });
    const { header } = decodeJwt(token);
    expect(header["alg"]).toBe("HS256");
    expect(header["typ"]).toBe("JWT");
  });

  it("includes sub, email, iat, exp in payload", () => {
    const exp = Math.floor(Date.now() / 1000) + 600;
    const token = signIdentity(secret, { sub: "u_42", email: "bob@example.com", exp });
    const { payload } = decodeJwt(token);
    expect(payload["sub"]).toBe("u_42");
    expect(payload["email"]).toBe("bob@example.com");
    expect(payload["exp"]).toBe(exp);
    expect(typeof payload["iat"]).toBe("number");
  });

  it("includes optional name, account_id, aud when provided", () => {
    const token = signIdentity(secret, {
      sub: "u_99",
      email: "carol@example.com",
      name: "Carol",
      account_id: "acme",
      aud: "pk_live_abc123",
      expiresIn: "10m",
    });
    const { payload } = decodeJwt(token);
    expect(payload["name"]).toBe("Carol");
    expect(payload["account_id"]).toBe("acme");
    expect(payload["aud"]).toBe("pk_live_abc123");
  });

  it("omits optional fields when not provided", () => {
    const token = signIdentity(secret, {
      sub: "u_1",
      email: "min@example.com",
      expiresIn: "5m",
    });
    const { payload } = decodeJwt(token);
    expect("name" in payload).toBe(false);
    expect("account_id" in payload).toBe(false);
    expect("aud" in payload).toBe(false);
  });

  it("produces a signature verifiable with the HKDF-derived key", () => {
    const exp = Math.floor(Date.now() / 1000) + 600;
    const token = signIdentity(secret, { sub: "u_1", email: "v@test.com", exp });
    const { signingInput, signature } = decodeJwt(token);

    const key = deriveSigningKey(secret);
    const expected = createHmac("sha256", key).update(signingInput).digest();
    expect(signature.toString("hex")).toBe(expected.toString("hex"));
  });

  it("parses expiresIn duration strings", () => {
    const before = Math.floor(Date.now() / 1000);
    const token = signIdentity(secret, {
      sub: "u_1",
      email: "d@test.com",
      expiresIn: "30m",
    });
    const { payload } = decodeJwt(token);
    const iat = payload["iat"] as number;
    const exp = payload["exp"] as number;
    expect(exp - iat).toBe(1800);
    expect(iat).toBeGreaterThanOrEqual(before);
  });

  it("rejects expiresIn > 1 hour", () => {
    expect(() =>
      signIdentity(secret, {
        sub: "u_1",
        email: "x@test.com",
        expiresIn: "2h",
      }),
    ).toThrow("max lifetime");
  });
});
