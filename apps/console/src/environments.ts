// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { makeFetchQuery } from "@probo/relay";
import { Environment, Network, RecordSource, Store } from "relay-runtime";

export const coreEnvironment = new Environment({
  configName: "core",
  network: Network.create(makeFetchQuery("/api/console/v1/graphql")),
  store: new Store(new RecordSource(), {
    queryCacheExpirationTime: 1 * 60 * 1000,
    gcReleaseBufferSize: 20,
  }),
});

export const iamEnvironment = new Environment({
  configName: "iam",
  network: Network.create(makeFetchQuery("/api/connect/v1/graphql")),
  store: new Store(new RecordSource(), {
    queryCacheExpirationTime: 1 * 60 * 1000,
    gcReleaseBufferSize: 20,
  }),
});
