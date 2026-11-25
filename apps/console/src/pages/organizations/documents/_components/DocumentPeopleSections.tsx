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

import { CardSkeleton } from "@probo/ui/src/v2/Card/CardSkeleton";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Suspense } from "react";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";

import type { DocumentPeopleSections_document$key } from "#/__generated__/core/DocumentPeopleSections_document.graphql";
import type { DocumentPeopleSections_version$key } from "#/__generated__/core/DocumentPeopleSections_version.graphql";
import type { DocumentVersionContentQuery } from "#/__generated__/core/DocumentVersionContentQuery.graphql";

import { DocumentApprovalsSection } from "./DocumentApprovalsSection";
import { DocumentSignatoriesSection } from "./DocumentSignatoriesSection";
import { documentVersionContentQuery } from "./DocumentVersionList/DocumentVersionContent";

const documentPeopleSectionsFragment = graphql`
  fragment DocumentPeopleSections_document on Document {
    id
    ...DocumentApprovalsSection_document
    ...DocumentSignatoriesSection_document
  }
`;

const documentPeopleSectionsVersionFragment = graphql`
  fragment DocumentPeopleSections_version on DocumentVersion {
    document {
      id
    }
    ...DocumentApprovalsSectionVersion_version
    ...DocumentSignatoriesSectionList_version @arguments(filter: {}, baseFilter: {})
  }
`;

interface DocumentPeopleSectionsProps {
  documentKey: DocumentPeopleSections_document$key;
  // Whether a previous version is on screen. Its query may still be loading.
  viewingVersion: boolean;
  versionQueryRef: PreloadedQuery<DocumentVersionContentQuery> | null;
}

// The approvers and signatories of the version on screen: the live
// document's, or those of the previous version shown in its place.
export function DocumentPeopleSections({ documentKey, viewingVersion, versionQueryRef }: DocumentPeopleSectionsProps) {
  const document = useFragment(documentPeopleSectionsFragment, documentKey);

  if (!viewingVersion) {
    return (
      <>
        <DocumentApprovalsSection documentKey={document} />
        <DocumentSignatoriesSection documentKey={document} />
      </>
    );
  }

  if (versionQueryRef == null) {
    return <DocumentPeopleSectionsSkeleton />;
  }

  return (
    <ErrorBoundary fallback={null}>
      <Suspense fallback={<DocumentPeopleSectionsSkeleton />}>
        <DocumentPeopleSectionsLoaded documentKey={documentKey} queryRef={versionQueryRef} />
      </Suspense>
    </ErrorBoundary>
  );
}

function DocumentPeopleSectionsLoaded({
  documentKey,
  queryRef,
}: {
  documentKey: DocumentPeopleSections_document$key;
  queryRef: PreloadedQuery<DocumentVersionContentQuery>;
}) {
  const data = usePreloadedQuery<DocumentVersionContentQuery>(documentVersionContentQuery, queryRef);

  return (
    <DocumentPeopleSectionsVersion
      documentKey={documentKey}
      versionKey={data.node.__typename === "DocumentVersion" ? data.node : null}
    />
  );
}

function DocumentPeopleSectionsVersion({
  documentKey,
  versionKey,
}: {
  documentKey: DocumentPeopleSections_document$key;
  versionKey: DocumentPeopleSections_version$key | null;
}) {
  const document = useFragment(documentPeopleSectionsFragment, documentKey);
  const version = useFragment(documentPeopleSectionsVersionFragment, versionKey);

  // The editor falls back to the live document for a version of another
  // document, and so do the cards.
  const shown = version != null && version.document.id === document.id ? version : null;

  return (
    <>
      <DocumentApprovalsSection documentKey={document} versionKey={shown} />
      <DocumentSignatoriesSection documentKey={document} versionKey={shown} />
    </>
  );
}

function DocumentPeopleSectionsSkeleton() {
  return (
    <>
      <CardSkeleton size={4} />
      <CardSkeleton size={4} />
    </>
  );
}
