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

import {
  formatDate,
  getDocumentClassificationLabel,
  getDocumentTypeLabel,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Td, Tr } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { ApprovableDocumentRowFragment$key } from "#/__generated__/core/ApprovableDocumentRowFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment ApprovableDocumentRowFragment on EmployeeDocument {
    id
    title
    approvalState
    updatedAt
    lastVersion: versions(first: 1 orderBy: { field: CREATED_AT direction: DESC }) {
      edges {
        node {
          documentType
          classification
        }
      }
    }
  }
`;

export function ApprovableDocumentRow({
  fKey,
}: {
  fKey: ApprovableDocumentRowFragment$key;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const document = useFragment<ApprovableDocumentRowFragment$key>(fragment, fKey);

  const lastVersionEdge = document.lastVersion.edges[0];
  if (!lastVersionEdge) return null;
  const lastVersion = lastVersionEdge.node;

  const stateVariant = document.approvalState === "APPROVED"
    ? "success"
    : document.approvalState === "REJECTED"
      ? "danger"
      : document.approvalState === "VOIDED"
        ? "neutral"
        : "warning";

  const stateLabel = document.approvalState === "APPROVED"
    ? __("Approved")
    : document.approvalState === "REJECTED"
      ? __("Rejected")
      : document.approvalState === "VOIDED"
        ? __("No longer required")
        : __("Pending");

  return (
    <Tr to={`/organizations/${organizationId}/employee/approvals/${document.id}`}>
      <Td>{document.title}</Td>
      <Td className="w-48">
        {getDocumentTypeLabel(__, lastVersion.documentType)}
      </Td>
      <Td className="w-36">
        <Badge variant="neutral">
          {getDocumentClassificationLabel(__, lastVersion.classification)}
        </Badge>
      </Td>
      <Td className="w-40">{formatDate(document.updatedAt)}</Td>
      <Td className="w-32">
        <Badge variant={stateVariant}>
          {stateLabel}
        </Badge>
      </Td>
    </Tr>
  );
}
