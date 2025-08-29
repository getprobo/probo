import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const ComplianceRegistriesConnectionKey = "ComplianceRegistriesPage_complianceRegistries";

export const complianceRegistriesQuery = graphql`
  query ComplianceRegistryGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        ...ComplianceRegistriesPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const complianceRegistryNodeQuery = graphql`
  query ComplianceRegistryGraphNodeQuery($complianceRegistryId: ID!) {
    node(id: $complianceRegistryId) {
      ... on ComplianceRegistry {
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

export const createComplianceRegistryMutation = graphql`
  mutation ComplianceRegistryGraphCreateMutation(
    $input: CreateComplianceRegistryInput!
    $connections: [ID!]!
  ) {
    createComplianceRegistry(input: $input) {
      complianceRegistryEdge @prependEdge(connections: $connections) {
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

export const updateComplianceRegistryMutation = graphql`
  mutation ComplianceRegistryGraphUpdateMutation($input: UpdateComplianceRegistryInput!) {
    updateComplianceRegistry(input: $input) {
      complianceRegistry {
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

export const deleteComplianceRegistryMutation = graphql`
  mutation ComplianceRegistryGraphDeleteMutation(
    $input: DeleteComplianceRegistryInput!
    $connections: [ID!]!
  ) {
    deleteComplianceRegistry(input: $input) {
      deletedComplianceRegistryId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteComplianceRegistry = (
  registry: { id: string; referenceId: string },
  connectionId: string
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteComplianceRegistryMutation, {
    successMessage: __("Compliance registry entry deleted successfully"),
    errorMessage: __("Failed to delete compliance registry entry"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              complianceRegistryId: registry.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the compliance registry entry %s. This action cannot be undone."
          ),
          registry.referenceId
        ),
      }
    );
  };
};

export const useCreateComplianceRegistry = (connectionId: string) => {
  const [mutate] = useMutation(createComplianceRegistryMutation);
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
      return alert(__("Failed to create compliance registry entry: organization is required"));
    }
    if (!input.referenceId) {
      return alert(__("Failed to create compliance registry entry: reference ID is required"));
    }
    if (!input.ownerId) {
      return alert(__("Failed to create compliance registry entry: owner is required"));
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

export const useUpdateComplianceRegistry = () => {
  const [mutate] = useMutation(updateComplianceRegistryMutation);
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
      return alert(__("Failed to update compliance registry entry: registry ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
