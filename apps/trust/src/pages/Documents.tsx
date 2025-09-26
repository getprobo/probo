import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { trustDocumentsQuery } from "/queries/TrustGraph";
import type { TrustGraphDocumentsQuery } from "/queries/__generated__/TrustGraphDocumentsQuery.graphql.ts";
import { groupBy, objectEntries } from "@probo/helpers";
import { documentTypeLabel } from "/helpers/documents";
import { useTranslate } from "@probo/i18n";
import { Table, Tbody, Th, Thead, Tr } from "@probo/ui";
import { Fragment } from "react";
import { DocumentRow } from "/components/DocumentRow";

type Props = {
  queryRef: PreloadedQuery<TrustGraphDocumentsQuery>;
};

export function Documents({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(trustDocumentsQuery, queryRef);
  const documents =
    data.trustCenterBySlug?.documents.edges.map((edge) => edge.node) ?? [];
  const documentsPerType = groupBy(documents, (document) =>
    documentTypeLabel(document.documentType, __)
  );
  return (
    <div>
      <h2 className="font-medium mb-1">{__("Documents")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {__("Security and compliance documentation:")}
      </p>
      <Table className="mb-8">
        {objectEntries(documentsPerType).map(([label, documents]) => (
          <Fragment key={label}>
            <Thead>
              <Tr className="bg-subtle">
                <Th colSpan={2}>{label}</Th>
              </Tr>
            </Thead>
            <Tbody>
              {documents.map((document) => (
                <DocumentRow key={document.id} document={document} />
              ))}
            </Tbody>
          </Fragment>
        ))}
      </Table>
    </div>
  );
}
