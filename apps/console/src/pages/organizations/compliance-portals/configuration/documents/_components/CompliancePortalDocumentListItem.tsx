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
import { Badge, Checkbox, DocumentTypeBadge, Field, Option, Td, Tr } from "@probo/ui";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type {
  CompliancePortalDocumentListItem_compliancePortal$key,
} from "#/__generated__/core/CompliancePortalDocumentListItem_compliancePortal.graphql";
import type {
  CompliancePortalDocumentListItem_document$key,
} from "#/__generated__/core/CompliancePortalDocumentListItem_document.graphql";
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
  fragment CompliancePortalDocumentListItem_compliancePortal on CompliancePortal {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const documentFragment = graphql`
  fragment CompliancePortalDocumentListItem_document on Document
  @argumentDefinitions(compliancePortalId: { type: "ID!" }) {
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
    compliancePortalDocument(compliancePortalId: $compliancePortalId) {
      id
      visibility
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
        document {
          id
        }
      }
    }
  }
`;

const removeDocumentMutation = graphql`
  mutation CompliancePortalDocumentListItem_removeMutation(
    $input: DeleteCompliancePortalDocumentInput!
  ) {
    deleteCompliancePortalDocument(input: $input) {
      deletedCompliancePortalDocumentId @deleteRecord
    }
  }
`;

export function CompliancePortalDocumentListItem(props: {
  compliancePortalKey: CompliancePortalDocumentListItem_compliancePortal$key;
  documentKey: CompliancePortalDocumentListItem_document$key;
}) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation("organizations/compliance-portals");
  const visibilityOptions = getCompliancePortalLinkedVisibilityOptions(t);

  const compliancePortal = useFragment<CompliancePortalDocumentListItem_compliancePortal$key>(
    compliancePortalFragment,
    props.compliancePortalKey,
  );
  const document = useFragment<CompliancePortalDocumentListItem_document$key>(
    documentFragment,
    props.documentKey,
  );
  const catalogDocument = document.compliancePortalDocument;
  const serverLinked = catalogDocument !== null;
  const [pendingLinked, setPendingLinked] = useState<boolean | null>(null);
  const isLinked = pendingLinked ?? serverLinked;

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
      if (!catalogDocument) {
        return;
      }

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
    [catalogDocument, compliancePortal.id, document.id, updateDocumentVisibility],
  );

  const handleLinkedChange = useCallback(
    async (checked: boolean) => {
      if (!compliancePortal.canUpdate || checked === isLinked) {
        return;
      }

      setPendingLinked(checked);

      try {
        if (checked) {
          await updateDocumentVisibility({
            variables: {
              input: {
                compliancePortalId: compliancePortal.id,
                documentId: document.id,
                compliancePortalVisibility: "RESTRICTED",
              },
            },
            updater: (store) => {
              const payload = store.getRootField("updateCompliancePortalDocumentVisibility");
              const link = payload?.getLinkedRecord("catalogDocument");
              const documentRecord = store.get(document.id);
              if (link && documentRecord) {
                documentRecord.setLinkedRecord(
                  link,
                  "compliancePortalDocument",
                  { compliancePortalId: compliancePortal.id },
                );
              }
            },
          });
          setPendingLinked(null);
          return;
        }

        if (!catalogDocument) {
          setPendingLinked(null);
          return;
        }

        await removeDocument({
          variables: {
            input: {
              id: catalogDocument.id,
            },
          },
          updater: (store) => {
            store.get(document.id)?.setValue(
              null,
              "compliancePortalDocument",
              { compliancePortalId: compliancePortal.id },
            );
          },
        });
        setPendingLinked(null);
      } catch {
        setPendingLinked(null);
      }
    },
    [
      catalogDocument,
      compliancePortal.canUpdate,
      compliancePortal.id,
      document.id,
      isLinked,
      removeDocument,
      updateDocumentVisibility,
    ],
  );

  const isMutating = isUpdatingDocumentVisibility || isRemoving;

  const latestVersion = document.latestPublishedVersion.edges[0]?.node;
  const versionTitle = latestVersion?.title;

  return (
    <Tr to={`/organizations/${organizationId}/documents/${document.id}`}>
      <Td noLink>
        <Checkbox
          checked={isLinked}
          onChange={checked => void handleLinkedChange(checked)}
          disabled={isMutating || !compliancePortal.canUpdate}
          aria-label={t("documentListItem.actions.toggle", {
            title: versionTitle,
          })}
        />
      </Td>
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
      <Td noLink width={130}>
        <Field
          type="select"
          value={catalogDocument?.visibility ?? "RESTRICTED"}
          onValueChange={value => void handleVisibilityChange(value)}
          disabled={!isLinked || isMutating || !compliancePortal.canUpdate}
          className="w-26.25"
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
    </Tr>
  );
}
