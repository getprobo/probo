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

import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { deriveSigningKey, signIdentity } from "./index";

interface Vector {
  description: string;
  org_secret: string;
  claims: {
    sub: string;
    email: string;
    name?: string;
    account_id?: string;
    aud?: string;
    iat: number;
    exp: number;
  };
  derived_key_hex: string;
  jwt: string;
}

interface VectorFile {
  vectors: Vector[];
}

const { vectors } = JSON.parse(
  readFileSync(join(__dirname, "../../test-vectors.json"), "utf-8"),
) as VectorFile;

describe("test-vectors.json", () => {
  for (const v of vectors) {
    it(v.description, () => {
      const key = deriveSigningKey(v.org_secret);
      expect(key.toString("hex")).toBe(v.derived_key_hex);

      const jwt = signIdentity(v.org_secret, {
        sub: v.claims.sub,
        email: v.claims.email,
        iat: v.claims.iat,
        exp: v.claims.exp,
        ...(v.claims.name !== undefined && { name: v.claims.name }),
        ...(v.claims.account_id !== undefined && { account_id: v.claims.account_id }),
        ...(v.claims.aud !== undefined && { aud: v.claims.aud }),
      });

      expect(jwt).toBe(v.jwt);
    });
  }
});
