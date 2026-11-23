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

import { ChatCircleIcon } from "@phosphor-icons/react";
import type { RepoCollaborationHandle } from "@probo/collaboration";
import { BubbleMenuToolbar } from "@probo/ui/src/RichEditor/BubbleMenu";
import { MenuButton } from "@probo/ui/src/RichEditor/MenuButton";
import type { Editor } from "@tiptap/core";
import { useTranslation } from "react-i18next";

import { commentDraftFromSelection, type DocumentCommentDraft } from "../../_lib/documentCommentDraft";

interface DocumentCommentBubbleItemProps {
  editor: Editor;
  collaborationHandle: RepoCollaborationHandle;
  onComment: (draft: DocumentCommentDraft) => void;
}

export function DocumentCommentBubbleItem({
  editor,
  collaborationHandle,
  onComment,
}: DocumentCommentBubbleItemProps) {
  const { t } = useTranslation("organizations/documents");

  function startComment() {
    const draft = commentDraftFromSelection(editor, collaborationHandle);
    if (draft == null) {
      return;
    }
    // The draft highlight marks the text from here. A selection left in
    // place would bring the bubble back once the composer closes.
    editor.commands.setTextSelection(editor.state.selection.to);
    onComment(draft);
  }

  return (
    <BubbleMenuToolbar>
      <MenuButton onClick={startComment}>
        <ChatCircleIcon size={16} weight="bold" />
        {t("details.review.comment")}
      </MenuButton>
    </BubbleMenuToolbar>
  );
}
