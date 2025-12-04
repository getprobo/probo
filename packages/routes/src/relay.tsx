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
