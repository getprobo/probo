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
import type { UIEvent } from "react";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { DocumentVersionHistory_approvals$key } from "#/__generated__/core/DocumentVersionHistory_approvals.graphql";
import type { DocumentVersionHistory_signatures$key } from "#/__generated__/core/DocumentVersionHistory_signatures.graphql";
import type { DocumentVersionHistoryApprovalsPaginationQuery } from "#/__generated__/core/DocumentVersionHistoryApprovalsPaginationQuery.graphql";
import type { DocumentVersionHistoryItem_version$key } from "#/__generated__/core/DocumentVersionHistoryItem_version.graphql";
import type { DocumentVersionHistorySignaturesPaginationQuery } from "#/__generated__/core/DocumentVersionHistorySignaturesPaginationQuery.graphql";
import {
  DOCUMENT_VERSION_PAGE_SIZE,
  DOCUMENT_VERSION_PEEK_PX,
  DOCUMENT_VERSION_ROW_HEIGHT_PX,
  DOCUMENT_VERSION_VISIBLE_COUNT,
  type DocumentVersionHistoryKind,
} from "#/pages/_lib/documentVersion";

import { DocumentVersionHistoryItem } from "./DocumentVersionHistoryItem";
import { documentVersionHistory } from "./variants";

const documentVersionHistorySignaturesFragment = graphql`
  fragment DocumentVersionHistory_signatures on Viewer
  @argumentDefinitions(
    documentId: { type: "ID!" }
    first: { type: "Int", defaultValue: 25 }
    after: { type: "CursorKey" }
  )
  @refetchable(queryName: "DocumentVersionHistorySignaturesPaginationQuery")
  @throwOnFieldError {
    historyDocument: signableDocument(id: $documentId) {
      versions(
        first: $first
        after: $after
        orderBy: { field: CREATED_AT, direction: DESC }
      ) @connection(key: "DocumentVersionHistory_signatures_versions") {
        totalCount
        edges {
          node {
            id
            ...DocumentVersionHistoryItem_version
          }
        }
      }
    }
  }
`;

const documentVersionHistoryApprovalsFragment = graphql`
  fragment DocumentVersionHistory_approvals on Viewer
  @argumentDefinitions(
    documentId: { type: "ID!" }
    first: { type: "Int", defaultValue: 25 }
    after: { type: "CursorKey" }
  )
  @refetchable(queryName: "DocumentVersionHistoryApprovalsPaginationQuery")
  @throwOnFieldError {
    historyDocument: approvableDocument(id: $documentId) {
      versions(
        first: $first
        after: $after
        orderBy: { field: CREATED_AT, direction: DESC }
      ) @connection(key: "DocumentVersionHistory_approvals_versions") {
        totalCount
        edges {
          node {
            id
            ...DocumentVersionHistoryItem_version
          }
        }
      }
    }
  }
`;

type VersionRow = {
  id: string;
} & DocumentVersionHistoryItem_version$key;

type VersionConnection = {
  totalCount: number;
  edges: ReadonlyArray<{
    node: VersionRow;
  }>;
};

export type DocumentVersionHistoryProps = {
  selectedVersionId: string;
  onSelect: (versionId: string) => void;
} & (
  | {
    kind: "signatures";
    viewerKey: DocumentVersionHistory_signatures$key;
  }
  | {
    kind: "approvals";
    viewerKey: DocumentVersionHistory_approvals$key;
  }
);

export function DocumentVersionHistory({
  kind,
  viewerKey,
  selectedVersionId,
  onSelect,
}: DocumentVersionHistoryProps) {
  if (kind === "signatures") {
    return (
      <SignaturesVersionHistory
        viewerKey={viewerKey}
        selectedVersionId={selectedVersionId}
        onSelect={onSelect}
      />
    );
  }

  return (
    <ApprovalsVersionHistory
      viewerKey={viewerKey}
      selectedVersionId={selectedVersionId}
      onSelect={onSelect}
    />
  );
}

function SignaturesVersionHistory({
  viewerKey,
  selectedVersionId,
  onSelect,
}: {
  viewerKey: DocumentVersionHistory_signatures$key;
  selectedVersionId: string;
  onSelect: (versionId: string) => void;
}) {
  const { data, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    DocumentVersionHistorySignaturesPaginationQuery,
    DocumentVersionHistory_signatures$key
  >(documentVersionHistorySignaturesFragment, viewerKey);

  return (
    <VersionHistoryList
      kind="signatures"
      versions={data.historyDocument?.versions}
      hasNext={hasNext}
      isLoadingNext={isLoadingNext}
      selectedVersionId={selectedVersionId}
      onSelect={onSelect}
      onLoadNext={() => {
        loadNext(DOCUMENT_VERSION_PAGE_SIZE);
      }}
    />
  );
}

function ApprovalsVersionHistory({
  viewerKey,
  selectedVersionId,
  onSelect,
}: {
  viewerKey: DocumentVersionHistory_approvals$key;
  selectedVersionId: string;
  onSelect: (versionId: string) => void;
}) {
  const { data, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    DocumentVersionHistoryApprovalsPaginationQuery,
    DocumentVersionHistory_approvals$key
  >(documentVersionHistoryApprovalsFragment, viewerKey);

  return (
    <VersionHistoryList
      kind="approvals"
      versions={data.historyDocument?.versions}
      hasNext={hasNext}
      isLoadingNext={isLoadingNext}
      selectedVersionId={selectedVersionId}
      onSelect={onSelect}
      onLoadNext={() => {
        loadNext(DOCUMENT_VERSION_PAGE_SIZE);
      }}
    />
  );
}

function VersionHistoryList({
  kind,
  versions,
  hasNext,
  isLoadingNext,
  selectedVersionId,
  onSelect,
  onLoadNext,
}: {
  kind: DocumentVersionHistoryKind;
  versions: VersionConnection | null | undefined;
  hasNext: boolean;
  isLoadingNext: boolean;
  selectedVersionId: string;
  onSelect: (versionId: string) => void;
  onLoadNext: () => void;
}) {
  const { t } = useTranslation();
  const count = versions?.totalCount ?? 0;
  const edges = versions?.edges ?? [];
  const peek = count > DOCUMENT_VERSION_VISIBLE_COUNT;
  const slots = documentVersionHistory({ peek });
  const showGhost = peek && (hasNext || edges.length < count);
  const viewportHeight = peek
    ? DOCUMENT_VERSION_VISIBLE_COUNT * DOCUMENT_VERSION_ROW_HEIGHT_PX
    + DOCUMENT_VERSION_PEEK_PX
    : edges.length * DOCUMENT_VERSION_ROW_HEIGHT_PX;

  const onScroll = (event: UIEvent<HTMLDivElement>) => {
    if (!hasNext || isLoadingNext) {
      return;
    }
    const { scrollTop, clientHeight, scrollHeight } = event.currentTarget;
    if (scrollTop + clientHeight >= scrollHeight - DOCUMENT_VERSION_ROW_HEIGHT_PX) {
      onLoadNext();
    }
  };

  return (
    <div className={slots.root()}>
      <div className={slots.header()}>
        <Text size={2} weight="medium" highContrast>
          {t("documents.versions.heading")}
        </Text>
        <Text size={2} color="neutral">
          {t("documents.versions.count", { count })}
        </Text>
      </div>
      <div className={slots.frame()}>
        <div
          className={slots.viewport()}
          style={{ height: viewportHeight }}
          onScroll={onScroll}
        >
          <div className={slots.list()}>
            {edges.map(({ node }, index) => (
              <DocumentVersionHistoryItem
                key={node.id}
                versionKey={node}
                kind={kind}
                selected={node.id === selectedVersionId}
                current={index === 0}
                onSelect={() => {
                  onSelect(node.id);
                }}
              />
            ))}
            {showGhost ? <div className={slots.ghost()} aria-hidden /> : null}
          </div>
        </div>
      </div>
    </div>
  );
}
