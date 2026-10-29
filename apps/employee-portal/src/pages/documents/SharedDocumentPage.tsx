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
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { SharedDocumentPageQuery } from "#/__generated__/core/SharedDocumentPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { DocumentRequestPanel } from "#/pages/_components/DocumentRequestPanel";
import { DocumentWorkspace } from "#/pages/_components/DocumentWorkspace";
import { useExportEmployeeDocumentPdf } from "#/pages/_lib/useExportEmployeeDocumentPdf";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

const NS = "documents";

export const sharedDocumentPageQuery = graphql`
  query SharedDocumentPageQuery($documentId: ID!) @throwOnFieldError {
    document: node(id: $documentId) @required(action: THROW) {
      __typename
      ... on EmployeePortalDocument {
        title
        documentType
        publishedVersionId
      }
    }
  }
`;

interface SharedDocumentPageProps {
  queryRef: PreloadedQuery<SharedDocumentPageQuery>;
}

export function SharedDocumentPage({ queryRef }: SharedDocumentPageProps) {
  const { t } = useTranslation(NS);
  const { t: tApp } = useTranslation();
  const { employeePortalId } = useEmployeePortalRoute();
  const { document } = usePreloadedQuery<SharedDocumentPageQuery>(
    sharedDocumentPageQuery,
    queryRef,
  );
  if (document.__typename !== "EmployeePortalDocument") {
    throw new NotFoundError("shared document not found");
  }

  const dataUri = useExportEmployeeDocumentPdf(document.publishedVersionId);

  return (
    <DocumentWorkspace
      title={document.title}
      dataUri={dataUri}
      request={(
        <DocumentRequestPanel
          backLabel={t("title")}
          backTo={`/${employeePortalId}/documents`}
          title={document.title}
          detail={(
            <Text size={2} color="gold">
              {tApp(`documents.types.${document.documentType.toLowerCase()}`)}
            </Text>
          )}
        />
      )}
    />
  );
}
