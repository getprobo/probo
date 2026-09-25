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

import { useEffect, useState } from "react";
import { graphql } from "react-relay";

import type { useExportEmployeeDocumentPdfMutation } from "#/__generated__/core/useExportEmployeeDocumentPdfMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const exportEmployeeDocumentPdfMutation = graphql`
  mutation useExportEmployeeDocumentPdfMutation(
    $input: ExportEmployeeDocumentVersionPDFInput!
  ) {
    exportEmployeeDocumentVersionPDF(input: $input) {
      data
    }
  }
`;

type ExportedPdf = {
  documentVersionId: string;
  dataUri: string;
};

// Loads the latest employee PDF for a document version and keeps the data URI
// until the version id changes.
export function useExportEmployeeDocumentPdf(documentVersionId: string | null): string | null {
  const [exportPdf] = useMutation<useExportEmployeeDocumentPdfMutation>(
    exportEmployeeDocumentPdfMutation,
  );
  const [exported, setExported] = useState<ExportedPdf | null>(null);

  useEffect(() => {
    if (documentVersionId == null) {
      return;
    }

    let cancelled = false;

    void exportPdf({
      variables: { input: { documentVersionId } },
    }).then((response) => {
      if (!cancelled) {
        setExported({
          documentVersionId,
          dataUri: response.exportEmployeeDocumentVersionPDF.data,
        });
      }
    }).catch(() => {});

    return () => {
      cancelled = true;
    };
  }, [documentVersionId, exportPdf]);

  if (documentVersionId == null || exported?.documentVersionId !== documentVersionId) {
    return null;
  }

  return exported.dataUri;
}
