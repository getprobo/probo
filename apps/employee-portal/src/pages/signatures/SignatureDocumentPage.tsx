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

import { useMemo } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useParams } from "react-router";

import type { SignatureDocumentPageQuery } from "#/__generated__/core/SignatureDocumentPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
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
  query SignatureDocumentPageQuery($organizationId: ID!, $documentId: ID!)
  @throwOnFieldError {
    viewer @required(action: THROW) {
      signableDocument(id: $documentId) {
        id
        title
        signed
        versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              consentText
              signed
            }
          }
        }
      }
      pendingQueue: signableDocuments(
        organizationId: $organizationId
        first: 200
        filter: { signed: false }
        orderBy: { field: UPDATED_AT, direction: DESC }
      ) {
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
  const { snapshot, goTo, close } = useDocumentQueue();
  const queueActive = useDocumentQueueActive();
  const data = usePreloadedQuery<SignatureDocumentPageQuery>(
    signatureDocumentPageQuery,
    queryRef,
  );

  const document = data.viewer.signableDocument;
  if (document == null || documentId == null) {
    throw new NotFoundError("signable document not found");
  }

  const version = document.versions.edges[0]?.node;
  if (version == null) {
    throw new NotFoundError("document version not found");
  }

  const pendingIds = useMemo(
    () => data.viewer.pendingQueue.edges.map(({ node }) => node.id),
    [data.viewer.pendingQueue.edges],
  );

  useSyncDocumentQueue({
    kind: "signatures",
    documentId,
    isPending: document.signed !== true,
    pendingIds,
  });

  const [signDocument, isSigning] = useSignDocument(document.id);
  const dataUri = useExportEmployeeDocumentPdf(version.id);

  const index = snapshot?.ids.indexOf(documentId) ?? -1;
  const nextId = index >= 0 && snapshot != null && index < snapshot.ids.length - 1
    ? snapshot.ids[index + 1]
    : null;

  return (
    <DocumentWorkspace
      title={document.title}
      dataUri={dataUri}
      request={(
        <SignatureRequestPanel
          title={document.title}
          signed={document.signed === true}
          consentText={version.consentText}
          queueActive={queueActive}
          hasNext={nextId != null}
          busy={isSigning}
          onSign={() => {
            void signDocument(version.id);
          }}
          onNext={() => {
            if (nextId != null) {
              goTo(nextId, "forward");
            }
          }}
          onFinish={() => close("signatures")}
        />
      )}
    />
  );
}
