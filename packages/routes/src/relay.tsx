// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { useCleanup } from "@probo/hooks";
import { type ComponentType } from "react";
import type { EnvironmentProviderOptions, PreloadedQuery } from "react-relay";
import { type LoaderFunction, type LoaderFunctionArgs, useLoaderData } from "react-router";
import { type OperationType } from "relay-runtime";

/**
 * @deprecated Use a `*PageLoader` component with `useQueryLoader` +
 * `usePreloadedQuery` instead. See contrib/claude/relay.md.
 *
 * Infer the concrete `queryRef` type from a naked type position. Relay 21's
 * first-party types model `PreloadedQuery#variables` as `VariablesOf<TQuery>`,
 * which prevents inferring `TQuery` through it, so we infer the whole queryRef.
 */
export function withQueryRef<
  TQueryRef extends PreloadedQuery<OperationType>,
>(
  Component: ComponentType<{ queryRef: TQueryRef }>,
) {
  return function WithQueryRef() {
    // `useLoaderData` is typed `any` (default generic), and its `SerializeFrom`
    // generic would strip the `dispose` function type. Assert the loader's
    // shape so the rest of the component stays type-safe; the assertion is not
    // redundant despite the rule flagging it (the source is `any`).
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
    const { queryRef, dispose } = useLoaderData() as {
      queryRef: TQueryRef;
      dispose: () => void;
    };

    useCleanup(dispose, 1000);

    return <Component queryRef={queryRef} />;
  };
}

/**
 * @deprecated Use a `*PageLoader` component with `useQueryLoader` +
 * `usePreloadedQuery` instead. See contrib/claude/relay.md.
 */
export function loaderFromQueryLoader<
  TQuery extends OperationType,
  TEnvironmentProviderOptions = EnvironmentProviderOptions,
>(
  queryLoader: (params: Record<string, string>) => PreloadedQuery<TQuery, TEnvironmentProviderOptions>,
): LoaderFunction {
  return ({ params }: LoaderFunctionArgs) => {
    const query = queryLoader(params as Record<string, string>);
    return {
      queryRef: query,
      dispose: query.dispose,
    };
  };
}
