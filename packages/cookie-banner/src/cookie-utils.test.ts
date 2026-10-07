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

import { cookieListItemDomain, parseCookieSetDomain } from "./cookie-utils";

describe("parseCookieSetDomain", () => {
  it("treats a missing Domain attribute as host-only", () => {
    expect(parseCookieSetDomain("sid=abc; path=/; max-age=3600")).toEqual({
      host_only: true,
    });
  });

  it("strips a leading dot and lowercases the Domain attribute", () => {
    expect(parseCookieSetDomain("sid=abc; Domain=.Example.COM")).toEqual({
      cookie_domain: "example.com",
      host_only: false,
    });
  });

  it("treats an empty Domain attribute as host-only", () => {
    expect(parseCookieSetDomain("sid=abc; Domain=")).toEqual({
      host_only: true,
    });
  });
});

describe("cookieListItemDomain", () => {
  it("treats a null domain as host-only", () => {
    expect(cookieListItemDomain(null)).toEqual({ host_only: true });
  });

  it("normalizes a Cookie Store Domain attribute", () => {
    expect(cookieListItemDomain(".Example.COM")).toEqual({
      cookie_domain: "example.com",
      host_only: false,
    });
  });
});
