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
                description: formatError(errorTitle, error),
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
              description: formatError(errorTitle, error),
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
