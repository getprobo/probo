// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
