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

import { useContext } from "react";
import { useTranslation } from "react-i18next";

import { CurrentUser } from "#/providers/CurrentUser";

import type { DocumentCommentDraft } from "../../_lib/documentCommentDraft";
import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import { useCreateDocumentCommentThread } from "../../_lib/useCreateDocumentCommentThread";

import { DocumentReviewAuthorView } from "./DocumentReviewAuthor";
import { DocumentReviewItem } from "./DocumentReviewItem";
import { DocumentReviewReplyForm } from "./DocumentReviewReplyForm";

interface DocumentCommentDraftCardProps {
  documentId: string;
  draft: DocumentCommentDraft;
  threadConnections: DocumentThreadConnections;
  // Receives the new thread id, or null when the draft is discarded.
  onDone: (threadId: string | null) => void;
}

export function DocumentCommentDraftCard({
  documentId,
  draft,
  threadConnections,
  onDone,
}: DocumentCommentDraftCardProps) {
  const { t } = useTranslation("organizations/documents");
  const profile = useContext(CurrentUser).profile;
  const [createDocumentCommentThread] = useCreateDocumentCommentThread();

  async function submit(content: string) {
    const threadId = await createDocumentCommentThread(documentId, draft, content, threadConnections);
    onDone(threadId);
  }

  return (
    <DocumentReviewItem kind="comment" active onActivate={() => {}}>
      {profile != null && (
        <DocumentReviewAuthorView
          fullName={profile.fullName}
          emailAddress={profile.emailAddress}
          avatarUrl={profile.avatarUrl}
        />
      )}
      <DocumentReviewReplyForm
        autoFocus
        placeholder={t("details.review.commentPlaceholder")}
        label={t("details.review.comment")}
        onSubmit={submit}
        onCancel={() => onDone(null)}
      />
    </DocumentReviewItem>
  );
}
