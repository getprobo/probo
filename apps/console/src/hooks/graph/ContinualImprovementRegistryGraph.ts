import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const ContinualImprovementRegistriesConnectionKey = "ContinualImprovementRegistriesPage_continualImprovementRegistries";

export const continualImprovementRegistriesQuery = graphql`
  query ContinualImprovementRegistryGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        ...ContinualImprovementRegistriesPageFragment
      }
    }
  }
`;

export const continualImprovementRegistryNodeQuery = graphql`
  query ContinualImprovementRegistryGraphNodeQuery($continualImprovementRegistryId: ID!) {
    node(id: $continualImprovementRegistryId) {
      ... on ContinualImprovementRegistry {
        id
        referenceId
        description
        source
        targetDate
        status
        priority
        audit {
          id
          name
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

export const createContinualImprovementRegistryMutation = graphql`
  mutation ContinualImprovementRegistryGraphCreateMutation(
    $input: CreateContinualImprovementRegistryInput!
    $connections: [ID!]!
  ) {
    createContinualImprovementRegistry(input: $input) {
      continualImprovementRegistryEdge @prependEdge(connections: $connections) {
        node {
          id
          referenceId
          description
          source
          targetDate
          status
          priority
          audit {
            id
            name
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

export const updateContinualImprovementRegistryMutation = graphql`
  mutation ContinualImprovementRegistryGraphUpdateMutation($input: UpdateContinualImprovementRegistryInput!) {
    updateContinualImprovementRegistry(input: $input) {
      continualImprovementRegistry {
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
        audit {
          id
          name
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

export const deleteContinualImprovementRegistryMutation = graphql`
  mutation ContinualImprovementRegistryGraphDeleteMutation(
    $input: DeleteContinualImprovementRegistryInput!
    $connections: [ID!]!
  ) {
    deleteContinualImprovementRegistry(input: $input) {
      deletedContinualImprovementRegistryId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteContinualImprovementRegistry = (
  registry: { id: string; referenceId: string },
  connectionId: string
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteContinualImprovementRegistryMutation, {
    successMessage: __("Continual improvement registry entry deleted successfully"),
    errorMessage: __("Failed to delete continual improvement registry entry"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              continualImprovementRegistryId: registry.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the continual improvement registry entry %s. This action cannot be undone."
          ),
          registry.referenceId
        ),
      }
    );
  };
};

export const useCreateContinualImprovementRegistry = (connectionId: string) => {
  const [mutate] = useMutation(createContinualImprovementRegistryMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    referenceId: string;
    description?: string;
    source?: string;
    auditId: string;
    ownerId: string;
    targetDate?: string;
    status: string;
    priority: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create continual improvement registry entry: organization is required"));
    }
    if (!input.referenceId) {
      return alert(__("Failed to create continual improvement registry entry: reference ID is required"));
    }
    if (!input.auditId) {
      return alert(__("Failed to create continual improvement registry entry: audit is required"));
    }
    if (!input.ownerId) {
      return alert(__("Failed to create continual improvement registry entry: owner is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          referenceId: input.referenceId,
          description: input.description,
          source: input.source,
          auditId: input.auditId,
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

export const useUpdateContinualImprovementRegistry = () => {
  const [mutate] = useMutation(updateContinualImprovementRegistryMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    referenceId?: string;
    description?: string;
    source?: string;
    auditId?: string;
    ownerId?: string;
    targetDate?: string;
    status?: string;
    priority?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update continual improvement registry entry: registry ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
