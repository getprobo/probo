import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const ObligationsConnectionKey = "ObligationsPage_obligations";

export const obligationsQuery = graphql`
  query ObligationGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        ...ObligationsPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const obligationNodeQuery = graphql`
  query ObligationGraphNodeQuery($obligationId: ID!) {
    node(id: $obligationId) {
      ... on Obligation {
        id
        snapshotId
        sourceId
        referenceId
        area
        source
        requirement
        actionsToBeImplemented
        regulator
        lastReviewDate
        dueDate
        status
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
      }
    }
  }
`;

export const createObligationMutation = graphql`
  mutation ObligationGraphCreateMutation(
    $input: CreateObligationInput!
    $connections: [ID!]!
  ) {
    createObligation(input: $input) {
      obligationEdge @prependEdge(connections: $connections) {
        node {
          id
          referenceId
          area
          source
          requirement
          actionsToBeImplemented
          regulator
          lastReviewDate
          dueDate
          status
          owner {
            id
            fullName
          }
          createdAt
        }
      }
    }
  }
`;

export const updateObligationMutation = graphql`
  mutation ObligationGraphUpdateMutation($input: UpdateObligationInput!) {
    updateObligation(input: $input) {
      obligation {
        id
        referenceId
        area
        source
        requirement
        actionsToBeImplemented
        regulator
        lastReviewDate
        dueDate
        status
        owner {
          id
          fullName
        }
        updatedAt
      }
    }
  }
`;

export const deleteObligationMutation = graphql`
  mutation ObligationGraphDeleteMutation(
    $input: DeleteObligationInput!
    $connections: [ID!]!
  ) {
    deleteObligation(input: $input) {
      deletedObligationId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteObligation = (
  obligation: { id: string; referenceId: string },
  connectionId: string
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteObligationMutation, {
    successMessage: __("Obligation deleted successfully"),
    errorMessage: __("Failed to delete obligation"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              obligationId: obligation.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the obligation %s. This action cannot be undone."
          ),
          obligation.referenceId
        ),
      }
    );
  };
};

export const useCreateObligation = (connectionId: string) => {
  const [mutate] = useMutation(createObligationMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    referenceId: string;
    area?: string;
    source?: string;
    requirement?: string;
    actionsToBeImplemented?: string;
    regulator?: string;
    ownerId: string;
    lastReviewDate?: string;
    dueDate?: string;
    status: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create obligation: organization is required"));
    }
    if (!input.referenceId) {
      return alert(__("Failed to create obligation: reference ID is required"));
    }
    if (!input.ownerId) {
      return alert(__("Failed to create obligation: owner is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          referenceId: input.referenceId,
          area: input.area,
          source: input.source,
          requirement: input.requirement,
          actionsToBeImplemented: input.actionsToBeImplemented,
          regulator: input.regulator,
          ownerId: input.ownerId,
          lastReviewDate: input.lastReviewDate,
          dueDate: input.dueDate,
          status: input.status || "OPEN",
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateObligation = () => {
  const [mutate] = useMutation(updateObligationMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    referenceId?: string;
    area?: string;
    source?: string;
    requirement?: string;
    actionsToBeImplemented?: string;
    regulator?: string;
    ownerId?: string;
    lastReviewDate?: string;
    dueDate?: string;
    status?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update obligation: ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
