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
import type { MentionSource } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import type { TFunction } from "i18next";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentReviewMarginMutations_acceptMutation } from "#/__generated__/core/DocumentReviewMarginMutations_acceptMutation.graphql";
import type { DocumentReviewMarginMutations_createThreadMutation } from "#/__generated__/core/DocumentReviewMarginMutations_createThreadMutation.graphql";
import type { DocumentReviewMarginMutations_deleteCommentMutation } from "#/__generated__/core/DocumentReviewMarginMutations_deleteCommentMutation.graphql";
import type { DocumentReviewMarginMutations_rejectMutation } from "#/__generated__/core/DocumentReviewMarginMutations_rejectMutation.graphql";
import type { DocumentReviewMarginMutations_replyMutation } from "#/__generated__/core/DocumentReviewMarginMutations_replyMutation.graphql";
import type { DocumentReviewMarginMutations_updateCommentMutation } from "#/__generated__/core/DocumentReviewMarginMutations_updateCommentMutation.graphql";
import type { DocumentReviewMarginSuggestion_suggestion$key } from "#/__generated__/core/DocumentReviewMarginSuggestion_suggestion.graphql";
import type { DocumentReviewMarginThread_thread$key } from "#/__generated__/core/DocumentReviewMarginThread_thread.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import {
  closeSuggestion,
  deleteCommentEdge,
  insertCommentEdge,
  insertThreadEdge,
} from "../../_lib/reviewMarginStore";
import type { SuggestionPreview } from "../../_lib/suggestionPreview";

import { DocumentReviewMarginCard } from "./DocumentReviewMarginCard";
import { DocumentReviewMarginComments } from "./DocumentReviewMarginComments";
import {
  acceptMutation,
  createThreadMutation,
  deleteCommentMutation,
  rejectMutation,
  replyMutation,
  updateCommentMutation,
} from "./DocumentReviewMarginMutations";
import { documentReviewMarginThreadFragment } from "./DocumentReviewMarginThread";

const mutationFeedback = { errorToast: false };

export const documentReviewMarginSuggestionFragment = graphql`
  fragment DocumentReviewMarginSuggestion_suggestion on DocumentSuggestion {
    id
    status
    createdAt
    author { id fullName }
  }
`;

type LocalSuggestion = {
  id: string;
  author: { fullName: string | null | undefined };
};

interface DocumentReviewMarginSuggestionProps {
  documentId: string;
  suggestionKey?: DocumentReviewMarginSuggestion_suggestion$key | null;
  localSuggestion?: LocalSuggestion | null;
  threadKey?: DocumentReviewMarginThread_thread$key | null;
  preview: Pick<SuggestionPreview, "parts" | "structure" | "formats">;
  active: boolean;
  canUpdate: boolean;
  canComment: boolean;
  profileId: string | null;
  mentionSource: MentionSource;
  onActivate: () => void;
  onDecided: (suggestionId: string) => void;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
  onLayout?: () => void;
}

export function DocumentReviewMarginSuggestion({
  documentId,
  suggestionKey,
  localSuggestion,
  threadKey,
  preview,
  active,
  canUpdate,
  canComment,
  profileId,
  mentionSource,
  onActivate,
  onDecided,
  onOwnReview,
  onOwnReviewFailed,
  onLayout,
}: DocumentReviewMarginSuggestionProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const loaded = useFragment(documentReviewMarginSuggestionFragment, suggestionKey ?? null);
  const thread = useFragment(documentReviewMarginThreadFragment, threadKey ?? null);
  const suggestion = loaded ?? localSuggestion;
  const [replyText, setReplyText] = useState("");
  const [createThread] = useMutation<DocumentReviewMarginMutations_createThreadMutation>(
    createThreadMutation,
    mutationFeedback,
  );
  const [reply, replying] = useMutation<DocumentReviewMarginMutations_replyMutation>(
    replyMutation,
    mutationFeedback,
  );
  const [updateComment] = useMutation<DocumentReviewMarginMutations_updateCommentMutation>(
    updateCommentMutation,
    mutationFeedback,
  );
  const [deleteComment] = useMutation<DocumentReviewMarginMutations_deleteCommentMutation>(
    deleteCommentMutation,
    mutationFeedback,
  );
  const [acceptSuggestion] = useMutation<DocumentReviewMarginMutations_acceptMutation>(
    acceptMutation,
    mutationFeedback,
  );
  const [rejectSuggestion] = useMutation<DocumentReviewMarginMutations_rejectMutation>(
    rejectMutation,
    mutationFeedback,
  );

  if (!suggestion) return null;

  const reportError = (error: Error) => {
    toast.add({
      title: t("documentDescriptionPage.errors.title"),
      description: error.message,
      type: "error",
    });
  };

  return (
    <DocumentReviewMarginCard
      active={active}
      label={t("documentDescriptionPage.review.selectSuggestion", {
        name: suggestion.author.fullName ?? "",
      })}
      onActivate={onActivate}
    >
      <div className="flex items-start justify-between gap-2">
        <p className="text-2 font-medium text-sand-12">{suggestion.author.fullName}</p>
        {canUpdate && !suggestion.id.startsWith("local:") && (
          <div className="flex gap-1">
            <Button
              variant="ghost"
              color="green"
              size={1}
              aria-label={t("documentDescriptionPage.review.accept")}
              iconStart={<CheckIcon />}
              onClick={() => {
                onOwnReview();
                void acceptSuggestion({
                  variables: { input: { suggestionId: suggestion.id } },
                  updater: store => closeSuggestion(store, documentId, suggestion.id),
                  onCompleted: (_response, errors) => {
                    if (errors?.length) {
                      onOwnReviewFailed();
                      return;
                    }
                    onDecided(suggestion.id);
                  },
                  onError: (error) => {
                    onOwnReviewFailed();
                    reportError(error);
                  },
                });
              }}
            />
            <Button
              variant="ghost"
              color="red"
              size={1}
              aria-label={t("documentDescriptionPage.review.reject")}
              iconStart={<XIcon />}
              onClick={() => {
                onOwnReview();
                void rejectSuggestion({
                  variables: { input: { suggestionId: suggestion.id } },
                  updater: store => closeSuggestion(store, documentId, suggestion.id),
                  onCompleted: (_response, errors) => {
                    if (errors?.length) {
                      onOwnReviewFailed();
                      return;
                    }
                    onDecided(suggestion.id);
                  },
                  onError: (error) => {
                    onOwnReviewFailed();
                    reportError(error);
                  },
                });
              }}
            />
          </div>
        )}
      </div>
      <SuggestionPreviewBody
        preview={preview}
        quote={thread?.anchorQuote}
      />
      <DocumentReviewMarginComments
        thread={thread}
        active={active}
        profileId={profileId}
        canComment={canComment && !suggestion.id.startsWith("local:")}
        replyText={active ? replyText : ""}
        mentionSource={mentionSource}
        onReplyText={setReplyText}
        replying={replying}
        onLayout={onLayout}
        onReply={(content) => {
          if (thread) {
            onOwnReview();
            void reply({
              variables: { input: { threadId: thread.id, content } },
              updater: store => insertCommentEdge(store, thread.id),
              onCompleted: (_response, errors) => {
                if (errors?.length) {
                  onOwnReviewFailed();
                  return;
                }
                setReplyText("");
              },
              onError: (error) => {
                onOwnReviewFailed();
                reportError(error);
              },
            });
            return;
          }
          onOwnReview();
          void createThread({
            variables: { input: { documentId: documentId, suggestionId: suggestion.id, content } },
            updater: store => insertThreadEdge(store, documentId),
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                onOwnReviewFailed();
                return;
              }
              setReplyText("");
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
        onUpdate={(commentId, content) => {
          onOwnReview();
          void updateComment({
            variables: { input: { commentId, content } },
            onCompleted: (_response, errors) => {
              if (errors?.length) onOwnReviewFailed();
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
        onDelete={(commentId) => {
          if (!thread) return;
          onOwnReview();
          void deleteComment({
            variables: { input: { commentId } },
            updater: store => deleteCommentEdge(store, thread.id, commentId),
            onCompleted: (_response, errors) => {
              if (errors?.length) onOwnReviewFailed();
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
      />
    </DocumentReviewMarginCard>
  );
}

function SuggestionPreviewBody(props: {
  preview: Pick<SuggestionPreview, "parts" | "structure" | "formats">;
  quote?: string | null;
}) {
  const { t } = useTranslation();
  const { preview, quote } = props;
  const adds = !preview.structure && preview.formats && preview.formats.adds.length > 0
    ? preview.formats.adds
    : null;
  const removes = !preview.structure && preview.formats && preview.formats.removes.length > 0
    ? preview.formats.removes
    : null;
  const fallback = preview.parts.length === 0 && !preview.formats ? quote : null;

  return (
    <div className="flex flex-col gap-1 text-2 text-sand-12">
      {preview.structure && (
        <p>{structureLabel(t, preview.structure)}</p>
      )}
      {adds && (
        <p className="text-green-11">
          {t("documentDescriptionPage.review.format", { styles: styleList(t, adds) })}
        </p>
      )}
      {removes && (
        <p className="text-red-11">
          {t("documentDescriptionPage.review.formatRemove", { styles: styleList(t, removes) })}
        </p>
      )}
      {!preview.structure && (
        <SuggestionPreviewText parts={preview.parts} quote={fallback} />
      )}
    </div>
  );
}

function SuggestionPreviewText(props: { parts: SuggestionPreview["parts"]; quote?: string | null }) {
  const { t } = useTranslation();
  const { parts, quote } = props;
  const text = parts.map(part => part.text).join("").replace(/\s+/g, " ").trim();
  if (parts.length > 0 && parts.every(part => part.kind === "inserted") && text !== "") {
    return <p>{t("documentDescriptionPage.review.addText", { text })}</p>;
  }
  if (parts.length > 0 && parts.every(part => part.kind === "deleted") && text !== "") {
    return <p>{t("documentDescriptionPage.review.deleteText", { text })}</p>;
  }
  if (parts.length === 0) {
    return quote ? <p className="whitespace-pre-wrap">{quote}</p> : null;
  }
  return (
    <p className="whitespace-pre-wrap">
      {parts.map((part, index) =>
        part.kind === "deleted"
          ? (
              <span key={index} className="text-red-11 line-through">{part.text}</span>
            )
          : part.kind === "inserted"
            ? (
                <span key={index} className="text-green-11">{part.text}</span>
              )
            : (
                <span key={index}>{part.text}</span>
              ),
      )}
    </p>
  );
}

function structureLabel(t: TFunction, structure: NonNullable<SuggestionPreview["structure"]>): string {
  if (structure.action === "format") return formatStructureLabel(t, structure);
  const text = structure.text;
  if (structure.target === "divider") {
    return structure.action === "add"
      ? t("documentDescriptionPage.review.addDivider")
      : t("documentDescriptionPage.review.deleteDivider");
  }
  if (structure.target === "image") {
    return structure.action === "add"
      ? t("documentDescriptionPage.review.addImage")
      : t("documentDescriptionPage.review.deleteImage");
  }
  if (structure.target === "file") {
    if (structure.action === "add") {
      return text === ""
        ? t("documentDescriptionPage.review.addFile")
        : t("documentDescriptionPage.review.addFileNamed", { name: text });
    }
    return text === ""
      ? t("documentDescriptionPage.review.deleteFile")
      : t("documentDescriptionPage.review.deleteFileNamed", { name: text });
  }
  if (structure.target === "mention") {
    if (structure.action === "add") {
      return text === ""
        ? t("documentDescriptionPage.review.addMention")
        : t("documentDescriptionPage.review.addMentionNamed", { name: text });
    }
    return text === ""
      ? t("documentDescriptionPage.review.deleteMention")
      : t("documentDescriptionPage.review.deleteMentionNamed", { name: text });
  }
  if (structure.action === "add" && structure.target === "row") {
    return text === "" ? t("documentDescriptionPage.review.addRow") : t("documentDescriptionPage.review.addRowWithText", { text });
  }
  if (structure.action === "add" && structure.target === "column") {
    return text === "" ? t("documentDescriptionPage.review.addColumn") : t("documentDescriptionPage.review.addColumnWithText", { text });
  }
  if (structure.action === "add") {
    return text === "" ? t("documentDescriptionPage.review.addTable") : t("documentDescriptionPage.review.addTableWithText", { text });
  }
  if (structure.target === "row") {
    return text === "" ? t("documentDescriptionPage.review.deleteRow") : t("documentDescriptionPage.review.deleteRowWithText", { text });
  }
  if (structure.target === "column") {
    return text === "" ? t("documentDescriptionPage.review.deleteColumn") : t("documentDescriptionPage.review.deleteColumnWithText", { text });
  }
  return text === "" ? t("documentDescriptionPage.review.deleteTable") : t("documentDescriptionPage.review.deleteTableWithText", { text });
}

function formatStructureLabel(t: TFunction, structure: NonNullable<SuggestionPreview["structure"]>): string {
  if (structure.target === "heading") {
    return t("documentDescriptionPage.review.formatHeading", { level: structure.text });
  }
  if (structure.target === "paragraph") return t("documentDescriptionPage.review.formatParagraph");
  if (structure.target === "quote") return t("documentDescriptionPage.review.formatQuote");
  if (structure.target === "code") return t("documentDescriptionPage.review.formatCode");
  if (structure.target === "task") {
    return structure.text === "uncheck"
      ? t("documentDescriptionPage.review.uncheckTask")
      : t("documentDescriptionPage.review.checkTask");
  }
  if (structure.text === "remove") return t("documentDescriptionPage.review.listRemove");
  if (structure.text === "nest") return t("documentDescriptionPage.review.listNest");
  if (structure.text === "unnest") return t("documentDescriptionPage.review.listUnnest");
  return t("documentDescriptionPage.review.listAdd");
}

function styleList(t: TFunction, styles: string[]): string {
  return styles.map((style) => {
    switch (style) {
      case "bold":
        return t("documentDescriptionPage.review.styleBold");
      case "italic":
        return t("documentDescriptionPage.review.styleItalic");
      case "underline":
        return t("documentDescriptionPage.review.styleUnderline");
      case "strikethrough":
        return t("documentDescriptionPage.review.styleStrike");
      case "code":
        return t("documentDescriptionPage.review.styleCode");
      case "link":
        return t("documentDescriptionPage.review.styleLink");
      default:
        return style;
    }
  }).join(", ");
}
