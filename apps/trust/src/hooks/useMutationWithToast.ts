import { useCallback } from "react";
import { useMutation, type UseMutationConfig } from "react-relay";
import { useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { MutationParameters, GraphQLTaggedNode } from "relay-runtime";

/**
 * A decorated useMutation hook that emits toast notifications on success or error.
 */
export function useMutationWithToasts<T extends MutationParameters>(
  query: GraphQLTaggedNode,
  baseOptions?: {
    onSuccess?: (response: T["response"]) => void;
    errorMessage?: string;
  }
) {
  const [mutate, isLoading] = useMutation<T>(query);
  const { toast } = useToast();
  const { __ } = useTranslate();
  const mutateWithToast = useCallback(
    (
      queryOptions: UseMutationConfig<T> & {
        onSuccess?: (response: T["response"]) => void;
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
              toast({
                title: __("Error"),
                description:
                  options.errorMessage ??
                  __("Failed to commit this operation."),
                variant: "error",
              });
              reject(error);
              return;
            }
            options.onSuccess?.(response);
            resolve();
          },
          onError: (error) => {
            toast({
              title: __("Error"),
              description:
                options.errorMessage ?? __("Failed to commit this operation."),
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
