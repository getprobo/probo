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

import { CommentEditor } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, useContext, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentReviewComment_comment$key } from "#/__generated__/core/DocumentReviewComment_comment.graphql";
import {
  isRichEditorContentEmpty,
  richEditorContentTextLength,
} from "#/pages/organizations/_lib/richEditorContent";
import { useMentionMembers } from "#/pages/organizations/_lib/useMentionMembers";
import { CurrentUser } from "#/providers/CurrentUser";

import { documentCommentMaxLength } from "../../_lib/documentCommentMaxLength";
import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import { useUpdateDocumentComment } from "../../_lib/useUpdateDocumentComment";
import { documentReviewComment } from "../../variants";

import { DocumentReviewAuthor } from "./DocumentReviewAuthor";
import { DocumentReviewCommentDeleteDialog } from "./DocumentReviewCommentDeleteDialog";
import { DocumentReviewCommentMenu } from "./DocumentReviewCommentMenu";

const documentReviewCommentFragment = graphql`
  fragment DocumentReviewComment_comment on DocumentComment {
    id
    content
    createdAt
    canUpdate: permission(action: "core:document:comment-update")
    canDelete: permission(action: "core:document:comment-delete")
    author {
      id
      ...DocumentReviewAuthor_profile
    }
    ...DocumentReviewCommentDeleteDialog_comment
  }
`;

interface DocumentReviewCommentProps {
  commentKey: DocumentReviewComment_comment$key;
  commentsConnectionId: string;
  threadConnections: DocumentThreadConnections;
  primary?: boolean;
  actions?: ReactNode;
}

export function DocumentReviewComment({
  commentKey,
  commentsConnectionId,
  threadConnections,
  primary = false,
  actions,
}: DocumentReviewCommentProps) {
  const comment = useFragment(documentReviewCommentFragment, commentKey);
  const { t } = useTranslation("organizations/documents");
  const profileId = useContext(CurrentUser).profile?.id;
  const mentionSource = useMentionMembers();
  const [updateDocumentComment, isSaving] = useUpdateDocumentComment();
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(comment.content);
  const [error, setError] = useState<string>();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { root, content, editor, save } = documentReviewComment();
  // The server only lets the author change a comment.
  const own = profileId != null && comment.author.id === profileId;
  const canEdit = own && comment.canUpdate;
  const canDelete = own && comment.canDelete;

  function startEditing() {
    setDraft(comment.content);
    setError(undefined);
    setEditing(true);
  }

  async function saveComment() {
    if (isSaving) {
      return;
    }
    if (isRichEditorContentEmpty(draft)) {
      setError(t("details.review.contentRequired"));
      return;
    }
    if (richEditorContentTextLength(draft) > documentCommentMaxLength) {
      setError(t("details.review.contentTooLong", { max: documentCommentMaxLength }));
      return;
    }
    try {
      await updateDocumentComment(comment.id, draft);
      setEditing(false);
    } catch {
      // The mutation already shows an error toast.
    }
  }

  const menu = (
    <DocumentReviewCommentMenu
      canEdit={canEdit && !editing}
      canDelete={canDelete}
      onEdit={startEditing}
      onDelete={() => setDeleteOpen(true)}
    />
  );

  return (
    <div className={root()}>
      <DocumentReviewAuthor
        profileKey={comment.author}
        createdAt={comment.createdAt}
        primary={primary}
        trailing={actions != null || canEdit || canDelete
          ? (
              <>
                {actions}
                {menu}
              </>
            )
          : undefined}
      />
      {editing
        ? (
            <>
              <Field error={error}>
                <CommentEditor
                  key="edit"
                  autoFocus
                  className={editor()}
                  content={draft}
                  controls={false}
                  disabled={isSaving}
                  mentionSource={mentionSource}
                  aria-label={t("details.review.edit")}
                  onModEnter={() => {
                    void saveComment();
                  }}
                  onChangeContent={(next) => {
                    setDraft(next);
                    setError(undefined);
                  }}
                />
              </Field>
              <div className={save()}>
                <Button
                  size={1}
                  variant="ghost"
                  color="neutral"
                  disabled={isSaving}
                  onClick={() => {
                    setEditing(false);
                    setError(undefined);
                  }}
                >
                  {t("details.review.cancel")}
                </Button>
                <Button
                  size={1}
                  variant="solid"
                  color="neutral"
                  highContrast
                  loading={isSaving}
                  onClick={() => {
                    void saveComment();
                  }}
                >
                  {t("details.review.save")}
                </Button>
              </div>
            </>
          )
        : (
            <ErrorBoundary
              fallback={(
                <Text size={2} color="faint">
                  {t("details.review.contentError")}
                </Text>
              )}
            >
              <CommentEditor
                key={comment.content}
                className={content()}
                content={comment.content}
                controls={false}
                disabled
              />
            </ErrorBoundary>
          )}
      {canDelete && (
        <DocumentReviewCommentDeleteDialog
          commentKey={comment}
          commentsConnectionId={commentsConnectionId}
          threadConnections={threadConnections}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
        />
      )}
    </div>
  );
}
