import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "../useMutationWithToasts";

export const NonconformitiesConnectionKey
  = "NonconformitiesPage_nonconformities";

export const nonconformitiesQuery = graphql`
  query NonconformityGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateNonconformity: permission(action: "core:nonconformity:create")
        ...NonconformitiesPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const nonconformityNodeQuery = graphql`
  query NonconformityGraphNodeQuery($nonconformityId: ID!) {
    node(id: $nonconformityId) {
      ... on Nonconformity {
        id
        snapshotId
        referenceId
        description
        dateIdentified
        rootCause
        correctiveAction
        dueDate
        status
        effectivenessCheck
        audit {
          id
          framework {
            id
            name
          }
        }
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
        canUpdate: permission(action: "core:nonconformity:update")
        canDelete: permission(action: "core:nonconformity:delete")
      }
    }
  }
`;

export const createNonconformityMutation = graphql`
  mutation NonconformityGraphCreateMutation(
    $input: CreateNonconformityInput!
    $connections: [ID!]!
  ) {
    createNonconformity(input: $input) {
      nonconformityEdge @prependEdge(connections: $connections) {
        node {
          id
          referenceId
          description
          status
          dateIdentified
          dueDate
          rootCause
          audit {
            id
            framework {
              name
            }
          }
          owner {
            id
            fullName
          }
          createdAt
          canUpdate: permission(action: "core:nonconformity:update")
          canDelete: permission(action: "core:nonconformity:delete")
        }
      }
    }
  }
`;

export const updateNonconformityMutation = graphql`
  mutation NonconformityGraphUpdateMutation($input: UpdateNonconformityInput!) {
    updateNonconformity(input: $input) {
      nonconformity {
        id
        referenceId
        description
        dateIdentified
        rootCause
        correctiveAction
        dueDate
        status
        effectivenessCheck
        owner {
          id
          fullName
        }
        audit {
          id
          framework {
            id
            name
          }
        }
        updatedAt
      }
    }
  }
`;

export const deleteNonconformityMutation = graphql`
  mutation NonconformityGraphDeleteMutation(
    $input: DeleteNonconformityInput!
    $connections: [ID!]!
  ) {
    deleteNonconformity(input: $input) {
      deletedNonconformityId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteNonconformity = (
  nonconformity: { id: string; referenceId: string },
  connectionId: string,
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteNonconformityMutation, {
    successMessage: __("Nonconformity deleted successfully"),
    errorMessage: __("Failed to delete nonconformity"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              nonconformityId: nonconformity.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the nonconformity %s. This action cannot be undone.",
          ),
          nonconformity.referenceId,
        ),
      },
    );
  };
};

export const useCreateNonconformity = (connectionId: string) => {
  const [mutate] = useMutation(createNonconformityMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    referenceId: string;
    description?: string;
    auditId?: string;
    dateIdentified?: string;
    rootCause: string;
    correctiveAction?: string;
    ownerId: string;
    dueDate?: string;
    status: string;
    effectivenessCheck?: string;
  }) => {
    if (!input.organizationId) {
      return alert(
        __("Failed to create nonconformity: organization is required"),
      );
    }
    if (!input.referenceId) {
      return alert(
        __("Failed to create nonconformity: reference ID is required"),
      );
    }
    if (!input.ownerId) {
      return alert(__("Failed to create nonconformity: owner is required"));
    }
    if (!input.rootCause) {
      return alert(
        __("Failed to create nonconformity: root cause is required"),
      );
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          referenceId: input.referenceId,
          description: input.description,
          auditId: input.auditId || undefined,
          dateIdentified: input.dateIdentified,
          rootCause: input.rootCause,
          correctiveAction: input.correctiveAction,
          ownerId: input.ownerId,
          dueDate: input.dueDate,
          status: input.status || "OPEN",
          effectivenessCheck: input.effectivenessCheck,
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateNonconformity = () => {
  const [mutate] = useMutation(updateNonconformityMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    referenceId?: string;
    description?: string;
    dateIdentified?: string | null;
    rootCause?: string;
    correctiveAction?: string;
    ownerId?: string;
    auditId?: string | null;
    dueDate?: string | null;
    status?: string;
    effectivenessCheck?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update nonconformity: ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          ...input,
          auditId: input.auditId || null,
        },
      },
    });
  };
};
