import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const SnapshotsConnectionKey = "SnapshotsPage_snapshots";

export const snapshotsQuery = graphql`
  query SnapshotGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...SnapshotsPageFragment
      }
    }
  }
`;

export const snapshotNodeQuery = graphql`
  query SnapshotGraphNodeQuery($snapshotId: ID!) {
    node(id: $snapshotId) {
      ... on Snapshot {
        id
        name
        description
        type
        organization {
          id
          name
        }
        createdAt
      }
    }
  }
`;

export const createSnapshotMutation = graphql`
  mutation SnapshotGraphCreateMutation(
    $input: CreateSnapshotInput!
    $connections: [ID!]!
  ) {
    createSnapshot(input: $input) {
      snapshotEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          type
          createdAt
        }
      }
    }
  }
`;

export const deleteSnapshotMutation = graphql`
  mutation SnapshotGraphDeleteMutation(
    $input: DeleteSnapshotInput!
    $connections: [ID!]!
  ) {
    deleteSnapshot(input: $input) {
      deletedSnapshotId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteSnapshot = (
  snapshot: { id: string; name: string },
  connectionId: string
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteSnapshotMutation, {
    successMessage: __("Snapshot deleted successfully"),
    errorMessage: __("Failed to delete snapshot"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              snapshotId: snapshot.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the snapshot %s. This action cannot be undone."
          ),
          snapshot.name
        ),
      }
    );
  };
};

export const useCreateSnapshot = (connectionId: string) => {
  const [mutate] = useMutation(createSnapshotMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    name: string;
    description?: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create snapshot: organization is required"));
    }
    if (!input.name) {
      return alert(__("Failed to create snapshot: name is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          name: input.name,
          description: input.description,
        },
        connections: [connectionId],
      },
    });
  };
};
