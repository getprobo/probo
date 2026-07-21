// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { DocumentGraphBulkExportDocumentsMutation } from "#/__generated__/core/DocumentGraphBulkExportDocumentsMutation.graphql";
import type { DocumentGraphDeleteMutation } from "#/__generated__/core/DocumentGraphDeleteMutation.graphql";

import { useMutationWithToasts } from "../useMutationWithToasts";

export const DocumentsConnectionKey = "DocumentsListQuery_documents";

const deleteDocumentMutation = graphql`
  mutation DocumentGraphDeleteMutation(
    $input: DeleteDocumentInput!
    $connections: [ID!]!
  ) {
    deleteDocument(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteDocumentMutation() {
  const { t } = useTranslation();

  return useMutationWithToasts<DocumentGraphDeleteMutation>(
    deleteDocumentMutation,
    {
      successMessage: t("documentGraph.messages.deleted"),
      errorMessage: t("documentGraph.errors.delete"),
    },
  );
}

const bulkDeleteDocumentsMutation = graphql`
  mutation DocumentGraphBulkDeleteDocumentsMutation(
    $input: BulkDeleteDocumentsInput!
  ) {
    bulkDeleteDocuments(input: $input) {
      deletedDocumentIds
    }
  }
`;

export function useBulkDeleteDocumentsMutation() {
  const { t } = useTranslation();

  return useMutationWithToasts(bulkDeleteDocumentsMutation, {
    successMessage: t("documentGraph.messages.bulkDeleted"),
    errorMessage: t("documentGraph.errors.bulkDelete"),
  });
}

const bulkExportDocumentsMutation = graphql`
  mutation DocumentGraphBulkExportDocumentsMutation(
    $input: BulkExportDocumentsInput!
  ) {
    bulkExportDocuments(input: $input) {
      exportJobId
    }
  }
`;

export function useBulkExportDocumentsMutation() {
  const { t } = useTranslation();

  return useMutationWithToasts<DocumentGraphBulkExportDocumentsMutation>(
    bulkExportDocumentsMutation,
    {
      successMessage: t("documentGraph.messages.exportStarted"),
      errorMessage: t("documentGraph.errors.export"),
    },
  );
}
