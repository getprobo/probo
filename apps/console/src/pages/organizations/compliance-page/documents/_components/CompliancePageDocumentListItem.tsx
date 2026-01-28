import { getTrustCenterVisibilityOptions } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, DocumentTypeBadge, DocumentVersionBadge, Field, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDocumentListItem_compliancePageFragment$key } from "#/__generated__/core/CompliancePageDocumentListItem_compliancePageFragment.graphql";
import type { CompliancePageDocumentListItem_documentFragment$key } from "#/__generated__/core/CompliancePageDocumentListItem_documentFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const compliancePageFragment = graphql`
  fragment CompliancePageDocumentListItem_compliancePageFragment on TrustCenter {
    canUpdate: permission(action: "core:trust-center:update")
  }
`;

const documentFragment = graphql`
  fragment CompliancePageDocumentListItem_documentFragment on Document {
    id
    documentType
    title
    trustCenterVisibility
    lastVersion: versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
      edges {
        node {
          id
          status
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
  const visibilityOptions = getTrustCenterVisibilityOptions(__);

  const compliancePage = useFragment<CompliancePageDocumentListItem_compliancePageFragment$key>(
    compliancePageFragment,
    compliancePageFragmentRef,
  );
  const document = useFragment<CompliancePageDocumentListItem_documentFragment$key>(
    documentFragment,
    documentFragmentRef,
  );
  const [updateDocumentVisibility, isUpdatingDocumentVisibility] = useMutationWithToasts(
    updateDocumentVisibilityMutation,
    {
      successMessage: __("Document visibility updated successfully."),
      errorMessage: __("Failed to update document visibility"),
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
            trustCenterVisibility: typedValue,
          },
        },
      });
    },
    [document.id, updateDocumentVisibility],
  );

  return (
    <Tr to={`/organizations/${organizationId}/documents/${document.id}`}>
      <Td>
        <div className="flex gap-4 items-center">{document.title}</div>
      </Td>
      <Td>
        <DocumentTypeBadge type={document.documentType} />
      </Td>
      <Td>
        <DocumentVersionBadge
          state={document.lastVersion.edges[0].node.status}
        />
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={document.trustCenterVisibility}
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
