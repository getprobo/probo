import { groupBy, objectEntries } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Fragment } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";

import { DocumentRow } from "#/components/DocumentRow";
import { RowHeader } from "#/components/RowHeader";
import { Rows } from "#/components/Rows";
import { TrustCenterFileRow } from "#/components/TrustCenterFileRow";
import { documentTypeLabel } from "#/helpers/documents";
import type { TrustGraphCurrentDocumentsQuery } from "#/queries/__generated__/TrustGraphCurrentDocumentsQuery.graphql";
import { currentTrustDocumentsQuery } from "#/queries/TrustGraph";

type Props = {
  queryRef: PreloadedQuery<TrustGraphCurrentDocumentsQuery>;
};

export function DocumentsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery<TrustGraphCurrentDocumentsQuery>(
    currentTrustDocumentsQuery,
    queryRef,
  );
  const documents
    = data.currentTrustCenter?.documents.edges.map(edge => edge.node) ?? [];
  const files
    = data.currentTrustCenter?.trustCenterFiles.edges.map(edge => edge.node) ?? [];
  const documentsPerType = groupBy(documents, document =>
    documentTypeLabel(document.documentType, __),
  );
  const filesPerCategory = groupBy(files, file => file.category);
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
            {documents.map(document => (
              <DocumentRow key={document.id} document={document} />
            ))}
          </Fragment>
        ))}
        {objectEntries(filesPerCategory).map(([category, files]) => (
          <Fragment key={category}>
            <RowHeader>{category}</RowHeader>
            {files.map(file => (
              <TrustCenterFileRow key={file.id} file={file} />
            ))}
          </Fragment>
        ))}
      </Rows>
    </div>
  );
}
