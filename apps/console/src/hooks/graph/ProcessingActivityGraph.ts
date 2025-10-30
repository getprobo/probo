import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const ProcessingActivitiesConnectionKey = "ProcessingActivitiesPage_processingActivities";

export const processingActivitiesQuery = graphql`
  query ProcessingActivityGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        ...ProcessingActivitiesPageFragment @arguments(snapshotId: $snapshotId)
      }
    }
  }
`;

export const processingActivityNodeQuery = graphql`
  query ProcessingActivityGraphNodeQuery($processingActivityId: ID!) {
    node(id: $processingActivityId) {
      ... on ProcessingActivity {
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
        vendors(first: 50) {
          edges {
            node {
              id
              name
              websiteUrl
              category
            }
          }
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

export const createProcessingActivityMutation = graphql`
  mutation ProcessingActivityGraphCreateMutation(
    $input: CreateProcessingActivityInput!
    $connections: [ID!]!
  ) {
    createProcessingActivity(input: $input) {
      processingActivityEdge @prependEdge(connections: $connections) {
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
          vendors(first: 50) {
            edges {
              node {
                id
                name
                websiteUrl
              }
            }
          }
          createdAt
        }
      }
    }
  }
`;

export const updateProcessingActivityMutation = graphql`
  mutation ProcessingActivityGraphUpdateMutation($input: UpdateProcessingActivityInput!) {
    updateProcessingActivity(input: $input) {
      processingActivity {
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
        vendors(first: 50) {
          edges {
            node {
              id
              name
              websiteUrl
            }
          }
        }
        updatedAt
      }
    }
  }
`;

export const deleteProcessingActivityMutation = graphql`
  mutation ProcessingActivityGraphDeleteMutation(
    $input: DeleteProcessingActivityInput!
    $connections: [ID!]!
  ) {
    deleteProcessingActivity(input: $input) {
      deletedProcessingActivityId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteProcessingActivity = (
  processingActivity: { id: string; name: string },
  connectionId: string
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteProcessingActivityMutation, {
    successMessage: __("Processing activity deleted successfully"),
    errorMessage: __("Failed to delete processing activity"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              processingActivityId: processingActivity.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the processing activity %s. This action cannot be undone."
          ),
          processingActivity.name
        ),
      }
    );
  };
};

export const useCreateProcessingActivity = (connectionId?: string) => {
  const [mutate] = useMutation(createProcessingActivityMutation);
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
    vendorIds?: string[];
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create processing activity: organization is required"));
    }
    if (!input.name) {
      return alert(__("Failed to create processing activity: name is required"));
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
          vendorIds: input.vendorIds,
        },
        connections: connectionId ? [connectionId] : [],
      },
    });
  };
};

export const useUpdateProcessingActivity = () => {
  const [mutate] = useMutation(updateProcessingActivityMutation);
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
    vendorIds?: string[];
  }) => {
    if (!input.id) {
      return alert(__("Failed to update processing activity: ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
