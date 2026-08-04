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

export const auditProgramsQuery = graphql`
  query AuditProgramGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateAuditProgram: permission(action: "core:audit-program:create")
        ...AuditProgramsPageFragment
      }
    }
  }
`;

export const auditProgramNodeQuery = graphql`
  query AuditProgramGraphNodeQuery($auditProgramId: ID!) {
    node(id: $auditProgramId) {
      ... on AuditProgram {
        id
        name
        validFrom
        validUntil
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
        canUpdate: permission(action: "core:audit-program:update")
        canDelete: permission(action: "core:audit-program:delete")
        ...AuditProgramDetailsPageAuditsFragment
      }
    }
  }
`;

export const createAuditProgramMutation = graphql`
  mutation AuditProgramGraphCreateMutation(
    $input: CreateAuditProgramInput!
    $connections: [ID!]!
  ) {
    createAuditProgram(input: $input) {
      auditProgramEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          validFrom
          validUntil
          framework {
            id
            name
          }
          createdAt
          canUpdate: permission(action: "core:audit-program:update")
          canDelete: permission(action: "core:audit-program:delete")
        }
      }
    }
  }
`;

export const updateAuditProgramMutation = graphql`
  mutation AuditProgramGraphUpdateMutation($input: UpdateAuditProgramInput!) {
    updateAuditProgram(input: $input) {
      auditProgram {
        id
        name
        validFrom
        validUntil
        framework {
          id
          name
        }
        updatedAt
      }
    }
  }
`;

export const deleteAuditProgramMutation = graphql`
  mutation AuditProgramGraphDeleteMutation(
    $input: DeleteAuditProgramInput!
    $connections: [ID!]!
  ) {
    deleteAuditProgram(input: $input) {
      deletedAuditProgramId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteAuditProgram = (
  auditProgram: { id: string; name: string },
  connectionId: string,
  onSuccess?: () => void,
) => {
  const { t } = useTranslation("organizations/audit-programs");
  const [mutate] = useMutationWithToasts(deleteAuditProgramMutation, {
    successMessage: t("auditProgramGraph.messages.deleted"),
    errorMessage: t("auditProgramGraph.errors.delete"),
  });
  const confirm = useConfirm();

  return () => {
    confirm(
      async () => {
        await mutate({
          variables: {
            input: {
              auditProgramId: auditProgram.id,
            },
            connections: [connectionId],
          },
        });
        onSuccess?.();
      },
      {
        message: t("auditProgramGraph.deleteConfirmation", {
          name: auditProgram.name,
        }),
      },
    );
  };
};

export const useCreateAuditProgram = (connectionId: string) => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(createAuditProgramMutation);
  const { t } = useTranslation("organizations/audit-programs");

  return (input: {
    organizationId: string;
    frameworkId: string;
    name: string;
    validFrom?: string;
    validUntil?: string;
  }) => {
    if (!input.organizationId) {
      return alert(t("auditProgramGraph.errors.createOrganizationRequired"));
    }
    if (!input.frameworkId) {
      return alert(t("auditProgramGraph.errors.createFrameworkRequired"));
    }
    if (!input.name) {
      return alert(t("auditProgramGraph.errors.createNameRequired"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input: {
          organizationId: input.organizationId,
          frameworkId: input.frameworkId,
          name: input.name,
          validFrom: input.validFrom,
          validUntil: input.validUntil,
        },
        connections: [connectionId],
      },
    });
  };
};

export const useUpdateAuditProgram = () => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(updateAuditProgramMutation);
  const { t } = useTranslation("organizations/audit-programs");

  return (input: {
    id: string;
    name?: string | null;
    validFrom?: string | null;
    validUntil?: string | null;
  }) => {
    if (!input.id) {
      return alert(t("auditProgramGraph.errors.updateIdRequired"));
    }

    return promisifyMutation(mutate)({
      variables: {
        input,
      },
    });
  };
};
