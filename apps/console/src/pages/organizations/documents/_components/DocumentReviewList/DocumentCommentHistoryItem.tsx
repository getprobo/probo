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

import { CheckIcon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentCommentHistoryItem_thread$key } from "#/__generated__/core/DocumentCommentHistoryItem_thread.graphql";

import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import { useResolveDocumentThread } from "../../_lib/useResolveDocumentThread";
import { documentCommentHistory } from "../../variants";

import { DocumentReviewAction } from "./DocumentReviewAction";
import { DocumentReviewItem } from "./DocumentReviewItem";
import { DocumentReviewThread } from "./DocumentReviewThread";

const documentCommentHistoryItemFragment = graphql`
  fragment DocumentCommentHistoryItem_thread on DocumentCommentThread {
    id
    anchorQuote
    suggestionId
    resolvedAt
    resolvedBy {
      id
      fullName
    }
    canResolve: permission(action: "core:document:comment-resolve")
    ...DocumentReviewThread_thread
  }
`;

interface DocumentCommentHistoryItemProps {
  threadKey: DocumentCommentHistoryItem_thread$key;
  threadConnections: DocumentThreadConnections;
  active: boolean;
  onActivate: () => void;
}

// A thread in the history: the text it was left on, the comments, and
// whether it was resolved. Resolved threads take no new replies.
export function DocumentCommentHistoryItem({
  threadKey,
  threadConnections,
  active,
  onActivate,
}: DocumentCommentHistoryItemProps) {
  const thread = useFragment(documentCommentHistoryItemFragment, threadKey);
  const { t } = useTranslation("organizations/documents");
  const [resolveDocumentThread, isResolving] = useResolveDocumentThread();
  const resolved = thread.resolvedAt != null;
  const { item, quote, quoteLabel } = documentCommentHistory({ resolved });

  const resolve = thread.canResolve && !resolved
    ? (
        <DocumentReviewAction
          label={t("details.review.resolve")}
          color="green"
          loading={isResolving}
          onClick={() => {
            void resolveDocumentThread(thread.id, threadConnections).catch(() => {
              // The mutation already shows an error toast.
            });
          }}
        >
          <CheckIcon />
        </DocumentReviewAction>
      )
    : undefined;

  return (
    <div className={item()}>
      <DocumentReviewItem kind="comment" active={active} onActivate={onActivate}>
        {(thread.anchorQuote !== "" || thread.suggestionId != null) && (
          <Text size={1} color="faint" className={quote()}>
            {thread.suggestionId != null && (
              <span className={quoteLabel()}>{t("details.review.history.onSuggestion")}</span>
            )}
            {thread.anchorQuote}
          </Text>
        )}
        <DocumentReviewThread
          threadKey={thread}
          threadConnections={threadConnections}
          expanded={active}
          titled
          firstActions={resolve}
          allowReply={!resolved}
        />
        {resolved && (
          <Text size={1} color="faint">
            {thread.resolvedBy != null
              ? t("details.review.history.resolvedBy", { name: thread.resolvedBy.fullName })
              : t("details.review.history.resolved")}
          </Text>
        )}
      </DocumentReviewItem>
    </div>
  );
}
