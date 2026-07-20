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

import { getCompliancePageVisibilityOptions } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, DocumentTypeBadge, Field, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDocumentListItem_compliancePageFragment$key } from "#/__generated__/core/CompliancePageDocumentListItem_compliancePageFragment.graphql";
import type { CompliancePageDocumentListItem_documentFragment$key } from "#/__generated__/core/CompliancePageDocumentListItem_documentFragment.graphql";
import type { CompliancePageDocumentListItem_updateVisibilityMutation } from "#/__generated__/core/CompliancePageDocumentListItem_updateVisibilityMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { CompliancePageAliasField } from "../../_components/CompliancePageAliasField";

const compliancePageFragment = graphql`
  fragment CompliancePageDocumentListItem_compliancePageFragment on CompliancePortal {
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const documentFragment = graphql`
  fragment CompliancePageDocumentListItem_documentFragment on Document {
    id
    alias
    canSetAlias: permission(action: "resourcealias:alias:set")
    canRemoveAlias: permission(action: "resourcealias:alias:remove")
    compliancePageVisibility: compliancePortalVisibility
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
`;

const updateDocumentVisibilityMutation = graphql`
  mutation CompliancePageDocumentListItem_updateVisibilityMutation(
    $input: UpdateDocumentInput!
  ) {
    updateDocument(input: $input) {
      document {
        ...CompliancePageDocumentListItem_documentFragment
      }
    }
  }
`;

export function CompliancePageDocumentListItem(props: {
  compliancePageFragmentRef: CompliancePageDocumentListItem_compliancePageFragment$key;
  documentFragmentRef: CompliancePageDocumentListItem_documentFragment$key;
}) {
  const { compliancePageFragmentRef, documentFragmentRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const visibilityOptions = getCompliancePageVisibilityOptions(__);

  const compliancePage = useFragment<CompliancePageDocumentListItem_compliancePageFragment$key>(
    compliancePageFragment,
    compliancePageFragmentRef,
  );
  const document = useFragment<CompliancePageDocumentListItem_documentFragment$key>(
    documentFragment,
    documentFragmentRef,
  );
  const [updateDocumentVisibility, isUpdatingDocumentVisibility]
    = useMutation<CompliancePageDocumentListItem_updateVisibilityMutation>(
      updateDocumentVisibilityMutation,
      {
        successMessage: __("Document visibility updated successfully."),
        errorToast: __("Failed to update document visibility"),
      },
    );
  const handleVsibilityChange = useCallback(
    async (value: string) => {
      const stringValue = typeof value === "string" ? value : "";
      const typedValue = stringValue as "NONE" | "PRIVATE" | "PUBLIC";
      await updateDocumentVisibility({
        variables: {
          input: {
            id: document.id,
            compliancePortalVisibility: typedValue,
          },
        },
      });
    },
    [document.id, updateDocumentVisibility],
  );

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
        <CompliancePageAliasField
          resourceId={document.id}
          alias={document.alias}
          canSetAlias={document.canSetAlias}
          canRemoveAlias={document.canRemoveAlias}
        />
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={document.compliancePageVisibility}
          onValueChange={value => void handleVsibilityChange(value)}
          disabled={isUpdatingDocumentVisibility || !compliancePage.canUpdate}
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
    </Tr>
  );
}
