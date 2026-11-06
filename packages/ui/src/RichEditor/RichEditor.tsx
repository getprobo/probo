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

import { PaperclipIcon } from "@phosphor-icons/react";
import { type Extensions } from "@tiptap/core";
import { Document } from "@tiptap/extension-document";
import { BulletList, ListKeymap, OrderedList, TaskList } from "@tiptap/extension-list";
import { Text } from "@tiptap/extension-text";
import { Dropcursor, UndoRedo } from "@tiptap/extensions";
import { type Content, Editor, EditorContent, type JSONContent, useEditor } from "@tiptap/react";
import { type ComponentProps, type ReactNode, useCallback, useEffect, useLayoutEffect, useMemo, useRef } from "react";
import { useTranslation } from "react-i18next";
import { tv } from "tailwind-variants";

import useToast from "../v2/Toaster/useToast";

import { anyAttachmentAccept, fileAccept, imageAccept } from "./attachments";
import {
  BlockquoteExtension,
  BoldExtension,
  CodeExtension,
  HardBreakExtension,
  HeadingExtension,
  HorizontalRuleExtension,
  ItalicExtension,
  ListItemExtension,
  ParagraphExtension,
  StrikeExtension,
  TableCellExtension,
  TableExtension,
  TableHeaderExtension,
  TableRowExtension,
  TaskItemExtension,
  UnderlineExtension,
} from "./automergeExtensions";
import { AutomergeMappingExtension } from "./automergeMapping";
import { AutomergeTableStructureExtension } from "./AutomergeTableStructureExtension";
import { AutomergeUnknownBlockExtension } from "./AutomergeUnknownBlockExtension";
import { BlockMenu } from "./BlockMenu/BlockMenu";
import { BubbleMenu } from "./BubbleMenu";
import { CodeBlockExtension } from "./CodeBlockExtension";
import { ContentUploadExtension } from "./ContentUploadExtension";
import { FileExtension } from "./FileExtension";
import {
  createRichEditorHistoryExtension,
  type RichEditorHistoryHighlight,
} from "./historyDecorations";
import { ImageExtension } from "./ImageExtension";
import { LinkExtension } from "./LinkExtension";
import { SuggestionDelete, SuggestionInsert } from "./SuggestionExtension";
import { MarkdownPasteExtension } from "./MarkdownPasteExtension";
import { getMentionStorage, MentionExtension, type MentionSource } from "./MentionExtension";
import { MentionMenu } from "./MentionMenu/MentionMenu";
import { OptionsMenu } from "./OptionsMenu/OptionsMenu";
import { PlaceholderExtension, setPlaceholder } from "./PlaceholderExtension";
import { SlashCommandExtension } from "./SlashCommandExtension";
import { submitOnModEnter } from "./submitOnModEnter";
import { TableCellMenu } from "./TableCellMenu/TableCellMenu";
import { TableColumnMenu } from "./TableColumnMenu/TableColumnMenu";
import { TableRowMenu } from "./TableRowMenu/TableRowMenu";
import { TableSelectionOverlay } from "./TableSelectionOverlay";
import { TaskItemFocusExtension } from "./TaskItemFocusExtension";
import {
  type AttachmentUploadError,
  type UploadAttachment,
  uploadAttachments,
} from "./uploadAttachments";

const coreExtensions = [
  Document,
  ParagraphExtension,
  Text,
  HeadingExtension,
  BoldExtension,
  ItalicExtension,
  StrikeExtension,
  UnderlineExtension,
  CodeExtension,
  CodeBlockExtension,
  LinkExtension,
  SuggestionInsert,
  SuggestionDelete,
  SlashCommandExtension,
  MentionExtension,
  BlockquoteExtension,
  BulletList,
  OrderedList,
  ListItemExtension,
  TaskList,
  TaskItemExtension,
  TaskItemFocusExtension,
  ListKeymap,
  HorizontalRuleExtension,
  HardBreakExtension,
  Dropcursor.configure({
    color: "#0081f1",
    width: 2,
  }),
  UndoRedo,
  TableExtension,
  TableCellExtension,
  TableHeaderExtension,
  TableRowExtension,
  AutomergeUnknownBlockExtension,
  MarkdownPasteExtension,
  PlaceholderExtension,
];

export const richEditorExtensions = [
  ...coreExtensions,
  ImageExtension,
  FileExtension,
  ContentUploadExtension,
];

export const richEditorCollaborationExtensions = [
  ...coreExtensions,
  ImageExtension,
  FileExtension,
  ContentUploadExtension,
  AutomergeMappingExtension,
  AutomergeTableStructureExtension,
];

const richEditorCommentExtensions = [
  Document,
  Paragraph,
  Text,
  Bold,
  Italic,
  Strike,
  Code,
  LinkExtension,
  Blockquote,
  BulletList,
  OrderedList,
  ListItem,
  ListKeymap,
  HardBreak,
  UndoRedo,
  MentionExtension,
  PlaceholderExtension,
];

const richEditorVariants = tv({
  base: ["relative flex flex-1 flex-col min-w-0 overflow-auto py-14 pr-8 bg-level-1 shadow-base"],
  variants: {
    disabled: {
      true: "pl-8",
      false: "pl-14",
    },
    compact: {
      true: "",
      false: "",
    },
    controls: {
      true: "",
      false: "",
    },
  },
  compoundVariants: [
    { compact: true, disabled: false, class: "pl-12" },
    { controls: false, class: "!pl-0" },
  ],
});

const attachmentToolbar = tv({
  slots: {
    bar: "mt-2 flex shrink-0 items-center justify-between gap-2",
    attach: [
      "inline-flex size-5 shrink-0 items-center justify-center rounded-full",
      "bg-sand-3 text-sand-11 hover:bg-sand-4",
      "cursor-pointer outline-none transition-colors",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
      "disabled:pointer-events-none disabled:opacity-50",
      "[&_svg]:size-3",
    ],
  },
});

function stripNonTextMarks(node: JSONContent) {
  if (node.type !== "text") delete node.marks;
  node.content?.forEach(stripNonTextMarks);
}

function hasContentUpload(node: JSONContent): boolean {
  if (node.type === "contentUpload") {
    return true;
  }

  return node.content?.some(hasContentUpload) ?? false;
}

function withoutContentUploads(node: JSONContent): JSONContent {
  if (!node.content) {
    return node;
  }

  const content = node.content
    .filter(child => child.type !== "contentUpload")
    .map(withoutContentUploads);
  const needsBlock = node.type === "doc"
    || node.type === "blockquote"
    || node.type === "listItem"
    || node.type === "taskItem"
    || node.type === "tableCell"
    || node.type === "tableHeader";

  return {
    ...node,
    content: needsBlock && content.length === 0 ? [{ type: "paragraph" }] : content,
  };
}

type RichEditorProps = ComponentProps<"div"> & {
  content: string;
  disabled?: boolean;
  // Pulls the block handles closer to the text and uses a shorter left gutter.
  compact?: boolean;
  // When false, the plus and drag handles are not mounted. Slash and selection menus stay.
  controls?: boolean;
  // Paperclip for images and files. Task composers hide the block handles.
  attachments?: boolean;
  // Right side of the attachment row, such as a comment send button.
  toolbar?: ReactNode;
  // TipTap binds Mod-Enter to a hard break. Composers override it to submit.
  onModEnter?: () => void;
  // Restricts the schema to comment-safe nodes (paragraphs, lists, blockquotes, mentions).
  preset?: "full" | "comment";
  placeholder?: string;
  onChangeContent?: (content: string) => void;
  onUploadingChange?: (uploading: boolean) => void;
  onUploadFile?: UploadAttachment;
  mentionSource?: MentionSource;
  historyHighlights?: RichEditorHistoryHighlight[];
  editorExtensions?: Extensions;
  initialContent?: Content;
  onPresence?: (anchor: number, head: number) => void;
  onEditor?: (editor: Editor | null) => void;
  formatting?: boolean;
  suspendContentChanges?: boolean;
  // Review cards rendered inside the scrollport so they move with the page.
  margin?: ReactNode;
};

function parseContent(content: string): Content {
  if (!content) {
    return "";
  }

  try {
    return JSON.parse(content) as Content;
  } catch {
    return "";
  }
}

export function RichEditor(props: RichEditorProps) {
  const {
    className,
    content,
    disabled = false,
    compact = false,
    controls = true,
    attachments = false,
    toolbar,
    onModEnter,
    preset = "full",
    placeholder,
    onChangeContent,
    onUploadingChange,
    onUploadFile,
    mentionSource,
    historyHighlights,
    editorExtensions: editorExtensionOverride,
    initialContent: initialContentOverride,
    onPresence,
    onEditor,
    formatting = true,
    suspendContentChanges = false,
    margin,
    ...divProps
  } = props;
  const { t } = useTranslation();
  const uploadRef = useRef<UploadAttachment | null>(null);
  const onChangeContentRef = useRef(onChangeContent);
  const onModEnterRef = useRef(onModEnter);
  const onUploadingChangeRef = useRef(onUploadingChange);
  const uploadingRef = useRef(false);
  const emittedRef = useRef<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const toast = useToast();
  const reportUploadErrorRef = useRef<(error: AttachmentUploadError) => void>(() => {});
  const canUpload = Boolean(onUploadFile && !disabled);

  useLayoutEffect(() => {
    onChangeContentRef.current = onChangeContent;
    onModEnterRef.current = onModEnter;
    onUploadingChangeRef.current = onUploadingChange;
  }, [onChangeContent, onModEnter, onUploadingChange]);

  useEffect(() => {
    return () => {
      if (!uploadingRef.current) {
        return;
      }

      uploadingRef.current = false;
      onUploadingChangeRef.current?.(false);
    };
  }, []);

  useLayoutEffect(() => {
    if (onUploadFile == null || disabled) {
      uploadRef.current = null;
    } else {
      uploadRef.current = onUploadFile;
    }

    reportUploadErrorRef.current = (error) => {
      toast.add({
        title: t("ui.richEditor.uploadFailed"),
        description: describeUploadError(t, error),
        type: "error",
      });
    };
  }, [disabled, onUploadFile, t, toast]);

  const reportUploadError = useCallback((error: AttachmentUploadError) => {
    reportUploadErrorRef.current(error);
  }, []);

  const handleUpdate = useCallback(
    ({ editor }: { editor: Editor }) => {
      if (editor.isDestroyed) {
        return;
      }

      const raw = editor.getJSON();
      const uploading = hasContentUpload(raw);
      const becameSettled = uploadingRef.current && !uploading;
      if (uploading !== uploadingRef.current) {
        uploadingRef.current = uploading;
        onUploadingChangeRef.current?.(uploading);
      }

      if (suspendContentChanges) {
        return;
      }

      const json = withoutContentUploads(raw);
      stripNonTextMarks(json);

      const next = JSON.stringify(json);
      if (next === emittedRef.current && !becameSettled) {
        return;
      }

      emittedRef.current = next;
      onChangeContentRef.current?.(next);
    },
    [suspendContentChanges],
  );

  const editorExtensions = useMemo(
    () => {
      if (editorExtensionOverride) {
        return editorExtensionOverride;
      }

      const base = preset === "comment"
        ? richEditorCommentExtensions
        : richEditorExtensions;
      if (!historyHighlights?.length) {
        return base;
      }

      return [...base, createRichEditorHistoryExtension(historyHighlights)];
    },
    [editorExtensionOverride, historyHighlights, preset],
  );
  const initialContent = initialContentOverride ?? parseContent(content);

  const editor = useEditor({
    editorProps: {
      attributes: {
        class: "min-h-full flex-1",
      },
      handleKeyDown(_view, event) {
        return submitOnModEnter(event, onModEnterRef.current);
      },
      handlePaste(view, event) {
        const upload = uploadRef.current;
        if (!upload || !view.editable) {
          return false;
        }

        const files = Array.from(event.clipboardData?.files ?? []);
        if (files.length === 0) {
          return false;
        }

        event.preventDefault();
        void uploadAttachments(view, files, view.state.selection.from, upload, reportUploadError);

        return true;
      },
      handleDrop(view, event, _slice, moved) {
        if (moved) {
          return false;
        }

        const upload = uploadRef.current;
        if (!upload || !view.editable) {
          return false;
        }

        const files = Array.from(event.dataTransfer?.files ?? []);
        if (files.length === 0) {
          return false;
        }

        event.preventDefault();
        const coords = view.posAtCoords({ left: event.clientX, top: event.clientY });
        void uploadAttachments(
          view,
          files,
          coords?.pos ?? view.state.selection.from,
          upload,
          reportUploadError,
        );

        return true;
      },
    },
    editable: !disabled,
    extensions: editorExtensions,
    content: initialContent,
    onCreate: ({ editor: created }) => {
      const json = withoutContentUploads(created.getJSON());
      stripNonTextMarks(json);
      emittedRef.current = JSON.stringify(json);
      onEditor?.(created);
      onPresence?.(
        created.state.selection.anchor,
        created.state.selection.head,
      );
    },
    onDestroy: () => {
      onEditor?.(null);
    },
    onUpdate: handleUpdate,
    onSelectionUpdate: ({ editor }) => {
      onPresence?.(
        editor.state.selection.anchor,
        editor.state.selection.head,
      );
    },
  }, [editorExtensions]);

  const openFilePicker = useCallback((kind: "image" | "file" | "any") => {
    const input = fileInputRef.current;
    const upload = uploadRef.current;
    if (!input || !upload || !editor) {
      return;
    }

    input.accept = kind === "image"
      ? imageAccept
      : kind === "file"
        ? fileAccept
        : anyAttachmentAccept;
    input.value = "";
    input.click();
  }, [editor]);

  useLayoutEffect(() => {
    if (!editor) {
      return;
    }

    setPlaceholder(editor, placeholder);
  }, [editor, placeholder]);

  useLayoutEffect(() => {
    if (!editor) {
      return;
    }

    const storage = getMentionStorage(editor);
    if (storage) {
      storage.source = mentionSource ?? null;
    }
  }, [editor, mentionSource]);

  useEffect(() => {
    if (!editor) {
      return;
    }

    editor.setEditable(!disabled, false);
  }, [editor, disabled]);

  if (!editor) return null;

  const showAttachment = attachments && onUploadFile != null;
  const showToolbar = showAttachment || toolbar != null;
  const { bar, attach } = attachmentToolbar();

  return (
    <div
      className={richEditorVariants({ className, disabled, compact, controls })}
      {...divProps}
      data-review-scroll={margin ? "" : undefined}
    >
      {!disabled && formatting
        && (
          <>
            <BubbleMenu editor={editor} />
            {preset !== "comment" && (
              <>
                <BlockMenu
                  editor={editor}
                  compact={compact}
                  handles={controls}
                  onPickFiles={canUpload ? openFilePicker : undefined}
                />
                {controls && <OptionsMenu editor={editor} compact={compact} />}
                <TableSelectionOverlay editor={editor} />
                <TableCellMenu editor={editor} />
                <TableColumnMenu editor={editor} />
                <TableRowMenu editor={editor} />
              </>
            )}
            <MentionMenu editor={editor} />
          </>
        )}

      <EditorContent className="flex min-h-0 flex-1 flex-col" editor={editor} />
      {margin}
      {showToolbar && (
        <div className={bar()}>
          {showAttachment && (
            <button
              type="button"
              className={attach()}
              aria-label={t("ui.richEditor.attachFiles")}
              disabled={!canUpload}
              onMouseDown={(event) => {
                event.preventDefault();
              }}
              onClick={() => {
                openFilePicker("any");
              }}
            >
              <PaperclipIcon />
            </button>
          )}
          {toolbar}
        </div>
      )}
      <input
        ref={fileInputRef}
        type="file"
        multiple
        className="hidden"
        onChange={(event) => {
          const upload = uploadRef.current;
          const files = Array.from(event.target.files ?? []);
          event.target.value = "";
          if (!upload || files.length === 0) {
            return;
          }

          void uploadAttachments(
            editor.view,
            files,
            editor.state.selection.from,
            upload,
            reportUploadError,
          );
        }}
      />
    </div>
  );
}

function describeUploadError(
  t: (key: string, options: { fileName: string }) => string,
  error: AttachmentUploadError,
) {
  switch (error.reason) {
    case "tooLarge":
      return t("ui.richEditor.tooLarge", { fileName: error.fileName });
    case "unsupported":
      return t("ui.richEditor.unsupported", { fileName: error.fileName });
    case "failed":
      return t("ui.richEditor.couldNotUpload", { fileName: error.fileName });
  }
}
