// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
