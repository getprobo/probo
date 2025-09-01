import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const RegistriesConnectionKey = "RegistriesPage_nonconformityRegistries";

export const nonconformityRegistriesQuery = graphql`
  query NonconformityRegistryGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        ...NonconformityRegistriesPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const nonconformityRegistryNodeQuery = graphql`
  query NonconformityRegistryGraphNodeQuery($nonconformityRegistryId: ID!) {
    node(id: $nonconformityRegistryId) {
      ... on NonconformityRegistry {
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
      }
    }
  }
`;

export const createNonconformityRegistryMutation = graphql`
  mutation NonconformityRegistryGraphCreateMutation(
    $input: CreateNonconformityRegistryInput!
    $connections: [ID!]!
  ) {
    createNonconformityRegistry(input: $input) {
      nonconformityRegistryEdge @prependEdge(connections: $connections) {
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
        }
      }
    }
  }
`;

export const updateNonconformityRegistryMutation = graphql`
  mutation NonconformityRegistryGraphUpdateMutation($input: UpdateNonconformityRegistryInput!) {
    updateNonconformityRegistry(input: $input) {
      nonconformityRegistry {
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

export const deleteNonconformityRegistryMutation = graphql`
  mutation NonconformityRegistryGraphDeleteMutation(
    $input: DeleteNonconformityRegistryInput!
    $connections: [ID!]!
  ) {
    deleteNonconformityRegistry(input: $input) {
      deletedNonconformityRegistryId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteNonconformityRegistry = (
  registry: { id: string; referenceId: string },
  connectionId: string
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteNonconformityRegistryMutation, {
    successMessage: __("Non conformity registry entry deleted successfully"),
    errorMessage: __("Failed to delete non conformity registry entry"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              nonconformityRegistryId: registry.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the non conformity registry entry %s. This action cannot be undone."
          ),
          registry.referenceId
        ),
      }
    );
  };
};

export const useCreateNonconformityRegistry = (connectionId: string) => {
  const [mutate] = useMutation(createNonconformityRegistryMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    referenceId: string;
    description?: string;
    auditId: string;
    dateIdentified?: string;
    rootCause: string;
    correctiveAction?: string;
    ownerId: string;
    dueDate?: string;
    status: string;
    effectivenessCheck?: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create non conformity registry entry: organization is required"));
    }
    if (!input.referenceId) {
      return alert(__("Failed to create non conformity registry entry: reference ID is required"));
    }
    if (!input.auditId) {
      return alert(__("Failed to create non conformity registry entry: audit is required"));
    }
    if (!input.ownerId) {
      return alert(__("Failed to create non conformity registry entry: owner is required"));
    }
    if (!input.rootCause) {
      return alert(__("Failed to create non conformity registry entry: root cause is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          referenceId: input.referenceId,
          description: input.description,
          auditId: input.auditId,
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

export const useUpdateNonconformityRegistry = () => {
  const [mutate] = useMutation(updateNonconformityRegistryMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    referenceId?: string;
    description?: string;
    dateIdentified?: string;
    rootCause?: string;
    correctiveAction?: string;
    ownerId?: string;
    auditId?: string;
    dueDate?: string;
    status?: string;
    effectivenessCheck?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update non conformity registry entry: registry ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
