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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { useMutation } from "#/lib/relay/useMutation";

import { openExportedFile } from "../_lib/openExportedFile";

import type { DocumentListItem_document$key } from "./__generated__/DocumentListItem_document.graphql";
import type { DocumentListItemExportMutation } from "./__generated__/DocumentListItemExportMutation.graphql";
import { DocumentAccessAction } from "./DocumentAccessAction";
import { documentListItem } from "./variants";

const documentListItemFragment = graphql`
  fragment DocumentListItem_document on Document @throwOnFieldError {
    id
    title
    documentType
    isUserAuthorized
    access {
      status
    }
  }
`;

const exportDocumentMutation = graphql`
  mutation DocumentListItemExportMutation($input: ExportDocumentPDFInput!) {
    exportDocumentPDF(input: $input) {
      data
    }
  }
`;

interface DocumentListItemProps {
  documentKey: DocumentListItem_document$key;
}

// A single Probo document row: title, its document type, and an access action
// that opens the exported PDF when the viewer is authorized.
export function DocumentListItem({ documentKey }: DocumentListItemProps) {
  const { t } = useTranslation("documents");
  const document = useFragment(documentListItemFragment, documentKey);
  const [exportDocument, isExporting] = useMutation<DocumentListItemExportMutation>(exportDocumentMutation);
  const { root, content } = documentListItem();

  const handleView = () => {
    exportDocument({
      variables: { input: { documentId: document.id } },
      onCompleted: response => openExportedFile(response.exportDocumentPDF.data),
    }).catch(() => {
      // The mutation failure is already surfaced through a toast.
    });
  };

  return (
    <div className={root()}>
      <div className={content()}>
        <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
          {document.title}
        </Text>
        <Text size={1} color="gold" className="truncate">
          {t(`types.${document.documentType}`)}
        </Text>
      </div>
      <DocumentAccessAction
        isAuthorized={document.isUserAuthorized}
        requested={document.access?.status === "REQUESTED"}
        onView={handleView}
        isViewing={isExporting}
      />
    </div>
  );
}
