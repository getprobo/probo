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

import { type ComponentType } from "react";
import type { EnvironmentProviderOptions, PreloadedQuery } from "react-relay";
import { type LoaderFunction, type LoaderFunctionArgs, useLoaderData } from "react-router";
import { type OperationType } from "relay-runtime";
import { useCleanup } from "@probo/hooks";

export function withQueryRef<
  TQuery extends OperationType,
  TEnvironmentProviderOptions = EnvironmentProviderOptions
>(
  Component: ComponentType<{ queryRef: PreloadedQuery<TQuery, TEnvironmentProviderOptions> }>,
) {
  return () => {
    const { queryRef, dispose } = useLoaderData();

    useCleanup(dispose, 1000);

    return <Component queryRef={queryRef} />
  }
}

export function loaderFromQueryLoader<
  TQuery extends OperationType,
  TEnvironmentProviderOptions = EnvironmentProviderOptions
>(
  queryLoader: (params: Record<string, string>) => PreloadedQuery<TQuery, TEnvironmentProviderOptions>
): LoaderFunction {
  return ({ params }: LoaderFunctionArgs) => {
    const query = queryLoader(params as Record<string, string>);
    return {
      queryRef: query,
      dispose: query.dispose,
    };
  }
}
