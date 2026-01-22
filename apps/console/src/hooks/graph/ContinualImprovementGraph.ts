import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "../useMutationWithToasts";

export const ContinualImprovementsConnectionKey
  = "ContinualImprovementsPage_continualImprovements";

export const continualImprovementsQuery = graphql`
  query ContinualImprovementGraphListQuery(
    $organizationId: ID!
    $snapshotId: ID
  ) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateContinualImprovement: permission(
          action: "core:continual-improvement:create"
        )
        ...ContinualImprovementsPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const continualImprovementNodeQuery = graphql`
  query ContinualImprovementGraphNodeQuery($continualImprovementId: ID!) {
    node(id: $continualImprovementId) {
      ... on ContinualImprovement {
        id
        snapshotId
        sourceId
        referenceId
        description
        source
        targetDate
        status
        priority
        owner {
          id
          fullName
        }
        organization {
          id
          name
        }
        createdAt
        updatedAt
        canUpdate: permission(action: "core:continual-improvement:update")
        canDelete: permission(action: "core:continual-improvement:delete")
      }
    }
  }
`;

export const createContinualImprovementMutation = graphql`
  mutation ContinualImprovementGraphCreateMutation(
    $input: CreateContinualImprovementInput!
    $connections: [ID!]!
  ) {
    createContinualImprovement(input: $input) {
      continualImprovementEdge @prependEdge(connections: $connections) {
        node {
          id
          referenceId
          description
          source
          targetDate
          status
          priority
          owner {
            id
            fullName
          }
          createdAt
          canUpdate: permission(action: "core:continual-improvement:update")
          canDelete: permission(action: "core:continual-improvement:delete")
        }
      }
    }
  }
`;

export const updateContinualImprovementMutation = graphql`
  mutation ContinualImprovementGraphUpdateMutation(
    $input: UpdateContinualImprovementInput!
  ) {
    updateContinualImprovement(input: $input) {
      continualImprovement {
        id
        referenceId
        description
        source
        targetDate
        status
        priority
        owner {
          id
          fullName
        }
        updatedAt
      }
    }
  }
`;

export const deleteContinualImprovementMutation = graphql`
  mutation ContinualImprovementGraphDeleteMutation(
    $input: DeleteContinualImprovementInput!
    $connections: [ID!]!
  ) {
    deleteContinualImprovement(input: $input) {
      deletedContinualImprovementId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteContinualImprovement = (
  improvement: { id: string; referenceId: string },
  connectionId: string,
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteContinualImprovementMutation, {
    successMessage: __("Continual improvement deleted successfully"),
    errorMessage: __("Failed to delete continual improvement"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              continualImprovementId: improvement.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the continual improvement %s. This action cannot be undone.",
          ),
          improvement.referenceId,
        ),
      },
    );
  };
};

export const useCreateContinualImprovement = (connectionId: string) => {
  const [mutate] = useMutation(createContinualImprovementMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    referenceId: string;
    description?: string;
    source?: string;
    ownerId: string;
    targetDate?: string;
    status: string;
    priority: string;
  }) => {
    if (!input.organizationId) {
      return alert(
        __("Failed to create continual improvement: organization is required"),
      );
    }
    if (!input.referenceId) {
      return alert(
        __("Failed to create continual improvement: reference ID is required"),
      );
    }
    if (!input.ownerId) {
      return alert(
        __("Failed to create continual improvement: owner is required"),
      );
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          referenceId: input.referenceId,
          description: input.description,
          source: input.source,
          ownerId: input.ownerId,
          targetDate: input.targetDate,
          status: input.status || "OPEN",
          priority: input.priority || "MEDIUM",
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateContinualImprovement = () => {
  const [mutate] = useMutation(updateContinualImprovementMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    referenceId?: string;
    description?: string;
    source?: string;
    ownerId?: string;
    targetDate?: string | null;
    status?: string;
    priority?: string;
  }) => {
    if (!input.id) {
      return alert(
        __("Failed to update continual improvement: ID is required"),
      );
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
