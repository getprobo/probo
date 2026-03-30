// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "../useMutationWithToasts";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const ProcessingActivitiesConnectionKey
  = "ProcessingActivitiesPage_processingActivities";
export type ProcessingActivityDPIAResidualRisk = "LOW" | "MEDIUM" | "HIGH";

export const processingActivitiesQuery = graphql`
  query ProcessingActivityGraphListQuery(
    $organizationId: ID!
    $snapshotId: ID
  ) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateProcessingActivity: permission(
          action: "core:processing-activity:create"
        )
        canExportProcessingActivities: permission(
          action: "core:processing-activity:export"
        )
        canExportDataProtectionImpactAssessments: permission(
          action: "core:data-protection-impact-assessment:export"
        )
        canExportTransferImpactAssessments: permission(
          action: "core:transfer-impact-assessment:export"
        )
        ...ProcessingActivitiesPageFragment @arguments(snapshotId: $snapshotId)
        ...ProcessingActivitiesPageDPIAFragment
          @arguments(snapshotId: $snapshotId)
        ...ProcessingActivitiesPageTIAFragment
          @arguments(snapshotId: $snapshotId)
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
        dataProtectionImpactAssessmentNeeded
        transferImpactAssessmentNeeded
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
        dataProtectionImpactAssessment {
          id
          description
          necessityAndProportionality
          potentialRisk
          mitigations
          residualRisk
          createdAt
          updatedAt
          canUpdate: permission(
            action: "core:data-protection-impact-assessment:update"
          )
          canDelete: permission(
            action: "core:data-protection-impact-assessment:delete"
          )
        }
        transferImpactAssessment {
          id
          dataSubjects
          legalMechanism
          transfer
          localLawRisk
          supplementaryMeasures
          createdAt
          updatedAt
          canUpdate: permission(
            action: "core:transfer-impact-assessment:update"
          )
          canDelete: permission(
            action: "core:transfer-impact-assessment:delete"
          )
        }
        organization {
          id
          name
        }
        createdAt
        updatedAt
        canCreateDPIA: permission(
          action: "core:data-protection-impact-assessment:create"
        )
        canCreateTIA: permission(
          action: "core:transfer-impact-assessment:create"
        )
        canUpdate: permission(action: "core:processing-activity:update")
        canDelete: permission(action: "core:processing-activity:delete")
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
          dataProtectionImpactAssessmentNeeded
          transferImpactAssessmentNeeded
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
          canUpdate: permission(action: "core:processing-activity:update")
          canDelete: permission(action: "core:processing-activity:delete")
        }
      }
    }
  }
`;

export const updateProcessingActivityMutation = graphql`
  mutation ProcessingActivityGraphUpdateMutation(
    $input: UpdateProcessingActivityInput!
  ) {
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
        dataProtectionImpactAssessmentNeeded
        transferImpactAssessmentNeeded
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
  connectionId: string,
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
            "This will permanently delete the processing activity %s. This action cannot be undone.",
          ),
          processingActivity.name,
        ),
      },
    );
  };
};

export const useCreateProcessingActivity = (connectionId?: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
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
    dataProtectionImpactAssessmentNeeded?: string;
    transferImpactAssessmentNeeded?: string;
    lastReviewDate?: string;
    nextReviewDate?: string;
    role: string;
    dataProtectionOfficerId?: string;
    vendorIds?: string[];
  }) => {
    if (!input.organizationId) {
      return alert(
        __("Failed to create processing activity: organization is required"),
      );
    }
    if (!input.name) {
      return alert(
        __("Failed to create processing activity: name is required"),
      );
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
          dataProtectionImpactAssessmentNeeded:
            input.dataProtectionImpactAssessmentNeeded,
          transferImpactAssessmentNeeded: input.transferImpactAssessmentNeeded,
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
  // eslint-disable-next-line relay/generated-typescript-types
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
    dataProtectionImpactAssessmentNeeded?: string;
    transferImpactAssessmentNeeded?: string;
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

export const createDataProtectionImpactAssessmentMutation = graphql`
  mutation ProcessingActivityGraphCreateDPIAMutation(
    $input: CreateDataProtectionImpactAssessmentInput!
  ) {
    createDataProtectionImpactAssessment(input: $input) {
      dataProtectionImpactAssessment {
        id
        description
        necessityAndProportionality
        potentialRisk
        mitigations
        residualRisk
        createdAt
        updatedAt
        canUpdate: permission(
          action: "core:data-protection-impact-assessment:update"
        )
        canDelete: permission(
          action: "core:data-protection-impact-assessment:delete"
        )
        processingActivity {
          id
          dataProtectionImpactAssessment {
            id
            description
            necessityAndProportionality
            potentialRisk
            mitigations
            residualRisk
            createdAt
            updatedAt
            canUpdate: permission(
              action: "core:data-protection-impact-assessment:update"
            )
            canDelete: permission(
              action: "core:data-protection-impact-assessment:delete"
            )
          }
        }
      }
    }
  }
`;

export const updateDataProtectionImpactAssessmentMutation = graphql`
  mutation ProcessingActivityGraphUpdateDPIAMutation(
    $input: UpdateDataProtectionImpactAssessmentInput!
  ) {
    updateDataProtectionImpactAssessment(input: $input) {
      dataProtectionImpactAssessment {
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

export const deleteDataProtectionImpactAssessmentMutation = graphql`
  mutation ProcessingActivityGraphDeleteDPIAMutation(
    $input: DeleteDataProtectionImpactAssessmentInput!
  ) {
    deleteDataProtectionImpactAssessment(input: $input) {
      deletedDataProtectionImpactAssessmentId
    }
  }
`;

export const useCreateDataProtectionImpactAssessment = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createDataProtectionImpactAssessmentMutation);
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
      return alert(
        __("Failed to create DPIA: Processing Activity ID is required"),
      );
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};

export const useUpdateDataProtectionImpactAssessment = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(updateDataProtectionImpactAssessmentMutation);
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

export const useDeleteDataProtectionImpactAssessment = (
  dpia: { id: string },
  options?: { onSuccess?: () => void },
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(
    deleteDataProtectionImpactAssessmentMutation,
    {
      successMessage: __("DPIA deleted successfully"),
      errorMessage: __("Failed to delete DPIA"),
    },
  );
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              dataProtectionImpactAssessmentId: dpia.id,
            },
          },
          onSuccess: options?.onSuccess,
        }),
      {
        message: __(
          "This will permanently delete this Data Protection Impact Assessment. This action cannot be undone.",
        ),
      },
    );
  };
};

export const createTransferImpactAssessmentMutation = graphql`
  mutation ProcessingActivityGraphCreateTIAMutation(
    $input: CreateTransferImpactAssessmentInput!
  ) {
    createTransferImpactAssessment(input: $input) {
      transferImpactAssessment {
        id
        dataSubjects
        legalMechanism
        transfer
        localLawRisk
        supplementaryMeasures
        createdAt
        updatedAt
        canUpdate: permission(
          action: "core:transfer-impact-assessment:update"
        )
        canDelete: permission(
          action: "core:transfer-impact-assessment:delete"
        )
        processingActivity {
          id
          transferImpactAssessment {
            id
            dataSubjects
            legalMechanism
            transfer
            localLawRisk
            supplementaryMeasures
            createdAt
            updatedAt
            canUpdate: permission(
              action: "core:transfer-impact-assessment:update"
            )
            canDelete: permission(
              action: "core:transfer-impact-assessment:delete"
            )
          }
        }
      }
    }
  }
`;

export const updateTransferImpactAssessmentMutation = graphql`
  mutation ProcessingActivityGraphUpdateTIAMutation(
    $input: UpdateTransferImpactAssessmentInput!
  ) {
    updateTransferImpactAssessment(input: $input) {
      transferImpactAssessment {
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

export const deleteTransferImpactAssessmentMutation = graphql`
  mutation ProcessingActivityGraphDeleteTIAMutation(
    $input: DeleteTransferImpactAssessmentInput!
  ) {
    deleteTransferImpactAssessment(input: $input) {
      deletedTransferImpactAssessmentId
    }
  }
`;

export const useCreateTransferImpactAssessment = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createTransferImpactAssessmentMutation);
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
      return alert(
        __("Failed to create TIA: Processing Activity ID is required"),
      );
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};

export const useUpdateTransferImpactAssessment = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(updateTransferImpactAssessmentMutation);
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

export const useDeleteTransferImpactAssessment = (
  tia: { id: string },
  options?: { onSuccess?: () => void },
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(
    deleteTransferImpactAssessmentMutation,
    {
      successMessage: __("TIA deleted successfully"),
      errorMessage: __("Failed to delete TIA"),
    },
  );
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              transferImpactAssessmentId: tia.id,
            },
          },
          onSuccess: options?.onSuccess,
        }),
      {
        message: __(
          "This will permanently delete this Transfer Impact Assessment. This action cannot be undone.",
        ),
      },
    );
  };
};
