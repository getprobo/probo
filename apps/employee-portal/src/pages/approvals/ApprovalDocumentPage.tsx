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

import type { ApprovalDocumentPageQuery } from "#/__generated__/core/ApprovalDocumentPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { DocumentVersionHistory } from "#/pages/_components/DocumentVersionHistory";
import { DocumentWorkspace } from "#/pages/_components/DocumentWorkspace";
import { canGoForward } from "#/pages/_lib/documentQueue";
import {
  useDocumentQueue,
  useDocumentQueueActive,
  useSyncDocumentQueue,
} from "#/pages/_lib/DocumentQueueContext";
import { useExportEmployeeDocumentPdf } from "#/pages/_lib/useExportEmployeeDocumentPdf";

import { ApprovalRequestPanel } from "./_components/ApprovalRequestPanel";
import { useApproveDocumentVersion } from "./_lib/useApproveDocumentVersion";
import { useRejectDocumentVersion } from "./_lib/useRejectDocumentVersion";

export const approvalDocumentPageQuery = graphql`
  query ApprovalDocumentPageQuery(
    $organizationId: ID!
    $documentId: ID!
    $first: Int!
    $versionsFirst: Int!
  ) @throwOnFieldError {
    viewer @required(action: THROW) {
      approvableDocument(id: $documentId) {
        id
        title
        approvalState
        latestVersion: versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              consentText
              approvalDecision {
                id
                state
              }
            }
          }
        }
      }
      ...DocumentVersionHistory_approvals @arguments(documentId: $documentId, first: $versionsFirst)
      pendingQueue: approvableDocuments(
        organizationId: $organizationId
        first: $first
        filter: { approvalStates: [PENDING] }
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

interface ApprovalDocumentPageProps {
  queryRef: PreloadedQuery<ApprovalDocumentPageQuery>;
}

export function ApprovalDocumentPage({ queryRef }: ApprovalDocumentPageProps) {
  const { documentId } = useParams();
  const { snapshot, advancing, goForward, close } = useDocumentQueue();
  const queueActive = useDocumentQueueActive();
  const data = usePreloadedQuery<ApprovalDocumentPageQuery>(
    approvalDocumentPageQuery,
    queryRef,
  );

  const document = data.viewer.approvableDocument;
  if (document == null || documentId == null) {
    throw new NotFoundError("approvable document not found");
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
  const state = document.approvalState ?? version.approvalDecision?.state ?? null;

  useSyncDocumentQueue({
    kind: "approvals",
    documentId,
    isPending: state === "PENDING" || state == null,
    pendingPage,
  });

  const [approveDocumentVersion, isApproving] = useApproveDocumentVersion(document.id);
  const [rejectDocumentVersion, isRejecting] = useRejectDocumentVersion(document.id);
  const currentVersion = selectedVersionId === version.id;
  const dataUri = useExportEmployeeDocumentPdf(selectedVersionId);

  const hasNext = snapshot != null && canGoForward(snapshot, documentId);

  return (
    <DocumentWorkspace
      title={document.title}
      dataUri={dataUri}
      history={(
        <DocumentVersionHistory
          viewerKey={data.viewer}
          kind="approvals"
          selectedVersionId={selectedVersionId}
          onSelect={(versionId) => {
            setSelected({ documentId: document.id, versionId });
          }}
        />
      )}
      request={(
        <ApprovalRequestPanel
          title={document.title}
          state={state}
          consentText={version.consentText}
          queueActive={queueActive}
          hasNext={hasNext}
          isApproving={isApproving}
          isRejecting={isRejecting}
          advancing={advancing}
          isCurrentVersion={currentVersion}
          onApprove={() => {
            if (!currentVersion) {
              return;
            }
            void approveDocumentVersion(version.id);
          }}
          onReject={() => {
            if (!currentVersion) {
              return;
            }
            void rejectDocumentVersion(version.id);
          }}
          onNext={goForward}
          onFinish={() => close("approvals")}
        />
      )}
    />
  );
}
