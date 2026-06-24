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

import { COOKIE_NAME } from "./cookie";

export function getVisitorId(bannerId: string): string | null {
  const key = `${COOKIE_NAME}:${bannerId}:vid`;

  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function getOrCreateVisitorId(bannerId: string): string {
  const existing = getVisitorId(bannerId);
  if (existing) {
    return existing;
  }

  const key = `${COOKIE_NAME}:${bannerId}:vid`;
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  const id = Array.from(array, (b) => b.toString(16).padStart(2, "0")).join("");

  try {
    localStorage.setItem(key, id);
  } catch {
    // localStorage unavailable
  }

  return id;
}
