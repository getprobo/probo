import { useMutation } from "react-relay";
import type { GraphQLTaggedNode } from "relay-runtime";

export type MutationFieldUpdate<T extends Record<string, unknown>, TKey extends keyof T> = (
  field: TKey,
  value: T[TKey],
) => void;

/**
 * Mutate a single field from a graphql mutation
 */
export function useMutateField<Input extends Record<string, unknown>>(
  mutation: GraphQLTaggedNode,
) {
  const [mutate, isUpdating] = useMutation(mutation);

  return {
    update: <T extends keyof Input>(id: string, fieldName: T, value: Input[T]) => {
      if (!id) {
        return;
      }
      mutate({
        variables: {
          input: {
            id: id,
            [fieldName]: value,
          },
        },
      });
    },
    isUpdating,
  };
}
