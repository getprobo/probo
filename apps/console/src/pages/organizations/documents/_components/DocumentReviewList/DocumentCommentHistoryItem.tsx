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
import { documentCommentHistory } from "../../variants";

import { DocumentReviewAction } from "./DocumentReviewAction";
import { DocumentReviewAuthor } from "./DocumentReviewAuthor";
import { DocumentReviewItem } from "./DocumentReviewItem";
import { DocumentReviewThread } from "./DocumentReviewThread";
import { DocumentSuggestionChangeList } from "./DocumentSuggestionChangeList";
import { DocumentSuggestionChangeSummary } from "./DocumentSuggestionChangeSummary";
import { DocumentSuggestionDecisions } from "./DocumentSuggestionDecisions";

const documentCommentHistoryItemFragment = graphql`
  fragment DocumentCommentHistoryItem_thread on DocumentCommentThread {
    id
    resolvedAt
    resolvedBy {
      id
      fullName
    }
    canResolve: permission(action: "core:document:comment-resolve")
    suggestion {
      status
      createdAt
      author {
        ...DocumentReviewAuthor_profile
      }
      ...DocumentSuggestionDecisions_suggestion
      ...DocumentSuggestionChangeList_suggestion
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
  const { item } = documentCommentHistory({ resolved: !open });
  const status = suggestionStatus(suggestion.status);
  const badge = (
    <Badge size={1} color={status.color}>
      {t(status.label)}
    </Badge>
  );
  const trailing = open
    ? (
        <>
          {badge}
          <DocumentSuggestionDecisions suggestionKey={suggestion} threadConnections={threadConnections} />
        </>
      )
    : badge;

  return (
    <div className={item()}>
      <DocumentReviewItem kind="suggestion" active={active} onActivate={onActivate}>
        <DocumentReviewAuthor
          profileKey={suggestion.author}
          createdAt={suggestion.createdAt}
          primary
          trailing={trailing}
        />
        {open && liveAnchor != null && (
          <DocumentSuggestionChangeSummary changes={liveAnchor.changes} changeCount={liveAnchor.changeCount} />
        )}
        {!open && <DocumentSuggestionChangeList suggestionKey={suggestion} />}
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
  const { item } = documentCommentHistory({ resolved });

  const actions = resolved
    ? (
        <Badge size={1} color="green">
          {t("details.review.history.resolved")}
        </Badge>
      )
    : thread.canResolve
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
        <DocumentReviewThread
          threadKey={thread}
          threadConnections={threadConnections}
          expanded={active}
          titled
          firstActions={actions}
          allowReply={!resolved}
        />
        {resolved && thread.resolvedBy != null && (
          <Text size={1} color="faint">
            {t("details.review.history.resolvedBy", { name: thread.resolvedBy.fullName })}
          </Text>
        )}
      </DocumentReviewItem>
    </div>
  );
}
