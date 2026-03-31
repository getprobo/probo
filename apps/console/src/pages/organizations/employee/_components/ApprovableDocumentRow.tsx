// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import {
  formatDate,
  getDocumentClassificationLabel,
  getDocumentTypeLabel,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Td, Tr } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { ApprovableDocumentRowFragment$key } from "#/__generated__/core/ApprovableDocumentRowFragment.graphql";

const fragment = graphql`
  fragment ApprovableDocumentRowFragment on EmployeeDocument {
    id
    title
    documentType
    classification
    approvalState
    updatedAt
  }
`;

export function ApprovableDocumentRow({
  fKey,
  organizationId,
}: {
  fKey: ApprovableDocumentRowFragment$key;
  organizationId: string;
}) {
  const document = useFragment<ApprovableDocumentRowFragment$key>(fragment, fKey);
  const { __ } = useTranslate();

  const stateVariant = document.approvalState === "APPROVED"
    ? "success"
    : document.approvalState === "REJECTED"
      ? "danger"
      : "warning";

  const stateLabel = document.approvalState === "APPROVED"
    ? __("Approved")
    : document.approvalState === "REJECTED"
      ? __("Rejected")
      : __("Pending");

  return (
    <Tr to={`/organizations/${organizationId}/employee/approvals/${document.id}`}>
      <Td>{document.title}</Td>
      <Td className="w-48">
        {getDocumentTypeLabel(__, document.documentType)}
      </Td>
      <Td className="w-36">
        <Badge variant="neutral">
          {getDocumentClassificationLabel(__, document.classification)}
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
