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
  type FaroSourcemapEnv,
  faroSourcemapUpload,
} from "./vite";

const app = { appName: "console", appId: "app-1" };

const complete: FaroSourcemapEnv = {
  FARO_SOURCEMAP_ENDPOINT: "https://faro.example/api",
  FARO_SOURCEMAP_API_KEY: "token",
  FARO_STACK_ID: "stack",
  GITHUB_SHA: "0123456789abcdef0123456789abcdef01234567",
};

describe("faroSourcemapUpload", () => {
  it("leaves sourcemaps off when any credential is missing", () => {
    expect(faroSourcemapUpload(app, {}).plugins).toEqual([]);
    expect(faroSourcemapUpload(app, {}).sourcemap).toBe(false);
    expect(faroSourcemapUpload({ ...app, appId: "  " }, complete).plugins).toEqual([]);
    expect(faroSourcemapUpload(app, {
      ...complete,
      FARO_SOURCEMAP_API_KEY: "",
    }).plugins).toEqual([]);
    expect(faroSourcemapUpload(app, {
      ...complete,
      FARO_SOURCEMAP_ENDPOINT: "  ",
    }).plugins).toEqual([]);
    expect(faroSourcemapUpload(app, {
      ...complete,
      FARO_STACK_ID: undefined,
    }).plugins).toEqual([]);
  });

  it("enables hidden sourcemaps when every credential is set", () => {
    const upload = faroSourcemapUpload(
      {
        appName: "employee-portal",
        appId: "  app-2  ",
        prefixPath: "employee-portal/",
      },
      {
        ...complete,
        FARO_SOURCEMAP_ENDPOINT: "  https://faro.example/api  ",
      },
    );

    expect(upload.sourcemap).toBe("hidden");
    expect(upload.plugins).toHaveLength(1);
  });
});
