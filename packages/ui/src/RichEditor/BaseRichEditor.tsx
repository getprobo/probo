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
import { type Content, Editor, EditorContent, type JSONContent, useEditor } from "@tiptap/react";
import { type ComponentProps, type MouseEvent as ReactMouseEvent, type ReactNode, useCallback, useEffect, useLayoutEffect, useMemo, useRef } from "react";
import { useTranslation } from "react-i18next";
import { tv } from "tailwind-variants";

import useToast from "../v2/Toaster/useToast";

import { anyAttachmentAccept, fileAccept, imageAccept } from "./attachments";
import { BlockMenu } from "./BlockMenu/BlockMenu";
import { BubbleMenu } from "./BubbleMenu";
import {
  createRichEditorHistoryExtension,
  type RichEditorHistoryHighlight,
} from "./historyDecorations";
import { getMentionStorage, type MentionSource } from "./MentionExtension";
import { MentionMenu } from "./MentionMenu/MentionMenu";
import { OptionsMenu } from "./OptionsMenu/OptionsMenu";
import { setPlaceholder } from "./PlaceholderExtension";
import { submitOnModEnter } from "./submitOnModEnter";
import { TableCellMenu } from "./TableCellMenu/TableCellMenu";
import { TableColumnMenu } from "./TableColumnMenu/TableColumnMenu";
import { TableRowMenu } from "./TableRowMenu/TableRowMenu";
import { TableSelectionOverlay } from "./TableSelectionOverlay";
import {
  type AttachmentUploadError,
  type UploadAttachment,
  uploadAttachments,
} from "./uploadAttachments";

const richEditorVariants = tv({
  base: "relative flex min-w-0 flex-col",
  variants: {
    page: {
      true: "flex-1 overflow-auto bg-level-1 py-14 pr-8 shadow-base",
      false: "",
    },
    disabled: {
      true: "",
      false: "pointer-events-auto",
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
    { page: true, disabled: true, class: "pl-8" },
    { page: true, disabled: false, class: "pl-14" },
    { page: true, compact: true, disabled: false, class: "pl-12" },
    { page: true, controls: false, class: "!pl-0" },
  ],
  defaultVariants: {
    page: true,
  },
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

export type RichEditorProps = ComponentProps<"div"> & {
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
  // Focus the caret when the editor mounts. Comment composers use this so
  // typing starts in the new box instead of the document.
  autoFocus?: boolean;
  // Document page chrome (gutter, paper background, page padding). Comment
  // editors turn this off.
  page?: boolean;
  formatting?: boolean;
  // While suggesting, formatting is off but table rows and columns can still
  // be inserted or deleted as tracked changes.
  suggestingMenus?: boolean;
  suspendContentChanges?: boolean;
  // Review cards rendered inside the scrollport so they move with the page.
  margin?: ReactNode;
};

type BaseRichEditorProps = RichEditorProps & {
  extensions: Extensions;
  blockMenu: boolean;
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

export function BaseRichEditor(props: BaseRichEditorProps) {
  const {
    className,
    content,
    disabled = false,
    compact = false,
    controls = true,
    attachments = false,
    toolbar,
    onModEnter,
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
    autoFocus = false,
    page = true,
    formatting = true,
    suggestingMenus = false,
    suspendContentChanges = false,
    margin,
    extensions,
    blockMenu,
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

      if (!historyHighlights?.length) {
        return extensions;
      }

      return [...extensions, createRichEditorHistoryExtension(historyHighlights)];
    },
    [editorExtensionOverride, extensions, historyHighlights],
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

  // Menus that mount with the same render read editor.isEditable in their own
  // effects, which run before this component's passive effects.
  useLayoutEffect(() => {
    if (!editor) {
      return;
    }

    editor.setEditable(!disabled, false);
  }, [editor, disabled]);

  useEffect(() => {
    if (!autoFocus || !editor || editor.isDestroyed || disabled) return;
    editor.commands.focus("end");
  }, [autoFocus, disabled, editor]);

  if (!editor) return null;

  const { onMouseDown, ...editorDomProps } = divProps;
  const showAttachment = attachments && onUploadFile != null;
  const showToolbar = showAttachment || toolbar != null;
  const { bar, attach } = attachmentToolbar();

  return (
    <div
      className={richEditorVariants({ className, disabled, compact, controls, page })}
      {...editorDomProps}
      data-review-scroll={margin ? "" : undefined}
      onMouseDown={(event) => {
        onMouseDown?.(event);
        if (disabled) return;
        reclaimSwallowedEditorFocus(editor, event);
      }}
    >
      {!disabled && formatting
        && (
          <>
            <BubbleMenu editor={editor} />
            {blockMenu && (
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
      {!disabled && !formatting && suggestingMenus && blockMenu && (
        <>
          <TableColumnMenu editor={editor} />
          <TableRowMenu editor={editor} />
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

// Clicks on editor padding are lost when an ancestor has pointer-events: none.
// Put the caret in the editor instead of letting the click fall through.
function reclaimSwallowedEditorFocus(
  editor: Editor,
  event: ReactMouseEvent<HTMLDivElement>,
) {
  if (event.defaultPrevented || editor.isDestroyed) return;
  const target = event.target;
  if (!(target instanceof Element) || editor.view.dom.contains(target)) return;
  const shell = event.currentTarget;
  if (target !== shell && target !== editor.view.dom.parentElement) return;

  let parent = shell.parentElement;
  while (parent) {
    if (getComputedStyle(parent).pointerEvents === "none") {
      event.preventDefault();
      editor.commands.focus();
      return;
    }
    parent = parent.parentElement;
  }
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
