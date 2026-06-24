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

import { formatError } from "@probo/helpers";
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
              description: formatError(errorTitle, error),
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
            description: formatError(errorTitle, error),
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
    const previousValue: unknown = node?.getValue(fieldName);
    if (node && typeof previousValue === "number") {
      node.setValue(previousValue + value, fieldName);
    }
  });
}
