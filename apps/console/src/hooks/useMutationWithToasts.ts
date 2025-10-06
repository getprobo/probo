import { useCallback } from "react";
import { useMutation, type UseMutationConfig } from "react-relay";
import { useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { MutationParameters, GraphQLTaggedNode } from "relay-runtime";
import { formatError, type GraphQLError } from "@probo/helpers";

/**
 * A decorated useMutation hook that emits toast notifications on success or error.
 */
export function useMutationWithToasts<T extends MutationParameters>(
  query: GraphQLTaggedNode,
  baseOptions?: {
    successMessage?: string | ((response: T["response"]) => string);
    errorMessage?: string;
  }
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
      }
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
                description: formatError(errorTitle, error as GraphQLError),
                variant: "error",
              });
              reject(error);
              return;
            }
            const successMessage = typeof options.successMessage === "function"
              ? options.successMessage(response)
              : options.successMessage;

            toast({
              title: __("Success"),
              description:
                successMessage ??
                __("Operation completed successfully"),
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
        })
      );
    },
    [mutate]
  );

  return [mutateWithToast, isLoading] as const;
}
