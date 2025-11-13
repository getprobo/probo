import type { GraphQLTaggedNode } from "relay-runtime";
import { useMutation } from "react-relay";

export type MutationFieldUpdate<T> = (
  field: keyof T,
  value: T[typeof field],
) => void;

/**
 * Mutate a single field from a graphql mutation
 */
export function useMutateField<Input extends Record<string, unknown>>(
  mutation: GraphQLTaggedNode,
) {
  const [mutate, isUpdating] = useMutation(mutation);

  return {
    update<T extends keyof Input>(id: string, fieldName: T, value: Input[T]) {
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
