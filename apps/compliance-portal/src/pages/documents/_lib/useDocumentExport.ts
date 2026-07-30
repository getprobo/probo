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

import { useEffect, useRef, useState } from "react";
import { graphql } from "react-relay";

import { useMutation } from "#/lib/relay/useMutation";

import type { useDocumentExportDocumentMutation } from "./__generated__/useDocumentExportDocumentMutation.graphql";
import type { useDocumentExportFileMutation } from "./__generated__/useDocumentExportFileMutation.graphql";
import type { useDocumentExportReportMutation } from "./__generated__/useDocumentExportReportMutation.graphql";

export type DocumentKind = "Document" | "CompliancePortalFile" | "AuditReport";

const exportDocumentMutation = graphql`
  mutation useDocumentExportDocumentMutation($input: ExportDocumentPDFInput!) {
    exportDocumentPDF(input: $input) {
      data
    }
  }
`;

const exportFileMutation = graphql`
  mutation useDocumentExportFileMutation($input: ExportCompliancePortalFileInput!) {
    exportCompliancePortalFile(input: $input) {
      data
    }
  }
`;

const exportReportMutation = graphql`
  mutation useDocumentExportReportMutation($input: ExportReportPDFInput!) {
    exportReportPDF(input: $input) {
      data
    }
  }
`;

interface DocumentExportState {
  // The exported base64 data URI, or null while it is still loading.
  dataUri: string | null;
  isExporting: boolean;
}

// Exports the aliased node's (watermarked) bytes for the viewer. Fires once per
// (kind, id) while enabled. Auth / full-name / NDA gates and error toasts are
// handled by useMutation; this hook only applies the bytes on success. Mutate
// functions are read from a ref so their identity churn cannot re-trigger the
// effect.
export function useDocumentExport(kind: DocumentKind, id: string, enabled: boolean): DocumentExportState {
  const [exportDocument, isExportingDocument] = useMutation<useDocumentExportDocumentMutation>(
    exportDocumentMutation,
  );
  const [exportFile, isExportingFile] = useMutation<useDocumentExportFileMutation>(
    exportFileMutation,
  );
  const [exportReport, isExportingReport] = useMutation<useDocumentExportReportMutation>(
    exportReportMutation,
  );

  const latest = useRef({ exportDocument, exportFile, exportReport });
  useEffect(() => {
    latest.current = { exportDocument, exportFile, exportReport };
  });

  const [dataUri, setDataUri] = useState<string | null>(null);

  // Drop the previous document's bytes as soon as the target changes so a stale
  // preview is never shown for the new document.
  const [loadedId, setLoadedId] = useState(id);
  if (loadedId !== id) {
    setLoadedId(id);
    setDataUri(null);
  }

  useEffect(() => {
    if (!enabled) {
      return;
    }

    let cancelled = false;
    const {
      exportDocument: exportDoc,
      exportFile: exportFil,
      exportReport: exportRep,
    } = latest.current;

    const run = async () => {
      try {
        let data: string;
        switch (kind) {
          case "Document": {
            const response = await exportDoc({
              variables: { input: { documentId: id } },
            });
            data = response.exportDocumentPDF.data;
            break;
          }
          case "CompliancePortalFile": {
            const response = await exportFil({
              variables: { input: { compliancePortalFileId: id } },
            });
            data = response.exportCompliancePortalFile.data;
            break;
          }
          case "AuditReport": {
            const response = await exportRep({
              variables: { input: { reportId: id } },
            });
            data = response.exportReportPDF.data;
            break;
          }
        }
        if (!cancelled) {
          setDataUri(data);
        }
      } catch {
        // Gate redirects and error toasts are handled by useMutation.
      }
    };

    void run();

    return () => {
      cancelled = true;
    };
  }, [enabled, kind, id]);

  return { dataUri, isExporting: isExportingDocument || isExportingFile || isExportingReport };
}
