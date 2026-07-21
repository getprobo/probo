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

import { useConfirm } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "../useMutationWithToasts";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const connectionListKey = "FrameworksListQuery_frameworks";

export const frameworksQuery = graphql`
  query FrameworkGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        canCreateFramework: permission(action: "core:framework:create")
        frameworks(first: 100)
          @connection(key: "FrameworksListQuery_frameworks") {
          __id
          edges {
            node {
              id
              canUpdate: permission(action: "core:framework:update")
              canDelete: permission(action: "core:framework:delete")
              ...FrameworksPageCardFragment
            }
          }
        }
      }
    }
  }
`;

const deleteFrameworkMutation = graphql`
  mutation FrameworkGraphDeleteMutation(
    $input: DeleteFrameworkInput!
    $connections: [ID!]!
  ) {
    deleteFramework(input: $input) {
      deletedFrameworkId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteFrameworkMutation = (
  framework: { id: string; name: string },
  connectionId: string,
) => {
  const { t } = useTranslation();
  const [commitDelete] = useMutationWithToasts(deleteFrameworkMutation, {
    errorMessage: t("frameworkGraph.errors.delete"),
    successMessage: t("frameworkGraph.messages.deleted"),
  });
  const confirm = useConfirm();

  return useCallback(
    (options?: { onSuccess?: () => void }) => {
      return confirm(
        () => {
          return commitDelete({
            variables: {
              input: {
                frameworkId: framework.id,
              },
              connections: [connectionId],
            },
            ...options,
          });
        },
        {
          message: t("frameworkGraph.deleteConfirmation", {
            name: framework.name,
          }),
        },
      );
    },
    [framework, connectionId, commitDelete, confirm, t],
  );
};

export const frameworkNodeQuery = graphql`
  query FrameworkGraphNodeQuery($frameworkId: ID!) {
    node(id: $frameworkId) {
      ... on Framework {
        id
        name
        ...FrameworkDetailPageFragment
      }
    }
  }
`;

export const frameworkControlNodeQuery = graphql`
  query FrameworkGraphControlNodeQuery($controlId: ID!) {
    node(id: $controlId) {
      ... on Control {
        id
        name
        sectionTitle
        description
        bestPractice
        notImplementedJustification
        maturityLevel
        canUpdate: permission(action: "core:control:update")
        canDelete: permission(action: "core:control:delete")
        canCreateMeasureMapping: permission(
          action: "core:control:create-measure-mapping"
        )
        canDeleteMeasureMapping: permission(
          action: "core:control:delete-measure-mapping"
        )
        canCreateDocumentMapping: permission(
          action: "core:control:create-document-mapping"
        )
        canDeleteDocumentMapping: permission(
          action: "core:control:delete-document-mapping"
        )
        canCreateAuditMapping: permission(
          action: "core:control:create-audit-mapping"
        )
        canDeleteAuditMapping: permission(
          action: "core:control:delete-audit-mapping"
        )
        canCreateObligationMapping: permission(
          action: "core:control:create-obligation-mapping"
        )
        canDeleteObligationMapping: permission(
          action: "core:control:delete-obligation-mapping"
        )
        ...FrameworkControlDialogFragment
        measures(first: 100)
          @connection(key: "FrameworkGraphControl_measures") {
          __id
          edges {
            node {
              id
              ...LinkedMeasuresCardFragment
            }
          }
        }
        documents(first: 100)
          @connection(key: "FrameworkGraphControl_documents") {
          __id
          edges {
            node {
              id
              ...LinkedDocumentsCardFragment
            }
          }
        }
        audits(first: 100) @connection(key: "FrameworkGraphControl_audits") {
          __id
          edges {
            node {
              id
              ...LinkedAuditsCardFragment
            }
          }
        }
        obligations(first: 100)
          @connection(key: "FrameworkGraphControl_obligations") {
          __id
          edges {
            node {
              id
              ...LinkedObligationsCardFragment
            }
          }
        }
      }
    }
  }
`;
