import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { useConfirm } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "../useMutationWithToasts";

export const auditsQuery = graphql`
  query AuditGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        ...AuditsPageFragment
      }
    }
  }
`;

export const auditNodeQuery = graphql`
  query AuditGraphNodeQuery($auditId: ID!) {
    node(id: $auditId) {
      ... on Audit {
        id
        name
        validFrom
        validUntil
        reports {
          id
          filename
          mimeType
          size
          downloadUrl
          createdAt
        }
        state
        framework {
          id
          name
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

export const createAuditMutation = graphql`
  mutation AuditGraphCreateMutation(
    $input: CreateAuditInput!
    $connections: [ID!]!
  ) {
    createAudit(input: $input) {
      auditEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          validFrom
          validUntil
          reports {
            id
            filename
          }
          state
          framework {
            id
            name
          }
          createdAt
        }
      }
    }
  }
`;

export const updateAuditMutation = graphql`
  mutation AuditGraphUpdateMutation($input: UpdateAuditInput!) {
    updateAudit(input: $input) {
      audit {
        id
        name
        validFrom
        validUntil
        reports {
          id
          filename
        }
        state
        framework {
          id
          name
        }
        updatedAt
      }
    }
  }
`;

export const deleteAuditMutation = graphql`
  mutation AuditGraphDeleteMutation(
    $input: DeleteAuditInput!
    $connections: [ID!]!
  ) {
    deleteAudit(input: $input) {
      deletedAuditId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteAudit = (
  audit: { id: string; framework: { name: string } },
  connectionId: string
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteAuditMutation, {
    successMessage: __("Audit deleted successfully"),
    errorMessage: __("Failed to delete audit"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      () =>
        mutate({
          variables: {
            input: {
              auditId: audit.id!,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the audit for %s. This action cannot be undone."
          ),
          audit.framework.name
        ),
      }
    );
  };
};

export const useCreateAudit = (connectionId: string) => {
  const [mutate] = useMutation(createAuditMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    frameworkId: string;
    name?: string;
    validFrom?: string;
    validUntil?: string;
    reportKey?: string;
    state?: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create audit: organization is required"));
    }
    if (!input.frameworkId) {
      return alert(__("Failed to create audit: framework is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          frameworkId: input.frameworkId,
          name: input.name,
          validFrom: input.validFrom,
          validUntil: input.validUntil,
          reportKey: input.reportKey,
          state: input.state || "NOT_STARTED",
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateAudit = () => {
  const [mutate] = useMutation(updateAuditMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    name?: string;
    validFrom?: string;
    validUntil?: string;
    state?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update audit: audit ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};

export const uploadAuditReportMutation = graphql`
  mutation AuditGraphUploadReportMutation($input: UploadAuditReportInput!) {
    uploadAuditReport(input: $input) {
      audit {
        id
        reports {
          id
          filename
          downloadUrl
          createdAt
          size
        }
        updatedAt
      }
    }
  }
`;

export const useUploadAuditReport = () => {
  const { __ } = useTranslate();
  const [mutate, isLoading] = useMutationWithToasts(uploadAuditReportMutation, {
    successMessage: __("Audit report uploaded successfully"),
    errorMessage: __("Failed to upload audit report"),
  });

  const uploadAuditReport = (input: { auditId: string; file: File }) => {
    if (!input.auditId) {
      return alert(__("Failed to upload report: audit ID is required"));
    }

    return mutate({
      variables: {
        input: {
          auditId: input.auditId,
          file: null,
        },
      },
      uploadables: {
        "input.file": input.file,
      },
    });
  };

  return [uploadAuditReport, isLoading] as const;
};

export const deleteAuditReportMutation = graphql`
  mutation AuditGraphDeleteReportMutation($input: DeleteAuditReportInput!) {
    deleteAuditReport(input: $input) {
      audit {
        id
        reports {
          id
          filename
          downloadUrl
          createdAt
          size
        }
        updatedAt
      }
    }
  }
`;

export const useDeleteAuditReport = () => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteAuditReportMutation, {
    successMessage: __("Audit report deleted successfully"),
    errorMessage: __("Failed to delete audit report"),
  });

  return (input: { auditId: string; reportId: string }) => {
    return mutate({
      variables: {
        input: {
          auditId: input.auditId,
          reportId: input.reportId,
        },
      },
    });
  };
};
