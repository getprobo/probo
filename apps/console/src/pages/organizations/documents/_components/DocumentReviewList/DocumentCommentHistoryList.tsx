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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { DocumentCommentHistoryList_document$key } from "#/__generated__/core/DocumentCommentHistoryList_document.graphql";
import type { DocumentCommentHistoryListPaginationQuery } from "#/__generated__/core/DocumentCommentHistoryListPaginationQuery.graphql";

import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import type { ThreadAnchorSource } from "../../_lib/reviewAnchors";
import { documentCommentHistory } from "../../variants";

import { DocumentCommentHistoryItem } from "./DocumentCommentHistoryItem";

const historyPageSize = 20;

// Every thread, open or resolved, including those on suggestions, newest
// first. The order must match documentCommentHistoryOrder.
const documentCommentHistoryListFragment = graphql`
  fragment DocumentCommentHistoryList_document on Document
  @refetchable(queryName: "DocumentCommentHistoryListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    commentThreads(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "DocumentCommentHistoryList_commentThreads", filters: ["orderBy"]) {
      edges {
        node {
          id
          anchorStart
          anchorEnd
          ...DocumentCommentHistoryItem_thread
        }
      }
    }
  }
`;

interface DocumentCommentHistoryListProps {
  documentKey: DocumentCommentHistoryList_document$key;
  threadConnections: DocumentThreadConnections;
  activeKey: string | null;
  onActivate: (thread: ThreadAnchorSource) => void;
  onPressInside: () => void;
}

export function DocumentCommentHistoryList({
  documentKey,
  threadConnections,
  activeKey,
  onActivate,
  onPressInside,
}: DocumentCommentHistoryListProps) {
  const { t } = useTranslation("organizations/documents");
  const { data, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    DocumentCommentHistoryListPaginationQuery,
    DocumentCommentHistoryList_document$key
  >(documentCommentHistoryListFragment, documentKey);
  const threads = data.commentThreads.edges.map(edge => edge.node);
  const { list, more } = documentCommentHistory();

  if (threads.length === 0) {
    return <Text size={1} color="faint">{t("details.review.history.empty")}</Text>;
  }

  return (
    <div className={list()}>
      {threads.map(thread => (
        <div key={thread.id} onPointerDownCapture={onPressInside}>
          <DocumentCommentHistoryItem
            threadKey={thread}
            threadConnections={threadConnections}
            active={activeKey === thread.id}
            onActivate={() => onActivate({
              id: thread.id,
              anchorStart: thread.anchorStart,
              anchorEnd: thread.anchorEnd,
            })}
          />
        </div>
      ))}
      {hasNext && (
        <Button
          size={1}
          variant="soft"
          color="neutral"
          className={more()}
          loading={isLoadingNext}
          onClick={() => loadNext(historyPageSize)}
        >
          {t("details.review.history.loadMore")}
        </Button>
      )}
    </div>
  );
}
