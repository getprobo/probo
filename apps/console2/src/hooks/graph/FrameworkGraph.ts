import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts";
import { useCallback } from "react";
import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";

export const frameworksQuery = graphql`
  query FrameworkGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        frameworks(first: 100)
          @connection(key: "FrameworksListQuery_frameworks") {
          __id
          edges {
            node {
              id
              ...FrameworksPageCardFragment
            }
          }
        }
      }
    }
  }
`;

const deleteFrameworkMutation = graphql`
  mutation FrameworkGraphDeleteMutation(
    $input: DeleteFrameworkInput!
    $connections: [ID!]!
  ) {
    deleteFramework(input: $input) {
      deletedFrameworkId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteFrameworkMutation = (
  framework: { id: string; name: string },
  connectionId: string
) => {
  const [commitDelete, isDeleting] = useMutationWithToasts(
    deleteFrameworkMutation,
    {
      errorMessage: "Failed to delete framework",
      successMessage: "Framework deleted successfully",
    }
  );
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return useCallback(() => {
    confirm(
      () => {
        return commitDelete({
          variables: {
            input: {
              frameworkId: framework.id!,
            },
            connections: [connectionId],
          },
        });
      },
      {
        message: sprintf(
          __(
            'This will permanently delete framework "%s". This action cannot be undone.'
          ),
          framework.name
        ),
      }
    );
  }, [framework, connectionId, commitDelete]);
};
