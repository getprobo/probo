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

import { useMemo, useState } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useParams } from "react-router";

import type { SignatureDocumentPageQuery } from "#/__generated__/core/SignatureDocumentPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { DocumentVersionHistory } from "#/pages/_components/DocumentVersionHistory";
import { DocumentWorkspace } from "#/pages/_components/DocumentWorkspace";
import {
  useDocumentQueue,
  useDocumentQueueActive,
  useSyncDocumentQueue,
} from "#/pages/_lib/DocumentQueueContext";
import { useExportEmployeeDocumentPdf } from "#/pages/_lib/useExportEmployeeDocumentPdf";

import { SignatureRequestPanel } from "./_components/SignatureRequestPanel";
import { useSignDocument } from "./_lib/useSignDocument";

export const signatureDocumentPageQuery = graphql`
  query SignatureDocumentPageQuery(
    $organizationId: ID!
    $documentId: ID!
    $first: Int!
    $versionsFirst: Int!
  ) @throwOnFieldError {
    viewer @required(action: THROW) {
      signableDocument(id: $documentId) {
        id
        title
        signed
        latestVersion: versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              consentText
              signed
            }
          }
        }
      }
      ...DocumentVersionHistory_signatures @arguments(documentId: $documentId, first: $versionsFirst)
      pendingQueue: signableDocuments(
        organizationId: $organizationId
        first: $first
        filter: { signed: false }
        orderBy: { field: UPDATED_AT, direction: DESC }
      ) {
        totalCount
        pageInfo {
          hasNextPage
          endCursor
        }
        edges {
          node {
            id
          }
        }
      }
    }
  }
`;

interface SignatureDocumentPageProps {
  queryRef: PreloadedQuery<SignatureDocumentPageQuery>;
}

export function SignatureDocumentPage({ queryRef }: SignatureDocumentPageProps) {
  const { documentId } = useParams();
  const { snapshot, advancing, goForward, close } = useDocumentQueue();
  const queueActive = useDocumentQueueActive();
  const data = usePreloadedQuery<SignatureDocumentPageQuery>(
    signatureDocumentPageQuery,
    queryRef,
  );

  const document = data.viewer.signableDocument;
  if (document == null || documentId == null) {
    throw new NotFoundError("signable document not found");
  }

  const version = document.latestVersion.edges[0]?.node;
  if (version == null) {
    throw new NotFoundError("document version not found");
  }

  const [selected, setSelected] = useState({
    documentId: document.id,
    versionId: version.id,
  });
  if (selected.documentId !== document.id) {
    setSelected({ documentId: document.id, versionId: version.id });
  }
  const selectedVersionId = selected.documentId === document.id
    ? selected.versionId
    : version.id;

  const pendingPage = useMemo(() => ({
    ids: data.viewer.pendingQueue.edges.map(({ node }) => node.id),
    totalCount: data.viewer.pendingQueue.totalCount,
    endCursor: data.viewer.pendingQueue.pageInfo.endCursor ?? null,
    hasNextPage: data.viewer.pendingQueue.pageInfo.hasNextPage,
  }), [data.viewer.pendingQueue]);

  useSyncDocumentQueue({
    kind: "signatures",
    documentId,
    isPending: document.signed !== true,
    pendingPage,
  });

  const [signDocument, isSigning] = useSignDocument(document.id);
  const dataUri = useExportEmployeeDocumentPdf(selectedVersionId);

  const index = snapshot?.ids.indexOf(documentId) ?? -1;
  const hasNext = snapshot != null
    && ((index >= 0 && index < snapshot.ids.length - 1) || snapshot.hasNextPage);

  return (
    <DocumentWorkspace
      title={document.title}
      dataUri={dataUri}
      history={(
        <DocumentVersionHistory
          viewerKey={data.viewer}
          kind="signatures"
          selectedVersionId={selectedVersionId}
          onSelect={(versionId) => {
            setSelected({ documentId: document.id, versionId });
          }}
        />
      )}
      request={(
        <SignatureRequestPanel
          title={document.title}
          signed={document.signed === true}
          consentText={version.consentText}
          queueActive={queueActive}
          hasNext={hasNext}
          busy={isSigning}
          advancing={advancing}
          onSign={() => {
            void signDocument(version.id);
          }}
          onNext={goForward}
          onFinish={() => close("signatures")}
        />
      )}
    />
  );
}
