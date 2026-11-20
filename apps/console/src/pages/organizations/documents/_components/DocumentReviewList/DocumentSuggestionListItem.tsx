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

import { CheckIcon, XIcon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentSuggestionListItem_suggestion$key } from "#/__generated__/core/DocumentSuggestionListItem_suggestion.graphql";
import type { DocumentSuggestionListItem_thread$key } from "#/__generated__/core/DocumentSuggestionListItem_thread.graphql";

import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import type { SuggestionAnchor } from "../../_lib/reviewAnchors";
import { useAcceptDocumentSuggestion } from "../../_lib/useAcceptDocumentSuggestion";
import { useCreateDocumentCommentThread } from "../../_lib/useCreateDocumentCommentThread";
import { useRejectDocumentSuggestion } from "../../_lib/useRejectDocumentSuggestion";
import { documentSuggestionListItem } from "../../variants";

import { DocumentReviewAction } from "./DocumentReviewAction";
import { DocumentReviewAuthor } from "./DocumentReviewAuthor";
import { DocumentReviewItem } from "./DocumentReviewItem";
import { DocumentReviewReplyForm } from "./DocumentReviewReplyForm";
import { DocumentReviewThread } from "./DocumentReviewThread";
import { DocumentSuggestionChange } from "./DocumentSuggestionChange";

const documentSuggestionListItemFragment = graphql`
  fragment DocumentSuggestionListItem_suggestion on DocumentSuggestion {
    id
    createdAt
    canAccept: permission(action: "core:document:suggest-accept")
    canReject: permission(action: "core:document:suggest-reject")
    author {
      ...DocumentReviewAuthor_profile
    }
  }
`;

const documentSuggestionListItemThreadFragment = graphql`
  fragment DocumentSuggestionListItem_thread on DocumentCommentThread {
    ...DocumentReviewThread_thread
  }
`;

interface DocumentSuggestionListItemProps {
  suggestionKey: DocumentSuggestionListItem_suggestion$key;
  threadKey: DocumentSuggestionListItem_thread$key | null;
  anchor: SuggestionAnchor;
  documentId: string;
  canComment: boolean;
  threadConnections: DocumentThreadConnections;
  active: boolean;
  onActivate: () => void;
}

export function DocumentSuggestionListItem({
  suggestionKey,
  threadKey,
  anchor,
  documentId,
  canComment,
  threadConnections,
  active,
  onActivate,
}: DocumentSuggestionListItemProps) {
  const suggestion = useFragment(documentSuggestionListItemFragment, suggestionKey);
  const thread = useFragment(documentSuggestionListItemThreadFragment, threadKey);
  const { t } = useTranslation("organizations/documents");
  const [acceptDocumentSuggestion, isAccepting] = useAcceptDocumentSuggestion();
  const [rejectDocumentSuggestion, isRejecting] = useRejectDocumentSuggestion();
  const [createSuggestionThread] = useCreateDocumentCommentThread();
  const { changes } = documentSuggestionListItem();
  const hidden = anchor.changeCount - anchor.changes.length;
  const deciding = isAccepting || isRejecting;

  const decisions = suggestion.canAccept || suggestion.canReject
    ? (
        <>
          {suggestion.canAccept && (
            <DocumentReviewAction
              label={t("details.review.accept")}
              color="green"
              loading={isAccepting}
              disabled={deciding}
              onClick={() => {
                void acceptDocumentSuggestion(suggestion.id).catch(() => {
                  // The mutation already shows an error toast.
                });
              }}
            >
              <CheckIcon />
            </DocumentReviewAction>
          )}
          {suggestion.canReject && (
            <DocumentReviewAction
              label={t("details.review.reject")}
              color="red"
              loading={isRejecting}
              disabled={deciding}
              onClick={() => {
                void rejectDocumentSuggestion(suggestion.id).catch(() => {
                  // The mutation already shows an error toast.
                });
              }}
            >
              <XIcon />
            </DocumentReviewAction>
          )}
        </>
      )
    : undefined;

  return (
    <DocumentReviewItem kind="suggestion" active={active} onActivate={onActivate}>
      <DocumentReviewAuthor
        profileKey={suggestion.author}
        createdAt={suggestion.createdAt}
        primary
        trailing={decisions}
      />
      <div className={changes()}>
        {anchor.changes.map((change, index) => (
          // The list is rebuilt from the document on every change and has
          // no stable identity beyond its order.
          <DocumentSuggestionChange key={index} change={change} />
        ))}
        {hidden > 0 && (
          <Text size={1} color="faint">
            {t("details.review.changes.moreChanges", { count: hidden })}
          </Text>
        )}
      </div>
      {thread != null
        ? (
            <DocumentReviewThread
              threadKey={thread}
              threadConnections={threadConnections}
              expanded={active}
            />
          )
        : active && canComment && (
          <DocumentReviewReplyForm
            onSubmit={content => createSuggestionThread(
              documentId,
              suggestion.id,
              content,
              threadConnections,
            )}
          />
        )}
    </DocumentReviewItem>
  );
}
