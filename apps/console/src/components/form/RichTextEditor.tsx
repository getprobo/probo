import { useTranslate } from "@probo/i18n";
import color from "@tiptap/extension-color";
import highlight from "@tiptap/extension-highlight";
import link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import subscript from "@tiptap/extension-subscript";
import superscript from "@tiptap/extension-superscript";
import textAlign from "@tiptap/extension-text-align";
import { TextStyle } from "@tiptap/extension-text-style";
import underline from "@tiptap/extension-underline";
import { EditorContent, useEditor } from "@tiptap/react";
import starterKit from "@tiptap/starter-kit";
import { clsx } from "clsx";
import { useCallback, useState } from "react";

type Props = {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
};

function ToolbarButton({
  onClick,
  active,
  disabled,
  title,
  children,
}: {
  onClick: () => void;
  active?: boolean;
  disabled?: boolean;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      title={title}
      onClick={onClick}
      disabled={disabled}
      className={clsx(
        "size-8 flex items-center justify-center rounded-lg transition-colors",
        disabled && "opacity-30 cursor-not-allowed",
        !disabled && active && "bg-primary/10 text-primary",
        !disabled && !active && "text-txt-secondary hover:bg-subtle hover:text-txt-primary",
      )}
    >
      {children}
    </button>
  );
}

function ToolbarDivider() {
  return <div className="w-px h-5 bg-border-low mx-1.5 shrink-0" />;
}

function ToolbarGroup({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center gap-0.5">{children}</div>;
}

function Ico({ d, size = 16 }: { d: string; size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d={d} />
    </svg>
  );
}

function LinkPopover({
  onSubmit,
  onClose,
  initialUrl,
}: {
  onSubmit: (url: string) => void;
  onClose: () => void;
  initialUrl: string;
}) {
  const { __ } = useTranslate();
  const [url, setUrl] = useState(initialUrl);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (url.trim()) {
      onSubmit(url.trim());
    }
    onClose();
  };

  return (
    <div className="absolute top-full left-0 mt-1 z-50 bg-level-1 border border-border-low rounded-xl shadow-mid p-3 flex items-center gap-2 min-w-[320px]">
      <form onSubmit={handleSubmit} className="flex items-center gap-2 w-full">
        <input
          type="url"
          value={url}
          onChange={e => setUrl(e.target.value)}
          placeholder={__("https://example.com")}
          className="flex-1 text-sm px-3 py-1.5 rounded-lg border border-border-low bg-level-2 text-txt-primary placeholder:text-txt-tertiary focus:outline-none focus:border-primary"
          autoFocus
        />
        <button
          type="submit"
          className="text-sm font-medium px-3 py-1.5 rounded-lg bg-primary text-invert hover:bg-primary-hover transition-colors"
        >
          {__("Apply")}
        </button>
        <button
          type="button"
          onClick={onClose}
          className="text-sm font-medium px-2 py-1.5 rounded-lg text-txt-secondary hover:bg-subtle transition-colors"
        >
          {__("Cancel")}
        </button>
      </form>
    </div>
  );
}

export function RichTextEditor({ value, onChange, placeholder }: Props) {
  const { __ } = useTranslate();
  const [showLinkPopover, setShowLinkPopover] = useState(false);

  const editor = useEditor({
    extensions: [
      starterKit.configure({
        heading: { levels: [1, 2, 3] },
      }),
      underline,
      textAlign.configure({
        types: ["heading", "paragraph"],
      }),
      Placeholder.configure({
        placeholder: placeholder ?? __("Add content"),
      }),
      highlight.configure({ multicolor: true }),
      superscript,
      subscript,
      link.configure({
        openOnClick: false,
        HTMLAttributes: { class: "text-primary underline cursor-pointer" },
      }),
      TextStyle,
      color,
    ],
    content: value ?? "",
    onUpdate: ({ editor: e }) => {
      onChange?.(e.getHTML());
    },
    editorProps: {
      attributes: {
        class: "prose prose-sm max-w-none focus:outline-none min-h-[200px] px-10 py-4 text-txt-primary",
      },
    },
  });

  const handleLinkSubmit = useCallback((url: string) => {
    if (!editor) return;
    editor.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
  }, [editor]);

  if (!editor) {
    return null;
  }

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="relative flex flex-wrap items-center gap-1 px-4 py-2 border-b border-b-border-low shrink-0 bg-level-2/50">
        {/* Undo / Redo */}
        <ToolbarGroup>
          <ToolbarButton
            onClick={() => editor.chain().focus().undo().run()}
            disabled={!editor.can().undo()}
            title={__("Undo (Ctrl+Z)")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M3 7v6h6" />
              <path d="M21 17a9 9 0 0 0-9-9 9 9 0 0 0-6 2.3L3 13" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().redo().run()}
            disabled={!editor.can().redo()}
            title={__("Redo (Ctrl+Shift+Z)")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 7v6h-6" />
              <path d="M3 17a9 9 0 0 1 9-9 9 9 0 0 1 6 2.3L21 13" />
            </svg>
          </ToolbarButton>
        </ToolbarGroup>

        <ToolbarDivider />

        {/* Headings */}
        <ToolbarGroup>
          <ToolbarButton
            onClick={() => editor.chain().focus().setParagraph().run()}
            active={editor.isActive("paragraph") && !editor.isActive("heading")}
            title={__("Normal text")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M13 4v16" />
              <path d="M17 4v16" />
              <path d="M13 4H7a4 4 0 0 0 0 8h6" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
            active={editor.isActive("heading", { level: 1 })}
            title={__("Heading 1")}
          >
            <span className="text-xs font-bold leading-none">H1</span>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
            active={editor.isActive("heading", { level: 2 })}
            title={__("Heading 2")}
          >
            <span className="text-xs font-bold leading-none">H2</span>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
            active={editor.isActive("heading", { level: 3 })}
            title={__("Heading 3")}
          >
            <span className="text-xs font-bold leading-none">H3</span>
          </ToolbarButton>
        </ToolbarGroup>

        <ToolbarDivider />

        {/* Text formatting */}
        <ToolbarGroup>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleBold().run()}
            active={editor.isActive("bold")}
            title={__("Bold (Ctrl+B)")}
          >
            <Ico d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6zM6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z" />
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleItalic().run()}
            active={editor.isActive("italic")}
            title={__("Italic (Ctrl+I)")}
          >
            <Ico d="M19 4h-9M14 20H5M15 4L9 20" />
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleUnderline().run()}
            active={editor.isActive("underline")}
            title={__("Underline (Ctrl+U)")}
          >
            <Ico d="M6 3v7a6 6 0 0 0 6 6 6 6 0 0 0 6-6V3M4 21h16" />
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleStrike().run()}
            active={editor.isActive("strike")}
            title={__("Strikethrough")}
          >
            <Ico d="M16 4H9a3 3 0 0 0-2.83 4M14 12a4 4 0 0 1 0 8H6M4 12h16" />
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleCode().run()}
            active={editor.isActive("code")}
            title={__("Inline code")}
          >
            <Ico d="m16 18 6-6-6-6M8 6l-6 6 6 6" />
          </ToolbarButton>
        </ToolbarGroup>

        <ToolbarDivider />

        {/* Highlight & superscript/subscript */}
        <ToolbarGroup>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleHighlight().run()}
            active={editor.isActive("highlight")}
            title={__("Highlight")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="m9 11-6 6v3h9l3-3" />
              <path d="m22 12-4.6 4.6a2 2 0 0 1-2.8 0l-5.2-5.2a2 2 0 0 1 0-2.8L14 4" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleSuperscript().run()}
            active={editor.isActive("superscript")}
            title={__("Superscript")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="m4 19 8-8" />
              <path d="m12 19-8-8" />
              <path d="M20 12h-4c0-1.5.44-2 1.5-2.5S20 8.33 20 7c0-1.06-.75-2-2-2a1.98 1.98 0 0 0-1.62.86" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleSubscript().run()}
            active={editor.isActive("subscript")}
            title={__("Subscript")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="m4 5 8 8" />
              <path d="m12 5-8 8" />
              <path d="M20 19h-4c0-1.5.44-2 1.5-2.5S20 15.33 20 14c0-1.06-.75-2-2-2a1.98 1.98 0 0 0-1.62.86" />
            </svg>
          </ToolbarButton>
        </ToolbarGroup>

        <ToolbarDivider />

        {/* Lists */}
        <ToolbarGroup>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleBulletList().run()}
            active={editor.isActive("bulletList")}
            title={__("Bullet list")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="8" y1="6" x2="21" y2="6" />
              <line x1="8" y1="12" x2="21" y2="12" />
              <line x1="8" y1="18" x2="21" y2="18" />
              <circle cx="3.5" cy="6" r="1" fill="currentColor" stroke="none" />
              <circle cx="3.5" cy="12" r="1" fill="currentColor" stroke="none" />
              <circle cx="3.5" cy="18" r="1" fill="currentColor" stroke="none" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleOrderedList().run()}
            active={editor.isActive("orderedList")}
            title={__("Numbered list")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="10" y1="6" x2="21" y2="6" />
              <line x1="10" y1="12" x2="21" y2="12" />
              <line x1="10" y1="18" x2="21" y2="18" />
              <text x="2" y="8" fill="currentColor" stroke="none" fontSize="8" fontFamily="sans-serif" fontWeight="600">1</text>
              <text x="2" y="14" fill="currentColor" stroke="none" fontSize="8" fontFamily="sans-serif" fontWeight="600">2</text>
              <text x="2" y="20" fill="currentColor" stroke="none" fontSize="8" fontFamily="sans-serif" fontWeight="600">3</text>
            </svg>
          </ToolbarButton>
        </ToolbarGroup>

        <ToolbarDivider />

        {/* Alignment */}
        <ToolbarGroup>
          <ToolbarButton
            onClick={() => editor.chain().focus().setTextAlign("left").run()}
            active={editor.isActive({ textAlign: "left" })}
            title={__("Align left")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="15" y2="12" />
              <line x1="3" y1="18" x2="18" y2="18" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().setTextAlign("center").run()}
            active={editor.isActive({ textAlign: "center" })}
            title={__("Align center")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="6" y1="12" x2="18" y2="12" />
              <line x1="4" y1="18" x2="20" y2="18" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().setTextAlign("right").run()}
            active={editor.isActive({ textAlign: "right" })}
            title={__("Align right")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="9" y1="12" x2="21" y2="12" />
              <line x1="6" y1="18" x2="21" y2="18" />
            </svg>
          </ToolbarButton>
        </ToolbarGroup>

        <ToolbarDivider />

        {/* Blocks */}
        <ToolbarGroup>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleBlockquote().run()}
            active={editor.isActive("blockquote")}
            title={__("Blockquote")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M3 21c3 0 7-1 7-8V5c0-1.25-.756-2.017-2-2H4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2 1 0 1 0 1 1v1c0 1-1 2-2 2s-1 .008-1 1.031V21z" />
              <path d="M15 21c3 0 7-1 7-8V5c0-1.25-.757-2.017-2-2h-4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2h.75c0 2.25.25 4-2.75 4v3z" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().toggleCodeBlock().run()}
            active={editor.isActive("codeBlock")}
            title={__("Code block")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <path d="m10 10-2 2 2 2" />
              <path d="m14 10 2 2-2 2" />
            </svg>
          </ToolbarButton>
          <ToolbarButton
            onClick={() => editor.chain().focus().setHorizontalRule().run()}
            title={__("Horizontal rule")}
          >
            <Ico d="M3 12h18" />
          </ToolbarButton>
        </ToolbarGroup>

        <ToolbarDivider />

        {/* Link */}
        <ToolbarGroup>
          <ToolbarButton
            onClick={() => {
              if (editor.isActive("link")) {
                editor.chain().focus().unsetLink().run();
              } else {
                setShowLinkPopover(true);
              }
            }}
            active={editor.isActive("link")}
            title={editor.isActive("link") ? __("Remove link") : __("Add link")}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
            </svg>
          </ToolbarButton>
        </ToolbarGroup>

        {/* Link popover */}
        {showLinkPopover && (
          <LinkPopover
            initialUrl={String(editor.getAttributes("link").href ?? "")}
            onSubmit={handleLinkSubmit}
            onClose={() => setShowLinkPopover(false)}
          />
        )}
      </div>

      {/* Editor */}
      <div className="flex-1 overflow-y-auto">
        <EditorContent editor={editor} className="h-full" />
      </div>
    </div>
  );
}
