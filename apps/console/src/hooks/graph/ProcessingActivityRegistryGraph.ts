import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const ProcessingActivityRegistriesConnectionKey = "ProcessingActivityRegistriesPage_processingActivityRegistries";

export const processingActivityRegistriesQuery = graphql`
  query ProcessingActivityRegistryGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        ...ProcessingActivityRegistriesPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const processingActivityRegistryNodeQuery = graphql`
  query ProcessingActivityRegistryGraphNodeQuery($processingActivityRegistryId: ID!) {
    node(id: $processingActivityRegistryId) {
      ... on ProcessingActivityRegistry {
        id
        snapshotId
        name
        purpose
        dataSubjectCategory
        personalDataCategory
        specialOrCriminalData
        consentEvidenceLink
        lawfulBasis
        recipients
        location
        internationalTransfers
        transferSafeguards
        retentionPeriod
        securityMeasures
        dataProtectionImpactAssessment
        transferImpactAssessment
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

export const createProcessingActivityRegistryMutation = graphql`
  mutation ProcessingActivityRegistryGraphCreateMutation(
    $input: CreateProcessingActivityRegistryInput!
    $connections: [ID!]!
  ) {
    createProcessingActivityRegistry(input: $input) {
      processingActivityRegistryEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          purpose
          dataSubjectCategory
          personalDataCategory
          specialOrCriminalData
          consentEvidenceLink
          lawfulBasis
          recipients
          location
          internationalTransfers
          transferSafeguards
          retentionPeriod
          securityMeasures
          dataProtectionImpactAssessment
          transferImpactAssessment
          createdAt
        }
      }
    }
  }
`;

export const updateProcessingActivityRegistryMutation = graphql`
  mutation ProcessingActivityRegistryGraphUpdateMutation($input: UpdateProcessingActivityRegistryInput!) {
    updateProcessingActivityRegistry(input: $input) {
      processingActivityRegistry {
        id
        name
        purpose
        dataSubjectCategory
        personalDataCategory
        specialOrCriminalData
        consentEvidenceLink
        lawfulBasis
        recipients
        location
        internationalTransfers
        transferSafeguards
        retentionPeriod
        securityMeasures
        dataProtectionImpactAssessment
        transferImpactAssessment
        updatedAt
      }
    }
  }
`;

export const deleteProcessingActivityRegistryMutation = graphql`
  mutation ProcessingActivityRegistryGraphDeleteMutation(
    $input: DeleteProcessingActivityRegistryInput!
    $connections: [ID!]!
  ) {
    deleteProcessingActivityRegistry(input: $input) {
      deletedProcessingActivityRegistryId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteProcessingActivityRegistry = (
  registry: { id: string; name: string },
  connectionId: string
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteProcessingActivityRegistryMutation, {
    successMessage: __("Processing activity entry deleted successfully"),
    errorMessage: __("Failed to delete processing activity entry"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              processingActivityRegistryId: registry.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the processing activity registry entry %s. This action cannot be undone."
          ),
          registry.name
        ),
      }
    );
  };
};

export const useCreateProcessingActivityRegistry = (connectionId?: string) => {
  const [mutate] = useMutation(createProcessingActivityRegistryMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    name: string;
    purpose?: string;
    dataSubjectCategory?: string;
    personalDataCategory?: string;
    specialOrCriminalData?: string;
    consentEvidenceLink?: string;
    lawfulBasis?: string;
    recipients?: string;
    location?: string;
    internationalTransfers: boolean;
    transferSafeguards?: string;
    retentionPeriod?: string;
    securityMeasures?: string;
    dataProtectionImpactAssessment?: string;
    transferImpactAssessment?: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create processing activity entry: organization is required"));
    }
    if (!input.name) {
      return alert(__("Failed to create processing activity entry: name is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          name: input.name,
          purpose: input.purpose,
          dataSubjectCategory: input.dataSubjectCategory,
          personalDataCategory: input.personalDataCategory,
          specialOrCriminalData: input.specialOrCriminalData,
          consentEvidenceLink: input.consentEvidenceLink,
          lawfulBasis: input.lawfulBasis,
          recipients: input.recipients,
          location: input.location,
          internationalTransfers: input.internationalTransfers,
          transferSafeguards: input.transferSafeguards,
          retentionPeriod: input.retentionPeriod,
          securityMeasures: input.securityMeasures,
          dataProtectionImpactAssessment: input.dataProtectionImpactAssessment,
          transferImpactAssessment: input.transferImpactAssessment,
        },
        connections: connectionId ? [connectionId] : [],
      },
    });
  };
};

export const useUpdateProcessingActivityRegistry = () => {
  const [mutate] = useMutation(updateProcessingActivityRegistryMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    name?: string;
    purpose?: string;
    dataSubjectCategory?: string;
    personalDataCategory?: string;
    specialOrCriminalData?: string;
    consentEvidenceLink?: string;
    lawfulBasis?: string;
    recipients?: string;
    location?: string;
    internationalTransfers?: boolean;
    transferSafeguards?: string;
    retentionPeriod?: string;
    securityMeasures?: string;
    dataProtectionImpactAssessment?: string;
    transferImpactAssessment?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update processing activity entry: entry ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
