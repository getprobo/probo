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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { useLocation } from "react-router";

import type { DocumentDetailsPage_document$key } from "#/__generated__/core/DocumentDetailsPage_document.graphql";
import type { DocumentDetailsPageQuery } from "#/__generated__/core/DocumentDetailsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { DocumentApprovalsSection } from "./_components/DocumentApprovalsSection";
import { DocumentSignatoriesSection } from "./_components/DocumentSignatoriesSection";
import { DocumentControlsSection } from "./_components/DocumentControlsSection";
import { DocumentPropertiesSection } from "./_components/DocumentPropertiesSection";
import { DocumentTitleField } from "./_components/DocumentTitleField";
import { documentDetailsPage } from "./variants";

export const documentDetailsPageFragment = graphql`
  fragment DocumentDetailsPage_document on Document {
    id
    title
    organization {
      id
    }
    ...DocumentTitleField_document
    ...DocumentPropertiesSection_document
    ...DocumentControlsSection_document
    ...DocumentApprovalsSection_document
    ...DocumentSignatoriesSection_document
  }
`;

export const documentDetailsPageQuery = graphql`
  query DocumentDetailsPageQuery($documentId: ID!) {
    node(id: $documentId) @required(action: THROW) {
      __typename
      ... on Document {
        ...DocumentDetailsPage_document
      }
    }
  }
`;

interface DocumentDetailsPageProps {
  queryRef: PreloadedQuery<DocumentDetailsPageQuery>;
}

export function DocumentDetailsPage({ queryRef }: DocumentDetailsPageProps) {
  const data = usePreloadedQuery<DocumentDetailsPageQuery>(
    documentDetailsPageQuery,
    queryRef,
  );
  const { t } = useTranslation("organizations/documents");
  const organizationId = useOrganizationId();
  const location = useLocation();
  const documentKey = data.node.__typename === "Document" ? data.node : null;
  const document = useFragment<DocumentDetailsPage_document$key>(
    documentDetailsPageFragment,
    documentKey,
  );
  const { root, body, main, side } = documentDetailsPage();
  usePageTitle(document?.title ?? "");

  if (document == null || document.organization.id !== organizationId) {
    throw new NotFoundError(t("details.notFound"));
  }

  return (
    <div className={root()}>
      <Link
        to={{
          pathname: `/organizations/${organizationId}/governance/documents`,
          search: location.search,
        }}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className="self-start"
      >
        {t("details.back")}
      </Link>
      <div className={body()}>
        <div className={main()}>
          <DocumentTitleField documentKey={document} />
        </div>
        <div className={side()}>
          <DocumentPropertiesSection documentKey={document} />
          <DocumentApprovalsSection documentKey={document} />
          <DocumentSignatoriesSection documentKey={document} />
          <DocumentControlsSection documentKey={document} />
        </div>
      </div>
    </div>
  );
}
