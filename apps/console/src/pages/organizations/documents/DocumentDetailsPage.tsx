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

import { usePageTitle } from "@probo/hooks";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation } from "react-router";

import type { DocumentDetailsPageQuery } from "#/__generated__/core/DocumentDetailsPageQuery.graphql";
import type { DocumentReviewListQuery } from "#/__generated__/core/DocumentReviewListQuery.graphql";
import type { DocumentVersionContentQuery } from "#/__generated__/core/DocumentVersionContentQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { DocumentDetails } from "./_components/DocumentDetails";
import { documentListSearch } from "./_lib/useDocumentVersionParam";

export const documentDetailsPageQuery = graphql`
  query DocumentDetailsPageQuery($documentId: ID!) {
    node(id: $documentId) @required(action: THROW) {
      __typename
      ... on Document {
        title
        organization {
          id
        }
        ...DocumentDetails_document
      }
    }
  }
`;

interface DocumentDetailsPageProps {
  queryRef: PreloadedQuery<DocumentDetailsPageQuery>;
  reviewQueryRef: PreloadedQuery<DocumentReviewListQuery> | null;
  versionQueryRef: PreloadedQuery<DocumentVersionContentQuery> | null;
}

export function DocumentDetailsPage({ queryRef, reviewQueryRef, versionQueryRef }: DocumentDetailsPageProps) {
  const data = usePreloadedQuery<DocumentDetailsPageQuery>(
    documentDetailsPageQuery,
    queryRef,
  );
  const { t } = useTranslation("organizations/documents");
  const organizationId = useOrganizationId();
  const location = useLocation();
  const document = data.node.__typename === "Document" ? data.node : null;
  usePageTitle(document?.title ?? "");

  if (document == null || document.organization.id !== organizationId) {
    throw new NotFoundError(t("details.notFound"));
  }

  return (
    <DocumentDetails
      documentKey={document}
      reviewQueryRef={reviewQueryRef}
      versionQueryRef={versionQueryRef}
      organizationId={organizationId}
      search={documentListSearch(location.search)}
    />
  );
}
