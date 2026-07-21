// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { promisifyMutation } from "@probo/helpers";
import { useConfirm } from "@probo/ui";
import { useTranslation } from "react-i18next";
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
  ) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateProcessingActivity: permission(
          action: "core:processing-activity:create"
        )
        canPublishProcessingActivities: permission(
          action: "core:processing-activity:publish"
        )
        canPublishDataProtectionImpactAssessments: permission(
          action: "core:data-protection-impact-assessment:publish"
        )
        canPublishTransferImpactAssessments: permission(
          action: "core:transfer-impact-assessment:publish"
        )
        processingActivitiesDocument {
          id
          defaultApprovers {
            id
          }
        }
        dataProtectionImpactAssessmentsDocument {
          id
          defaultApprovers {
            id
          }
        }
        transferImpactAssessmentsDocument {
          id
          defaultApprovers {
            id
          }
        }
        ...ProcessingActivitiesPageFragment
        ...ProcessingActivitiesPageDPIAFragment
        ...ProcessingActivitiesPageTIAFragment
      }
    }
  }
`;

export const processingActivityNodeQuery = graphql`
  query ProcessingActivityGraphNodeQuery($processingActivityId: ID!) {
    node(id: $processingActivityId) {
      ... on ProcessingActivity {
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
        thirdParties(first: 50) {
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
          thirdParties(first: 50) {
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
        thirdParties(first: 50) {
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
  const { t } = useTranslation();
  const [mutate] = useMutationWithToasts(deleteProcessingActivityMutation, {
    successMessage: t("processingActivityGraph.messages.deleted"),
    errorMessage: t("processingActivityGraph.errors.delete"),
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
        message: t("processingActivityGraph.deleteConfirmation", {
          name: processingActivity.name,
        }),
      },
    );
  };
};

export const useCreateProcessingActivity = (connectionId?: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createProcessingActivityMutation);
  const { t } = useTranslation();

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
    thirdPartyIds?: string[];
  }) => {
    if (!input.organizationId) {
      return alert(
        t("processingActivityGraph.errors.createOrganizationRequired"),
      );
    }
    if (!input.name) {
      return alert(
        t("processingActivityGraph.errors.createNameRequired"),
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
          thirdPartyIds: input.thirdPartyIds,
        },
        connections: connectionId ? [connectionId] : [],
      },
    });
  };
};

export const useUpdateProcessingActivity = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(updateProcessingActivityMutation);
  const { t } = useTranslation();

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
    thirdPartyIds?: string[];
  }) => {
    if (!input.id) {
      return alert(t("processingActivityGraph.errors.updateIdRequired"));
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
  const { t } = useTranslation();

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
        t("processingActivityGraph.errors.createDpiaProcessingActivityIdRequired"),
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
  const { t } = useTranslation();

  return (input: {
    id: string;
    description?: string;
    necessityAndProportionality?: string;
    potentialRisk?: string;
    mitigations?: string;
    residualRisk?: ProcessingActivityDPIAResidualRisk;
  }) => {
    if (!input.id) {
      return alert(t("processingActivityGraph.errors.updateDpiaIdRequired"));
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
  const { t } = useTranslation();
  const [mutate] = useMutationWithToasts(
    deleteDataProtectionImpactAssessmentMutation,
    {
      successMessage: t("processingActivityGraph.messages.dpiaDeleted"),
      errorMessage: t("processingActivityGraph.errors.deleteDpia"),
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
        message: t("processingActivityGraph.dpiaDeleteConfirmation"),
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
  const { t } = useTranslation();

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
        t("processingActivityGraph.errors.createTiaProcessingActivityIdRequired"),
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
  const { t } = useTranslation();

  return (input: {
    id: string;
    dataSubjects?: string;
    legalMechanism?: string;
    transfer?: string;
    localLawRisk?: string;
    supplementaryMeasures?: string;
  }) => {
    if (!input.id) {
      return alert(t("processingActivityGraph.errors.updateTiaIdRequired"));
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
  const { t } = useTranslation();
  const [mutate] = useMutationWithToasts(
    deleteTransferImpactAssessmentMutation,
    {
      successMessage: t("processingActivityGraph.messages.tiaDeleted"),
      errorMessage: t("processingActivityGraph.errors.deleteTia"),
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
        message: t("processingActivityGraph.tiaDeleteConfirmation"),
      },
    );
  };
};
