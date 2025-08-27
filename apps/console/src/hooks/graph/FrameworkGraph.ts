import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts";
import { useCallback } from "react";
import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";

export const connectionListKey = "FrameworksListQuery_frameworks";

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
  const [commitDelete] = useMutationWithToasts(deleteFrameworkMutation, {
    errorMessage: "Failed to delete framework",
    successMessage: "Framework deleted successfully",
  });
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return useCallback(
    (options?: { onSuccess?: () => void }) => {
      return confirm(
        () => {
          return commitDelete({
            variables: {
              input: {
                frameworkId: framework.id!,
              },
              connections: [connectionId],
            },
            ...options,
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
    },
    [framework, connectionId, commitDelete]
  );
};

export const frameworkNodeQuery = graphql`
  query FrameworkGraphNodeQuery($frameworkId: ID!) {
    node(id: $frameworkId) {
      ... on Framework {
        id
        name
        ...FrameworkDetailPageFragment
      }
    }
  }
`;

export const frameworkControlNodeQuery = graphql`
  query FrameworkGraphControlNodeQuery($controlId: ID!) {
    node(id: $controlId) {
      ... on Control {
        id
        name
        sectionTitle
        description
        status
        exclusionJustification
        ...FrameworkControlDialogFragment
        measures(first: 100)
          @connection(key: "FrameworkGraphControl_measures") {
          __id
          edges {
            node {
              id
              ...LinkedMeasuresCardFragment
            }
          }
        }
        documents(first: 100)
          @connection(key: "FrameworkGraphControl_documents") {
          __id
          edges {
            node {
              id
              ...LinkedDocumentsCardFragment
            }
          }
        }
        audits(first: 100)
          @connection(key: "FrameworkGraphControl_audits") {
          __id
          edges {
            node {
              id
              ...LinkedAuditsCardFragment
            }
          }
        }
        snapshots(first: 100)
          @connection(key: "FrameworkGraphControl_snapshots") {
          __id
          edges {
            node {
              id
              ...LinkedSnapshotsCardFragment
            }
          }
        }
      }
    }
  }
`;
