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
import { useMutation, type UseMutationConfig } from "react-relay";
import type { GraphQLTaggedNode, MutationParameters } from "relay-runtime";

/**
 * A decorated useMutation hook that emits toast notifications on success or error.
 */
export function useMutationWithToasts<T extends MutationParameters>(
  query: GraphQLTaggedNode,
  baseOptions?: {
    successMessage?: string | ((response: T["response"]) => string);
    errorMessage?: string;
  },
) {
  const [mutate, isLoading] = useMutation<T>(query);
  const { toast } = useToast();
  const { __ } = useTranslate();
  const mutateWithToast = useCallback(
    (
      queryOptions: UseMutationConfig<T> & {
        onSuccess?: () => void;
        successMessage?: string | ((response: T["response"]) => string);
        errorMessage?: string;
      },
    ) => {
      const options = { ...baseOptions, ...queryOptions };
      return new Promise<void>((resolve, reject) =>
        mutate({
          ...queryOptions,
          onCompleted: (response, error) => {
            options.onCompleted?.(response, error);
            if (error) {
              const errorTitle = options.errorMessage ?? __("Failed to commit this operation");
              toast({
                title: __("Error"),
                description: formatError(errorTitle, error as GraphQLError[]),
                variant: "error",
              });
              reject(new Error(errorTitle));
              return;
            }
            const successMessage = typeof options.successMessage === "function"
              ? options.successMessage(response)
              : options.successMessage;

            toast({
              title: __("Success"),
              description:
                successMessage
                ?? __("Operation completed successfully"),
              variant: "success",
            });
            options.onSuccess?.();
            resolve();
          },
          onError: (error) => {
            const errorTitle = options.errorMessage ?? __("Failed to commit this operation");
            toast({
              title: __("Error"),
              description: formatError(errorTitle, error as GraphQLError),
              variant: "error",
            });
            reject(error);
          },
        }),
      );
    },
    [mutate, toast, __, baseOptions],
  );

  return [mutateWithToast, isLoading] as const;
}
