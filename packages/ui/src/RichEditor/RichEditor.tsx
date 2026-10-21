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

import { type Extensions } from "@tiptap/core";
import { Blockquote } from "@tiptap/extension-blockquote";
import { Bold } from "@tiptap/extension-bold";
import { Code } from "@tiptap/extension-code";
import { Document } from "@tiptap/extension-document";
import { HardBreak } from "@tiptap/extension-hard-break";
import { Heading } from "@tiptap/extension-heading";
import { HorizontalRule } from "@tiptap/extension-horizontal-rule";
import { Italic } from "@tiptap/extension-italic";
import { BulletList, ListItem, ListKeymap, OrderedList } from "@tiptap/extension-list";
import { Paragraph } from "@tiptap/extension-paragraph";
import { Strike } from "@tiptap/extension-strike";
import { TableKit } from "@tiptap/extension-table";
import { Text } from "@tiptap/extension-text";
import { Underline } from "@tiptap/extension-underline";
import { Dropcursor, UndoRedo } from "@tiptap/extensions";
import { type Content, Editor, EditorContent, type JSONContent, useEditor } from "@tiptap/react";
import { type ComponentProps, useCallback, useEffect, useLayoutEffect, useMemo, useRef } from "react";
import { useTranslation } from "react-i18next";
import { tv } from "tailwind-variants";

import useToast from "../v2/Toaster/useToast";

import { fileAccept, imageAccept } from "./attachments";
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
import { MarkdownPasteExtension } from "./MarkdownPasteExtension";
import { getMentionStorage, MentionExtension, type MentionSource } from "./MentionExtension";
import { MentionMenu } from "./MentionMenu/MentionMenu";
import { OptionsMenu } from "./OptionsMenu/OptionsMenu";
import { PlaceholderExtension, setPlaceholder } from "./PlaceholderExtension";
import { SlashCommandExtension } from "./SlashCommandExtension";
import { TableCellMenu } from "./TableCellMenu/TableCellMenu";
import { TableColumnMenu } from "./TableColumnMenu/TableColumnMenu";
import { TableRowMenu } from "./TableRowMenu/TableRowMenu";
import { TableSelectionOverlay } from "./TableSelectionOverlay";
import {
  type AttachmentUploadError,
  type UploadAttachment,
  uploadAttachments,
} from "./uploadAttachments";

const tableExtension = TableKit.configure({
  table: { resizable: true },
});

const coreExtensions = [
  Document,
  Paragraph,
  Text,
  Heading,
  Bold,
  Italic,
  Strike,
  Underline,
  Code,
  CodeBlockExtension,
  LinkExtension,
  SlashCommandExtension,
  MentionExtension,
  Blockquote,
  BulletList,
  OrderedList,
  ListItem,
  ListKeymap,
  HorizontalRule,
  HardBreak,
  Dropcursor.configure({
    color: "#0081f1",
    width: 2,
  }),
  UndoRedo,
  tableExtension,
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
  AutomergeTableStructureExtension,
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
  placeholder?: string;
  onChangeContent?: (content: string) => void;
  onUploadingChange?: (uploading: boolean) => void;
  onUploadFile?: UploadAttachment;
  historyHighlights?: RichEditorHistoryHighlight[];
  editorExtensions?: Extensions;
  initialContent?: Content;
  onPresence?: (anchor: number, head: number) => void;
  suspendContentChanges?: boolean;
  mentionSource?: MentionSource;
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
    placeholder,
    onChangeContent,
    onUploadingChange,
    onUploadFile,
    historyHighlights,
    editorExtensions: editorExtensionOverride,
    initialContent: initialContentOverride,
    onPresence,
    suspendContentChanges = false,
    mentionSource,
    ...divProps
  } = props;
  const { t } = useTranslation();
  const uploadRef = useRef<UploadAttachment | null>(null);
  const onChangeContentRef = useRef(onChangeContent);
  const onUploadingChangeRef = useRef(onUploadingChange);
  const uploadingRef = useRef(false);
  const emittedRef = useRef<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const toast = useToast();
  const reportUploadErrorRef = useRef<(error: AttachmentUploadError) => void>(() => {});
  const canUpload = Boolean(onUploadFile && !disabled);

  useLayoutEffect(() => {
    onChangeContentRef.current = onChangeContent;
    onUploadingChangeRef.current = onUploadingChange;
  }, [onChangeContent, onUploadingChange]);

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

      if (!historyHighlights?.length) {
        return richEditorExtensions;
      }

      return [...richEditorExtensions, createRichEditorHistoryExtension(historyHighlights)];
    },
    [editorExtensionOverride, historyHighlights],
  );
  const initialContent = initialContentOverride ?? parseContent(content);

  const editor = useEditor({
    editorProps: {
      attributes: {
        class: "min-h-full flex-1",
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
      onPresence?.(
        created.state.selection.anchor,
        created.state.selection.head,
      );
    },
    onUpdate: handleUpdate,
    onSelectionUpdate: ({ editor }) => {
      onPresence?.(
        editor.state.selection.anchor,
        editor.state.selection.head,
      );
    },
  }, [editorExtensions]);

  const openFilePicker = useCallback((kind: "image" | "file") => {
    const input = fileInputRef.current;
    const upload = uploadRef.current;
    if (!input || !upload || !editor) {
      return;
    }

    input.accept = kind === "image" ? imageAccept : fileAccept;
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

  return (
    <div className={richEditorVariants({ className, disabled, compact, controls })} {...divProps}>
      {!disabled
        && (
          <>
            <BubbleMenu editor={editor} />
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
            <MentionMenu editor={editor} />
          </>
        )}

      <EditorContent className="flex min-h-0 flex-1 flex-col" editor={editor} />
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
