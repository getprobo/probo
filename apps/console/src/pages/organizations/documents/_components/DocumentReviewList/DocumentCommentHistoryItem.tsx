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
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type {
  DocumentCommentHistoryItem_thread$data,
  DocumentCommentHistoryItem_thread$key,
} from "#/__generated__/core/DocumentCommentHistoryItem_thread.graphql";

import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import type { SuggestionAnchor } from "../../_lib/reviewAnchors";
import { useResolveDocumentThread } from "../../_lib/useResolveDocumentThread";
import { documentCommentHistory, documentSuggestionListItem } from "../../variants";

import { DocumentReviewAction } from "./DocumentReviewAction";
import { DocumentReviewAuthor } from "./DocumentReviewAuthor";
import { DocumentReviewItem } from "./DocumentReviewItem";
import { DocumentReviewThread } from "./DocumentReviewThread";
import { DocumentSuggestionChange } from "./DocumentSuggestionChange";

const documentCommentHistoryItemFragment = graphql`
  fragment DocumentCommentHistoryItem_thread on DocumentCommentThread {
    id
    anchorQuote
    resolvedAt
    resolvedBy {
      id
      fullName
    }
    canResolve: permission(action: "core:document:comment-resolve")
    suggestion {
      status
      createdAt
      anchorQuote
      author {
        ...DocumentReviewAuthor_profile
      }
    }
    ...DocumentReviewThread_thread
  }
`;

type HistorySuggestion = NonNullable<DocumentCommentHistoryItem_thread$data["suggestion"]>;

interface DocumentCommentHistoryItemProps {
  threadKey: DocumentCommentHistoryItem_thread$key;
  threadConnections: DocumentThreadConnections;
  // The redlines of an open suggestion, read from the text.
  liveAnchor?: SuggestionAnchor;
  active: boolean;
  onActivate: () => void;
}

// A thread in the history: the text it was left on, the comments, and
// whether it was resolved. Resolved threads take no new replies. A
// suggestion's thread shows the suggestion instead and is settled by
// accepting or rejecting it.
export function DocumentCommentHistoryItem({
  threadKey,
  threadConnections,
  liveAnchor,
  active,
  onActivate,
}: DocumentCommentHistoryItemProps) {
  const thread = useFragment(documentCommentHistoryItemFragment, threadKey);

  if (thread.suggestion != null) {
    return (
      <SuggestionHistoryItem
        thread={thread}
        suggestion={thread.suggestion}
        threadConnections={threadConnections}
        liveAnchor={liveAnchor}
        active={active}
        onActivate={onActivate}
      />
    );
  }

  return (
    <CommentHistoryItem
      thread={thread}
      threadConnections={threadConnections}
      active={active}
      onActivate={onActivate}
    />
  );
}

function SuggestionHistoryItem({
  thread,
  suggestion,
  threadConnections,
  liveAnchor,
  active,
  onActivate,
}: Omit<DocumentCommentHistoryItemProps, "threadKey"> & {
  thread: DocumentCommentHistoryItem_thread$data;
  suggestion: HistorySuggestion;
}) {
  const { t } = useTranslation("organizations/documents");
  const open = suggestion.status === "OPEN";
  const { item, quote } = documentCommentHistory({ resolved: !open });
  const { changes } = documentSuggestionListItem();
  const status = suggestionStatus(suggestion.status);
  const badge = (
    <Badge size={1} color={status.color}>
      {t(status.label)}
    </Badge>
  );

  return (
    <div className={item()}>
      <DocumentReviewItem kind="suggestion" active={active} onActivate={onActivate}>
        <DocumentReviewAuthor
          profileKey={suggestion.author}
          createdAt={suggestion.createdAt}
          primary
          trailing={badge}
        />
        {open && liveAnchor != null && (
          <div className={changes()}>
            {liveAnchor.changes.map((change, index) => (
              // The list is rebuilt from the document on every change and has
              // no stable identity beyond its order.
              <DocumentSuggestionChange key={index} change={change} />
            ))}
            {liveAnchor.changeCount > liveAnchor.changes.length && (
              <Text size={1} color="faint">
                {t("details.review.changes.moreChanges", {
                  count: liveAnchor.changeCount - liveAnchor.changes.length,
                })}
              </Text>
            )}
          </div>
        )}
        {!open && suggestion.anchorQuote != null && suggestion.anchorQuote !== "" && (
          <Text size={1} color="faint" className={quote()}>
            {suggestion.anchorQuote}
          </Text>
        )}
        <DocumentReviewThread
          threadKey={thread}
          threadConnections={threadConnections}
          expanded={active}
          allowReply={open}
        />
      </DocumentReviewItem>
    </div>
  );
}

function suggestionStatus(status: HistorySuggestion["status"]) {
  switch (status) {
    case "ACCEPTED":
      return { color: "green", label: "details.review.history.status.accepted" } as const;
    case "REJECTED":
      return { color: "red", label: "details.review.history.status.rejected" } as const;
    default:
      return { color: "sky", label: "details.review.history.status.open" } as const;
  }
}

function CommentHistoryItem({
  thread,
  threadConnections,
  active,
  onActivate,
}: Omit<DocumentCommentHistoryItemProps, "threadKey" | "liveAnchor"> & {
  thread: DocumentCommentHistoryItem_thread$data;
}) {
  const { t } = useTranslation("organizations/documents");
  const [resolveDocumentThread, isResolving] = useResolveDocumentThread();
  const resolved = thread.resolvedAt != null;
  const { item, quote } = documentCommentHistory({ resolved });

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
        {thread.anchorQuote !== "" && (
          <Text size={1} color="faint" className={quote()}>
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
