import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const ProcessingActivitiesConnectionKey = "ProcessingActivitiesPage_processingActivities";
export type ProcessingActivityDPIAResidualRisk = "LOW" | "MEDIUM" | "HIGH";

export const processingActivitiesQuery = graphql`
  query ProcessingActivityGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        ...ProcessingActivitiesPageFragment @arguments(snapshotId: $snapshotId)
        ...ProcessingActivitiesPageDPIAFragment @arguments(snapshotId: $snapshotId)
        ...ProcessingActivitiesPageTIAFragment @arguments(snapshotId: $snapshotId)
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
        lastReviewDate
        nextReviewDate
        role
        dataProtectionOfficer {
          id
          fullName
        }
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
        dpia {
          id
          description
          necessityAndProportionality
          potentialRisk
          mitigations
          residualRisk
          createdAt
          updatedAt
        }
        tia {
          id
          dataSubjects
          legalMechanism
          transfer
          localLawRisk
          supplementaryMeasures
          createdAt
          updatedAt
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
          lastReviewDate
          nextReviewDate
          role
          dataProtectionOfficer {
            id
            fullName
          }
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
        lastReviewDate
        nextReviewDate
        role
        dataProtectionOfficer {
          id
          fullName
        }
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
    lastReviewDate?: string;
    nextReviewDate?: string;
    role: string;
    dataProtectionOfficerId?: string;
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
          lastReviewDate: input.lastReviewDate,
          nextReviewDate: input.nextReviewDate,
          role: input.role,
          dataProtectionOfficerId: input.dataProtectionOfficerId,
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
    lastReviewDate?: string | null;
    nextReviewDate?: string | null;
    role?: string;
    dataProtectionOfficerId?: string | null;
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

export const createProcessingActivityDPIAMutation = graphql`
  mutation ProcessingActivityGraphCreateDPIAMutation(
    $input: CreateProcessingActivityDPIAInput!
  ) {
    createProcessingActivityDPIA(input: $input) {
      processingActivityDpia {
        id
        description
        necessityAndProportionality
        potentialRisk
        mitigations
        residualRisk
        createdAt
        updatedAt
      }
    }
  }
`;

export const updateProcessingActivityDPIAMutation = graphql`
  mutation ProcessingActivityGraphUpdateDPIAMutation(
    $input: UpdateProcessingActivityDPIAInput!
  ) {
    updateProcessingActivityDPIA(input: $input) {
      processingActivityDpia {
        id
        description
        necessityAndProportionality
        potentialRisk
        mitigations
        residualRisk
        createdAt
        updatedAt
      }
    }
  }
`;

export const deleteProcessingActivityDPIAMutation = graphql`
  mutation ProcessingActivityGraphDeleteDPIAMutation(
    $input: DeleteProcessingActivityDPIAInput!
  ) {
    deleteProcessingActivityDPIA(input: $input) {
      deletedProcessingActivityDpiaId
    }
  }
`;

export const useCreateProcessingActivityDPIA = () => {
  const [mutate] = useMutation(createProcessingActivityDPIAMutation);
  const { __ } = useTranslate();

  return (input: {
    processingActivityId: string;
    description?: string;
    necessityAndProportionality?: string;
    potentialRisk?: string;
    mitigations?: string;
    residualRisk?: ProcessingActivityDPIAResidualRisk;
  }) => {
    if (!input.processingActivityId) {
      return alert(__("Failed to create DPIA: Processing Activity ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};

export const useUpdateProcessingActivityDPIA = () => {
  const [mutate] = useMutation(updateProcessingActivityDPIAMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    description?: string;
    necessityAndProportionality?: string;
    potentialRisk?: string;
    mitigations?: string;
    residualRisk?: ProcessingActivityDPIAResidualRisk;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update DPIA: ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};

export const useDeleteProcessingActivityDPIA = (
  dpia: { id: string },
  options?: { onSuccess?: () => void }
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteProcessingActivityDPIAMutation, {
    successMessage: __("DPIA deleted successfully"),
    errorMessage: __("Failed to delete DPIA"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              processingActivityDpiaId: dpia.id,
            },
          },
          onSuccess: options?.onSuccess,
        }),
      {
        message: __(
          "This will permanently delete this Data Protection Impact Assessment. This action cannot be undone."
        ),
      }
    );
  };
};

export const createProcessingActivityTIAMutation = graphql`
  mutation ProcessingActivityGraphCreateTIAMutation(
    $input: CreateProcessingActivityTIAInput!
  ) {
    createProcessingActivityTIA(input: $input) {
      processingActivityTia {
        id
        dataSubjects
        legalMechanism
        transfer
        localLawRisk
        supplementaryMeasures
        createdAt
        updatedAt
      }
    }
  }
`;

export const updateProcessingActivityTIAMutation = graphql`
  mutation ProcessingActivityGraphUpdateTIAMutation(
    $input: UpdateProcessingActivityTIAInput!
  ) {
    updateProcessingActivityTIA(input: $input) {
      processingActivityTia {
        id
        dataSubjects
        legalMechanism
        transfer
        localLawRisk
        supplementaryMeasures
        createdAt
        updatedAt
      }
    }
  }
`;

export const deleteProcessingActivityTIAMutation = graphql`
  mutation ProcessingActivityGraphDeleteTIAMutation(
    $input: DeleteProcessingActivityTIAInput!
  ) {
    deleteProcessingActivityTIA(input: $input) {
      deletedProcessingActivityTiaId
    }
  }
`;

export const useCreateProcessingActivityTIA = () => {
  const [mutate] = useMutation(createProcessingActivityTIAMutation);
  const { __ } = useTranslate();

  return (input: {
    processingActivityId: string;
    dataSubjects?: string;
    legalMechanism?: string;
    transfer?: string;
    localLawRisk?: string;
    supplementaryMeasures?: string;
  }) => {
    if (!input.processingActivityId) {
      return alert(__("Failed to create TIA: Processing Activity ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};

export const useUpdateProcessingActivityTIA = () => {
  const [mutate] = useMutation(updateProcessingActivityTIAMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    dataSubjects?: string;
    legalMechanism?: string;
    transfer?: string;
    localLawRisk?: string;
    supplementaryMeasures?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update TIA: ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};

export const useDeleteProcessingActivityTIA = (
  tia: { id: string },
  options?: { onSuccess?: () => void }
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteProcessingActivityTIAMutation, {
    successMessage: __("TIA deleted successfully"),
    errorMessage: __("Failed to delete TIA"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              processingActivityTiaId: tia.id,
            },
          },
          onSuccess: options?.onSuccess,
        }),
      {
        message: __(
          "This will permanently delete this Transfer Impact Assessment. This action cannot be undone."
        ),
      }
    );
  };
};
