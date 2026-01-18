import { graphql, useMutation } from "react-relay";
import { useConfirm, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import {
    promisifyMutation,
    sprintf,
    formatError,
    type GraphQLError,
} from "@probo/helpers";
import type { useDeleteStateOfApplicabilityMutation } from "/__generated__/core/useDeleteStateOfApplicabilityMutation.graphql";

const deleteStateOfApplicabilityMutation = graphql`
    mutation useDeleteStateOfApplicabilityMutation(
        $input: DeleteStateOfApplicabilityInput!
        $connections: [ID!]!
    ) {
        deleteStateOfApplicability(input: $input) {
            deletedStateOfApplicabilityId @deleteEdge(connections: $connections)
        }
    }
`;

export const StateOfApplicabilityConnectionKey =
    "StatesOfApplicabilityPage_statesOfApplicability";

export function useDeleteStateOfApplicability(
    stateOfApplicability: { id: string; name: string },
    connectionId: string,
    onSuccess?: () => void,
) {
    const [mutate] = useMutation<useDeleteStateOfApplicabilityMutation>(
        deleteStateOfApplicabilityMutation,
    );
    const confirm = useConfirm();
    const { toast } = useToast();
    const { __ } = useTranslate();

    return () => {
        confirm(
            () =>
                promisifyMutation(mutate)({
                    variables: {
                        input: {
                            stateOfApplicabilityId: stateOfApplicability.id,
                        },
                        connections: [connectionId],
                    },
                })
                    .then(() => {
                        onSuccess?.();
                    })
                    .catch((error) => {
                        toast({
                            title: __("Error"),
                            description: formatError(
                                __("Failed to delete state of applicability"),
                                error as GraphQLError,
                            ),
                            variant: "error",
                        });
                    }),
            {
                message: sprintf(
                    __(
                        'This will permanently delete "%s". This action cannot be undone.',
                    ),
                    stateOfApplicability.name,
                ),
            },
        );
    };
}

