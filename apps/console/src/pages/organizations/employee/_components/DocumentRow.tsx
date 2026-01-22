import {
  formatDate,
  getDocumentClassificationLabel,
  getDocumentTypeLabel,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Td, Tr } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { DocumentRowFragment$key } from "#/__generated__/core/DocumentRowFragment.graphql";

const fragment = graphql`
  fragment DocumentRowFragment on SignableDocument {
    id
    title
    documentType
    classification
    signed
    updatedAt
  }
`;

export function DocumentRow({
  fKey,
  organizationId,
}: {
  fKey: DocumentRowFragment$key;
  organizationId: string;
}) {
  const document = useFragment<DocumentRowFragment$key>(fragment, fKey);
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/employee/${document.id}`}>
      <Td className="min-w-0 pr-12">{document.title}</Td>
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
        <Badge variant={document.signed ? "success" : "danger"}>
          {document.signed ? __("Yes") : __("No")}
        </Badge>
      </Td>
    </Tr>
  );
}
