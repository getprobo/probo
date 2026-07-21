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

export const auditsQuery = graphql`
  query AuditGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateAudit: permission(action: "core:audit:create")
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
        reportFile {
          id
          fileName
          mimeType
          size
          downloadUrl
          createdAt
        }
        state
        framework {
          id
          name
          lightLogo {
            downloadUrl
          }
          darkLogo {
            downloadUrl
          }
        }
        organization {
          id
          name
        }
        createdAt
        updatedAt
        canUpdate: permission(action: "core:audit:update")
        canDelete: permission(action: "core:audit:delete")
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
          reportFile {
            id
            fileName
          }
          state
          framework {
            id
            name
          }
          createdAt
          canUpdate: permission(action: "core:audit:update")
          canDelete: permission(action: "core:audit:delete")
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
        reportFile {
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
  audit: { id: string; framework?: { name: string } | null },
  connectionId: string,
  onSuccess?: () => void,
) => {
  const { t } = useTranslation();
  const [mutate] = useMutationWithToasts(deleteAuditMutation, {
    successMessage: t("auditGraph.messages.deleted"),
    errorMessage: t("auditGraph.errors.delete"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      async () => {
        await mutate({
          variables: {
            input: {
              auditId: audit.id,
            },
            connections: [connectionId],
          },
        });
        onSuccess?.();
      },
      {
        message: t("auditGraph.deleteConfirmation", { frameworkName: audit.framework?.name ?? "" }),
      },
    );
  };
};

export const useCreateAudit = (connectionId: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createAuditMutation);
  const { t } = useTranslation();

  return (input: {
    organizationId: string;
    frameworkId: string;
    name?: string | null;
    validFrom?: string;
    validUntil?: string;
    reportKey?: string;
    state?: string;
    file?: File | null;
  }) => {
    if (!input.organizationId) {
      return alert(t("auditGraph.errors.createOrganizationRequired"));
    }
    if (!input.frameworkId) {
      return alert(t("auditGraph.errors.createFrameworkRequired"));
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
          file: input.file ? null : undefined,
        },
        connections: [connectionId],
      },
      ...(input.file ? { uploadables: { "input.file": input.file } } : {}),
    });
  };
};

export const useUpdateAudit = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(updateAuditMutation);
  const { t } = useTranslation();

  return (input: {
    id: string;
    name?: string | null;
    validFrom?: string | null;
    validUntil?: string | null;
    state?: string;
  }) => {
    if (!input.id) {
      return alert(t("auditGraph.errors.updateIdRequired"));
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
        reportFile {
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

export const useUploadAuditReport = () => {
  const { t } = useTranslation();
  const [mutate, isLoading] = useMutationWithToasts(uploadAuditReportMutation, {
    successMessage: t("auditGraph.messages.reportUploaded"),
    errorMessage: t("auditGraph.errors.uploadReport"),
  });

  const uploadAuditReport = (input: { auditId: string; file: File }) => {
    if (!input.auditId) {
      return alert(t("auditGraph.errors.uploadReportIdRequired"));
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
        reportFile {
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

export const useDeleteAuditReport = () => {
  const { t } = useTranslation();
  const [mutate] = useMutationWithToasts(deleteAuditReportMutation, {
    successMessage: t("auditGraph.messages.reportDeleted"),
    errorMessage: t("auditGraph.errors.deleteReport"),
  });

  return (input: { auditId: string }) => {
    return mutate({
      variables: {
        input: {
          auditId: input.auditId,
        },
      },
    });
  };
};
