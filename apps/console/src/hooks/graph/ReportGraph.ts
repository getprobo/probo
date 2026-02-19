import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "../useMutationWithToasts";

export const createReportMutation = graphql`
  mutation ReportGraphCreateMutation(
    $input: CreateReportInput!
    $connections: [ID!]!
  ) {
    createReport(input: $input) {
      reportEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          frameworkType
          validFrom
          validUntil
          file {
            id
            fileName
          }
          state
          framework {
            id
            name
          }
          createdAt
          canUpdate: permission(action: "core:report:update")
          canDelete: permission(action: "core:report:delete")
        }
      }
    }
  }
`;

export const updateReportMutation = graphql`
  mutation ReportGraphUpdateMutation($input: UpdateReportInput!) {
    updateReport(input: $input) {
      report {
        id
        name
        frameworkType
        validFrom
        validUntil
        file {
          id
          fileName
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

export const deleteReportMutation = graphql`
  mutation ReportGraphDeleteMutation(
    $input: DeleteReportInput!
    $connections: [ID!]!
  ) {
    deleteReport(input: $input) {
      deletedReportId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteReport = (
  report: { id: string; framework: { name: string } },
  connectionId: string,
  onSuccess?: () => void,
) => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteReportMutation, {
    successMessage: __("Report deleted successfully"),
    errorMessage: __("Failed to delete report"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      async () => {
        await mutate({
          variables: {
            input: {
              reportId: report.id,
            },
            connections: [connectionId],
          },
        });
        onSuccess?.();
      },
      {
        message: sprintf(
          __(
            "This will permanently delete the report for %s. This action cannot be undone.",
          ),
          report.framework.name,
        ),
      },
    );
  };
};

export const useCreateReport = (connectionId: string) => {
  const [mutate] = useMutation(createReportMutation);
  const { __ } = useTranslate();

  return (input: {
    organizationId: string;
    frameworkId: string;
    name?: string | null;
    frameworkType?: string | null;
    validFrom?: string;
    validUntil?: string;
    state?: string;
  }) => {
    if (!input.organizationId) {
      return alert(__("Failed to create report: organization is required"));
    }
    if (!input.frameworkId) {
      return alert(__("Failed to create report: framework is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          frameworkId: input.frameworkId,
          name: input.name,
          frameworkType: input.frameworkType,
          validFrom: input.validFrom,
          validUntil: input.validUntil,
          state: input.state || "NOT_STARTED",
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateReport = () => {
  const [mutate] = useMutation(updateReportMutation);
  const { __ } = useTranslate();

  return (input: {
    id: string;
    name?: string | null;
    frameworkType?: string | null;
    validFrom?: string | null;
    validUntil?: string | null;
    state?: string;
  }) => {
    if (!input.id) {
      return alert(__("Failed to update report: report ID is required"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};

export const uploadReportFileMutation = graphql`
  mutation ReportGraphUploadFileMutation($input: UploadReportFileInput!) {
    uploadReportFile(input: $input) {
      report {
        id
        file {
          id
          fileName
          downloadUrl
          createdAt
        }
        updatedAt
      }
    }
  }
`;

export const useUploadReportFile = () => {
  const { __ } = useTranslate();
  const [mutate, isLoading] = useMutationWithToasts(uploadReportFileMutation, {
    successMessage: __("Report file uploaded successfully"),
    errorMessage: __("Failed to upload report file"),
  });

  const uploadReportFile = (input: { reportId: string; file: File }) => {
    if (!input.reportId) {
      return alert(__("Failed to upload file: report ID is required"));
    }

    return mutate({
      variables: {
        input: {
          reportId: input.reportId,
          file: null,
        },
      },
      uploadables: {
        "input.file": input.file,
      },
    });
  };

  return [uploadReportFile, isLoading] as const;
};

export const deleteReportFileMutation = graphql`
  mutation ReportGraphDeleteFileMutation($input: DeleteReportFileInput!) {
    deleteReportFile(input: $input) {
      report {
        id
        file {
          id
          fileName
          downloadUrl
          createdAt
        }
        updatedAt
      }
    }
  }
`;

export const useDeleteReportFile = () => {
  const { __ } = useTranslate();
  const [mutate] = useMutationWithToasts(deleteReportFileMutation, {
    successMessage: __("Report file deleted successfully"),
    errorMessage: __("Failed to delete report file"),
  });

  return (input: { reportId: string }) => {
    return mutate({
      variables: {
        input: {
          reportId: input.reportId,
        },
      },
    });
  };
};
