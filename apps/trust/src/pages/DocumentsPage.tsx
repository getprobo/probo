import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { currentTrustDocumentsQuery } from "/queries/TrustGraph";
import type { TrustGraphCurrentDocumentsQuery } from "/queries/__generated__/TrustGraphCurrentDocumentsQuery.graphql.ts";
import { groupBy, objectEntries } from "@probo/helpers";
import { documentTypeLabel } from "/helpers/documents";
import { useTranslate } from "@probo/i18n";
import { Fragment } from "react";
import { DocumentRow } from "/components/DocumentRow";
import { Rows } from "/components/Rows.tsx";
import { RowHeader } from "/components/RowHeader.tsx";

type Props = {
  queryRef: PreloadedQuery<TrustGraphCurrentDocumentsQuery>;
};

export function DocumentsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery<TrustGraphCurrentDocumentsQuery>(
    currentTrustDocumentsQuery,
    queryRef
  );
  const documents =
    data.currentTrustCenter?.documents.edges.map((edge) => edge.node) ?? [];
  const documentsPerType = groupBy(documents, (document) =>
    documentTypeLabel(document.documentType, __)
  );
  return (
    <div>
      <h2 className="font-medium mb-1">{__("Documents")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {__("Security and compliance documentation:")}
      </p>
      <Rows className="mb-8">
        {objectEntries(documentsPerType).map(([label, documents]) => (
          <Fragment key={label}>
            <RowHeader>{label}</RowHeader>
            {documents.map((document) => (
              <DocumentRow key={document.id} document={document} />
            ))}
          </Fragment>
        ))}
      </Rows>
    </div>
  );
}
