import { useCallback } from "react";
import { useMutation, type UseMutationConfig } from "react-relay";
import { useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { MutationParameters, GraphQLTaggedNode } from "relay-runtime";

/**
 * A decorated useMutation hook that emits toast notifications on success or error.
 */
export function useMutationWithToasts<T extends MutationParameters>(
  query: GraphQLTaggedNode
) {
  const [mutate, isLoading] = useMutation<T>(query);
  const { toast } = useToast();
  const { __ } = useTranslate();
  const mutateWithToast = useCallback(
    (
      options: UseMutationConfig<T> & {
        onSuccess?: () => void;
        successMessage: string;
        errorMessage?: string;
      }
    ) => {
      mutate({
        ...options,
        onCompleted: (_, error) => {
          if (error) {
            toast({
              title: __("Error"),
              description:
                options.errorMessage ?? __("Failed to commit this operation."),
              variant: "error",
            });
            return;
          }
          toast({
            title: __("Success"),
            description: options.successMessage,
            variant: "success",
          });
          options.onSuccess?.();
        },
        onError: () => {
          toast({
            title: __("Error"),
            description:
              options.errorMessage ?? __("Failed to commit this operation."),
            variant: "error",
          });
        },
      });
    },
    [mutate]
  );

  return [mutateWithToast, isLoading] as const;
}
