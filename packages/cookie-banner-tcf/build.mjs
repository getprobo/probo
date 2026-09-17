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
import * as esbuild from "esbuild";

const { version } = JSON.parse(readFileSync("./package.json", "utf-8"));

const shared = {
  bundle: true,
  target: "es2020",
  define: {
    __SDK_VERSION__: JSON.stringify(version),
  },
};

await Promise.all([
  esbuild.build({
    ...shared,
    entryPoints: ["src/index.ts"],
    outfile: "dist/cookie-banner-tcf.mjs",
    format: "esm",
    external: ["@probo/cookie-banner"],
  }),
  esbuild.build({
    ...shared,
    entryPoints: ["src/iife.ts"],
    outfile: "dist/cookie-banner-tcf.iife.js",
    format: "iife",
    globalName: "ProboCookieBannerTCF",
    minify: true,
  }),
]);
