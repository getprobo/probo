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
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { DocumentReviewThread_thread$key } from "#/__generated__/core/DocumentReviewThread_thread.graphql";
import type { DocumentReviewThreadCommentsPaginationQuery } from "#/__generated__/core/DocumentReviewThreadCommentsPaginationQuery.graphql";

import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import { useCreateDocumentComment } from "../../_lib/useCreateDocumentComment";
import { documentReviewThread } from "../../variants";

import { DocumentReviewComment } from "./DocumentReviewComment";
import { DocumentReviewReplyForm } from "./DocumentReviewReplyForm";

const threadPageSize = 20;

const documentReviewThreadFragment = graphql`
  fragment DocumentReviewThread_thread on DocumentCommentThread
  @refetchable(queryName: "DocumentReviewThreadCommentsPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    id
    canReply: permission(action: "core:document:comment")
    comments(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: ASC }
    ) @connection(key: "DocumentReviewThread_comments", filters: ["orderBy"]) {
      __id
      totalCount
      edges {
        node {
          id
          ...DocumentReviewComment_comment
        }
      }
    }
  }
`;

interface DocumentReviewThreadProps {
  threadKey: DocumentReviewThread_thread$key;
  threadConnections: DocumentThreadConnections;
  expanded: boolean;
  // The first comment is the card's title. Off when the thread sits under a
  // suggestion, whose author row is the title instead.
  titled?: boolean;
  firstActions?: ReactNode;
  // Off for a resolved thread, which takes no new replies.
  allowReply?: boolean;
}

export function DocumentReviewThread({
  threadKey,
  threadConnections,
  expanded,
  titled = false,
  firstActions,
  allowReply = true,
}: DocumentReviewThreadProps) {
  const { t } = useTranslation("organizations/documents");
  const { data: thread, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    DocumentReviewThreadCommentsPaginationQuery,
    DocumentReviewThread_thread$key
  >(documentReviewThreadFragment, threadKey);
  const [createDocumentComment] = useCreateDocumentComment();
  const { root, more } = documentReviewThread();
  const commentsConnectionId = thread.comments.__id;
  const comments = thread.comments.edges.map(edge => edge.node);
  const visible = expanded ? comments : comments.slice(0, 1);
  // totalCount is not updated by a reply or a delete, so the loaded list is the
  // count once every page is in the store.
  const total = hasNext ? thread.comments.totalCount : comments.length;
  const replies = total - 1;

  return (
    <div className={root()}>
      {visible.map((comment, index) => (
        <DocumentReviewComment
          key={comment.id}
          commentKey={comment}
          commentsConnectionId={commentsConnectionId}
          threadConnections={threadConnections}
          primary={titled && index === 0}
          actions={index === 0 ? firstActions : undefined}
        />
      ))}
      {!expanded && replies > 0 && (
        <Text size={1} color="faint">
          {t("details.review.replies", { count: replies })}
        </Text>
      )}
      {expanded && hasNext && (
        <Button
          size={1}
          variant="ghost"
          color="neutral"
          className={more()}
          disabled={isLoadingNext}
          onClick={() => loadNext(threadPageSize)}
        >
          {t("details.review.showMoreReplies")}
        </Button>
      )}
      {expanded && allowReply && thread.canReply && (
        <DocumentReviewReplyForm
          onSubmit={content => createDocumentComment(thread.id, content, commentsConnectionId)}
        />
      )}
    </div>
  );
}
