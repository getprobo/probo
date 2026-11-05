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

import { bridgeCanResume, bridgeStatus } from "./scimProvider";

describe("bridgeCanResume", () => {
  it("allows resume when sync is on hold or failing", () => {
    expect(bridgeCanResume("DISABLED")).toBe(true);
    expect(bridgeCanResume("FAILED")).toBe(true);
  });

  it("hides resume while the bridge is syncing or connected", () => {
    expect(bridgeCanResume("ACTIVE")).toBe(false);
    expect(bridgeCanResume("SYNCING")).toBe(false);
    expect(bridgeCanResume("PENDING")).toBe(false);
  });
});

describe("bridgeStatus", () => {
  it("maps a held bridge to the on-hold status", () => {
    expect(bridgeStatus("DISABLED")).toBe("disabled");
    expect(bridgeStatus("FAILED")).toBe("error");
    expect(bridgeStatus("ACTIVE")).toBe("connected");
  });
});
