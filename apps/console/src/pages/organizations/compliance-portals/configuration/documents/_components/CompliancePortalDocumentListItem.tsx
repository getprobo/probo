// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { getCompliancePortalLinkedVisibilityOptions } from "@probo/helpers";
import { Badge, Button, DocumentTypeBadge, Field, IconCrossLargeX, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type {
  CompliancePortalDocumentListItem_catalogDocumentFragment$key,
} from "#/__generated__/core/CompliancePortalDocumentListItem_catalogDocumentFragment.graphql";
import type {
  CompliancePortalDocumentListItem_compliancePortalFragment$key,
} from "#/__generated__/core/CompliancePortalDocumentListItem_compliancePortalFragment.graphql";
import type {
  CompliancePortalDocumentListItem_removeMutation,
} from "#/__generated__/core/CompliancePortalDocumentListItem_removeMutation.graphql";
import type {
  CompliancePortalDocumentListItem_updateVisibilityMutation,
} from "#/__generated__/core/CompliancePortalDocumentListItem_updateVisibilityMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { CompliancePortalAliasField } from "../../_components/CompliancePortalAliasField";

const compliancePortalFragment = graphql`
  fragment CompliancePortalDocumentListItem_compliancePortalFragment on CompliancePortal {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const catalogDocumentFragment = graphql`
  fragment CompliancePortalDocumentListItem_catalogDocumentFragment on CompliancePortalDocument {
    id
    visibility
    document {
      id
      alias
      canSetAlias: permission(action: "resourcealias:alias:set")
      canRemoveAlias: permission(action: "resourcealias:alias:remove")
      latestPublishedVersion: versions(
        first: 1
        orderBy: { field: CREATED_AT, direction: DESC }
        filter: { statuses: [PUBLISHED] }
      ) {
        edges {
          node {
            title
            documentType
          }
        }
      }
    }
  }
`;

const updateDocumentVisibilityMutation = graphql`
  mutation CompliancePortalDocumentListItem_updateVisibilityMutation(
    $input: UpdateCompliancePortalDocumentVisibilityInput!
  ) {
    updateCompliancePortalDocumentVisibility(input: $input) {
      catalogDocument {
        id
        visibility
      }
    }
  }
`;

const removeDocumentMutation = graphql`
  mutation CompliancePortalDocumentListItem_removeMutation(
    $input: DeleteCompliancePortalDocumentInput!
    $connections: [ID!]!
  ) {
    deleteCompliancePortalDocument(input: $input) {
      deletedCompliancePortalDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

export function CompliancePortalDocumentListItem(props: {
  compliancePortalFragmentRef: CompliancePortalDocumentListItem_compliancePortalFragment$key;
  catalogDocumentFragmentRef: CompliancePortalDocumentListItem_catalogDocumentFragment$key;
  documentsConnectionId: DataID;
}) {
  const { compliancePortalFragmentRef, catalogDocumentFragmentRef, documentsConnectionId } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");
  const visibilityOptions = getCompliancePortalLinkedVisibilityOptions(t);

  const compliancePortal = useFragment<CompliancePortalDocumentListItem_compliancePortalFragment$key>(
    compliancePortalFragment,
    compliancePortalFragmentRef,
  );
  const catalogDocument = useFragment<CompliancePortalDocumentListItem_catalogDocumentFragment$key>(
    catalogDocumentFragment,
    catalogDocumentFragmentRef,
  );
  const document = catalogDocument.document;

  const [updateDocumentVisibility, isUpdatingDocumentVisibility]
    = useMutation<CompliancePortalDocumentListItem_updateVisibilityMutation>(
      updateDocumentVisibilityMutation,
      {
        successMessage: t("documentListItem.messages.visibilityUpdated"),
        errorToast: t("documentListItem.errors.updateVisibility"),
      },
    );

  const [removeDocument, isRemoving]
    = useMutation<CompliancePortalDocumentListItem_removeMutation>(
      removeDocumentMutation,
      {
        successMessage: t("documentListItem.messages.removed"),
        errorToast: t("documentListItem.errors.remove"),
      },
    );

  const handleVisibilityChange = useCallback(
    async (value: string) => {
      const typedValue = value === "PUBLIC" ? "PUBLIC" : "RESTRICTED";
      await updateDocumentVisibility({
        variables: {
          input: {
            compliancePortalId: compliancePortal.id,
            documentId: document.id,
            compliancePortalVisibility: typedValue,
          },
        },
      });
    },
    [compliancePortal.id, document.id, updateDocumentVisibility],
  );

  const handleRemove = useCallback(async () => {
    await removeDocument({
      variables: {
        connections: [documentsConnectionId],
        input: {
          id: catalogDocument.id,
        },
      },
    });
  }, [catalogDocument.id, documentsConnectionId, removeDocument]);

  const latestVersion = document.latestPublishedVersion.edges[0]?.node;
  const versionTitle = latestVersion?.title;

  return (
    <Tr to={`/organizations/${organizationId}/documents/${document.id}`}>
      <Td>
        <div className="flex gap-4 items-center">{versionTitle}</div>
      </Td>
      <Td>
        {latestVersion && <DocumentTypeBadge type={latestVersion.documentType} />}
      </Td>
      <Td noLink>
        <CompliancePortalAliasField
          resourceId={document.id}
          alias={document.alias}
          canSetAlias={document.canSetAlias}
          canRemoveAlias={document.canRemoveAlias}
        />
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={catalogDocument.visibility}
          onValueChange={value => void handleVisibilityChange(value)}
          disabled={isUpdatingDocumentVisibility || !compliancePortal.canUpdate}
          className="w-[105px]"
        >
          {visibilityOptions.map(option => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>{option.label}</Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
      <Td noLink width={48}>
        {compliancePortal.canUpdate && (
          <Button
            variant="tertiary"
            icon={IconCrossLargeX}
            aria-label={t("documentListItem.actions.remove")}
            disabled={isRemoving}
            onClick={() => void handleRemove()}
          />
        )}
      </Td>
    </Tr>
  );
}
