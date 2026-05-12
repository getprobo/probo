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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useCallback } from "react";
import {
  useMutation,
  type UseMutationConfig,
  useRelayEnvironment,
} from "react-relay";
import {
  commitLocalUpdate,
  type Environment,
  type GraphQLTaggedNode,
  type MutationParameters,
} from "relay-runtime";

const defaultOptions = {
  field: "totalCount",
  value: 1,
};

/**
 * A decorated useMutation hook that increments the store on complete.
 */
export function useMutationWithIncrement<T extends MutationParameters>(
  query: GraphQLTaggedNode,
  baseOptions: {
    id: string;
    node: string;
    field?: string;
    value?: 1 | -1;
    errorMessage?: string;
  },
) {
  const [mutate, isLoading] = useMutation<T>(query);
  const relayEnv = useRelayEnvironment();
  const { toast } = useToast();
  const { __ } = useTranslate();
  const options = { ...defaultOptions, ...baseOptions };
  const mutateAndIncrement = useCallback(
    (queryOptions: UseMutationConfig<T>) => {
      return mutate({
        ...queryOptions,
        onCompleted: (response, error) => {
          if (error) {
            const errorTitle = options.errorMessage ?? __("Failed to commit this operation");
            toast({
              title: __("Error"),
              description: formatError(errorTitle, error as GraphQLError[]),
              variant: "error",
            });
          } else {
            updateStoreCounter(
              relayEnv,
              options.id,
              options.node,
              options.value,
              options.field,
            );
          }
          queryOptions.onCompleted?.(response, error);
        },
        onError: (error) => {
          const errorTitle = options.errorMessage ?? __("Failed to commit this operation");
          toast({
            title: __("Error"),
            description: formatError(errorTitle, error as GraphQLError),
            variant: "error",
          });
          queryOptions.onError?.(error);
        },
      });
    },
    [mutate, options.id, options.node, options.field, options.value, options.errorMessage, relayEnv, toast, __],
  );

  return [mutateAndIncrement, isLoading] as const;
}

export function updateStoreCounter(
  relayEnv: Environment,
  recordId: string,
  nodeName: string,
  value: number = 1,
  fieldName: string = "totalCount",
) {
  commitLocalUpdate(relayEnv, (store) => {
    const node = store?.get(recordId)?.getLinkedRecord(nodeName);
    const previousValue = node?.getValue(fieldName);
    if (node && typeof previousValue === "number") {
      node.setValue(previousValue + value, fieldName);
    }
  });
}
