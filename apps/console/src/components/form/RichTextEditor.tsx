import { useTranslate } from "@probo/i18n";
import color from "@tiptap/extension-color";
import highlight from "@tiptap/extension-highlight";
import image from "@tiptap/extension-image";
import link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import subscript from "@tiptap/extension-subscript";
import superscript from "@tiptap/extension-superscript";
import { Table } from "@tiptap/extension-table";
import tableCell from "@tiptap/extension-table-cell";
import tableHeader from "@tiptap/extension-table-header";
import tableRow from "@tiptap/extension-table-row";
import taskItem from "@tiptap/extension-task-item";
import taskList from "@tiptap/extension-task-list";
import textAlign from "@tiptap/extension-text-align";
import { TextStyle } from "@tiptap/extension-text-style";
import typography from "@tiptap/extension-typography";
import underline from "@tiptap/extension-underline";
import { type Editor, EditorContent, useEditor } from "@tiptap/react";
import starterKit from "@tiptap/starter-kit";
import { clsx } from "clsx";
import { useCallback, useState } from "react";

type Props = {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
};

// ─── Toolbar primitives ─────────────────────────────────────────────

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
      onMouseDown={e => e.preventDefault()}
      onClick={onClick}
      disabled={disabled}
      className={clsx(
        "size-7 flex items-center justify-center rounded-md transition-colors",
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
  return <div className="w-px h-4 bg-border-low mx-1 shrink-0" />;
}

function ToolbarGroup({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center">{children}</div>;
}

// ─── SVG icon helper ────────────────────────────────────────────────

function Ico({ children, size = 15 }: { children: React.ReactNode; size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {children}
    </svg>
  );
}

// ─── Popovers ───────────────────────────────────────────────────────

function InputPopover({
  onSubmit,
  onClose,
  initialValue,
  inputPlaceholder,
  submitLabel,
}: {
  onSubmit: (val: string) => void;
  onClose: () => void;
  initialValue: string;
  inputPlaceholder: string;
  submitLabel: string;
}) {
  const { __ } = useTranslate();
  const [val, setVal] = useState(initialValue);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (val.trim()) {
      onSubmit(val.trim());
    }
    onClose();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      e.stopPropagation();
      if (val.trim()) {
        onSubmit(val.trim());
      }
      onClose();
    }
    if (e.key === "Escape") {
      onClose();
    }
  };

  return (
    <div className="absolute left-0 top-full mt-1 z-50 bg-level-1 border border-border-low rounded-xl shadow-mid p-3 min-w-85">
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={val}
          onChange={e => setVal(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={inputPlaceholder}
          className="flex-1 text-sm px-3 py-1.5 rounded-lg border border-border-low bg-level-2 text-txt-primary placeholder:text-txt-tertiary focus:outline-none focus:border-primary"
          autoFocus
        />
        <button
          type="button"
          onClick={handleSubmit}
          className="text-sm font-medium px-3 py-1.5 rounded-lg bg-primary text-invert hover:bg-primary-hover transition-colors whitespace-nowrap"
        >
          {submitLabel}
        </button>
        <button
          type="button"
          onClick={onClose}
          className="text-sm px-2 py-1.5 rounded-lg text-txt-secondary hover:bg-subtle transition-colors"
        >
          {__("Cancel")}
        </button>
      </div>
    </div>
  );
}

function ColorPicker({
  onSelect,
  onClose,
  currentColor,
}: {
  onSelect: (color: string) => void;
  onClose: () => void;
  currentColor: string;
}) {
  const { __ } = useTranslate();
  const colors = [
    { label: __("Default"), value: "" },
    { label: __("Gray"), value: "#6b7280" },
    { label: __("Brown"), value: "#92400e" },
    { label: __("Red"), value: "#dc2626" },
    { label: __("Orange"), value: "#ea580c" },
    { label: __("Yellow"), value: "#ca8a04" },
    { label: __("Green"), value: "#16a34a" },
    { label: __("Teal"), value: "#0d9488" },
    { label: __("Blue"), value: "#2563eb" },
    { label: __("Indigo"), value: "#4f46e5" },
    { label: __("Purple"), value: "#9333ea" },
    { label: __("Pink"), value: "#db2777" },
  ];

  return (
    <div className="absolute left-0 top-full mt-1 z-50 bg-level-1 border border-border-low rounded-xl shadow-mid p-2 min-w-45">
      <div className="text-xs font-medium text-txt-tertiary px-2 py-1 mb-1">{__("Text color")}</div>
      <div className="grid grid-cols-6 gap-1">
        {colors.map(c => (
          <button
            key={c.value || "default"}
            type="button"
            title={c.label}
            onMouseDown={e => e.preventDefault()}
            onClick={() => {
              onSelect(c.value);
              onClose();
            }}
            className={clsx(
              "size-7 rounded-md flex items-center justify-center transition-colors hover:ring-2 hover:ring-primary/30",
              currentColor === c.value && "ring-2 ring-primary",
            )}
          >
            {c.value
              ? <span className="size-4 rounded-full" style={{ backgroundColor: c.value }} />
              : (
                  <span className="size-4 rounded-full border border-border-low bg-level-2 flex items-center justify-center text-[9px] text-txt-tertiary">
                    A
                  </span>
                )}
          </button>
        ))}
      </div>
    </div>
  );
}

function HighlightColorPicker({
  onSelect,
  onClose,
}: {
  onSelect: (color: string) => void;
  onClose: () => void;
}) {
  const { __ } = useTranslate();
  const colors = [
    { label: __("Yellow"), value: "#fef08a" },
    { label: __("Green"), value: "#bbf7d0" },
    { label: __("Blue"), value: "#bfdbfe" },
    { label: __("Purple"), value: "#e9d5ff" },
    { label: __("Pink"), value: "#fce7f3" },
    { label: __("Orange"), value: "#fed7aa" },
    { label: __("Red"), value: "#fecaca" },
    { label: __("Gray"), value: "#e5e7eb" },
  ];

  return (
    <div className="absolute left-0 top-full mt-1 z-50 bg-level-1 border border-border-low rounded-xl shadow-mid p-2 min-w-37.5">
      <div className="text-xs font-medium text-txt-tertiary px-2 py-1 mb-1">{__("Highlight")}</div>
      <div className="grid grid-cols-4 gap-1">
        {colors.map(c => (
          <button
            key={c.value}
            type="button"
            title={c.label}
            onMouseDown={e => e.preventDefault()}
            onClick={() => {
              onSelect(c.value);
              onClose();
            }}
            className="size-7 rounded-md flex items-center justify-center hover:ring-2 hover:ring-primary/30"
          >
            <span className="size-5 rounded" style={{ backgroundColor: c.value }} />
          </button>
        ))}
      </div>
      <button
        type="button"
        onMouseDown={e => e.preventDefault()}
        onClick={() => {
          onSelect("");
          onClose();
        }}
        className="w-full text-xs text-txt-secondary hover:text-txt-primary py-1 mt-1 hover:bg-subtle rounded-md transition-colors"
      >
        {__("Remove highlight")}
      </button>
    </div>
  );
}

// ─── Toolbar rows ───────────────────────────────────────────────────

function ToolbarRow1({ editor, __: t }: { editor: Editor; __: (s: string) => string }) {
  return (
    <div className="flex items-center gap-0.5 flex-wrap">
      {/* Undo / Redo */}
      <ToolbarGroup>
        <ToolbarButton onClick={() => editor.chain().focus().undo().run()} disabled={!editor.can().undo()} title={t("Undo")}>
          <Ico>
            <path d="M3 7v6h6" />
            <path d="M21 17a9 9 0 0 0-9-9 9 9 0 0 0-6 2.3L3 13" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().redo().run()} disabled={!editor.can().redo()} title={t("Redo")}>
          <Ico>
            <path d="M21 7v6h-6" />
            <path d="M3 17a9 9 0 0 1 9-9 9 9 0 0 1 6 2.3L21 13" />
          </Ico>
        </ToolbarButton>
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Paragraph & Headings */}
      <ToolbarGroup>
        <ToolbarButton
          onClick={() => editor.chain().focus().setParagraph().run()}
          active={editor.isActive("paragraph") && !editor.isActive("heading")}
          title={t("Normal text")}
        >
          <Ico>
            <path d="M13 4v16" />
            <path d="M17 4v16" />
            <path d="M13 4H7a4 4 0 0 0 0 8h6" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} active={editor.isActive("heading", { level: 1 })} title={t("Heading 1")}>
          <span className="text-[10px] font-bold leading-none">H1</span>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} active={editor.isActive("heading", { level: 2 })} title={t("Heading 2")}>
          <span className="text-[10px] font-bold leading-none">H2</span>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()} active={editor.isActive("heading", { level: 3 })} title={t("Heading 3")}>
          <span className="text-[10px] font-bold leading-none">H3</span>
        </ToolbarButton>
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Text formatting */}
      <ToolbarGroup>
        <ToolbarButton onClick={() => editor.chain().focus().toggleBold().run()} active={editor.isActive("bold")} title={t("Bold")}>
          <Ico><path d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6zM6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z" /></Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleItalic().run()} active={editor.isActive("italic")} title={t("Italic")}>
          <Ico>
            <path d="M19 4h-9" />
            <path d="M14 20H5" />
            <path d="M15 4L9 20" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleUnderline().run()} active={editor.isActive("underline")} title={t("Underline")}>
          <Ico>
            <path d="M6 3v7a6 6 0 0 0 6 6 6 6 0 0 0 6-6V3" />
            <path d="M4 21h16" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleStrike().run()} active={editor.isActive("strike")} title={t("Strikethrough")}>
          <Ico>
            <path d="M16 4H9a3 3 0 0 0-2.83 4" />
            <path d="M14 12a4 4 0 0 1 0 8H6" />
            <path d="M4 12h16" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleCode().run()} active={editor.isActive("code")} title={t("Inline code")}>
          <Ico>
            <path d="m16 18 6-6-6-6" />
            <path d="m8 6-6 6 6 6" />
          </Ico>
        </ToolbarButton>
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Superscript / Subscript */}
      <ToolbarGroup>
        <ToolbarButton onClick={() => editor.chain().focus().toggleSuperscript().run()} active={editor.isActive("superscript")} title={t("Superscript")}>
          <Ico>
            <path d="m4 19 8-8" />
            <path d="m12 19-8-8" />
            <path d="M20 12h-4c0-1.5.44-2 1.5-2.5S20 8.33 20 7c0-1.06-.75-2-2-2a1.98 1.98 0 0 0-1.62.86" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleSubscript().run()} active={editor.isActive("subscript")} title={t("Subscript")}>
          <Ico>
            <path d="m4 5 8 8" />
            <path d="m12 5-8 8" />
            <path d="M20 19h-4c0-1.5.44-2 1.5-2.5S20 15.33 20 14c0-1.06-.75-2-2-2a1.98 1.98 0 0 0-1.62.86" />
          </Ico>
        </ToolbarButton>
      </ToolbarGroup>
    </div>
  );
}

function ToolbarRow2({
  editor,
  __: t,
  showLinkPopover,
  setShowLinkPopover,
  showColorPicker,
  setShowColorPicker,
  showHighlightPicker,
  setShowHighlightPicker,
  handleLinkSubmit,
  handleImageSubmit,
  showImagePopover,
  setShowImagePopover,
}: {
  editor: Editor;
  __: (s: string) => string;
  showLinkPopover: boolean;
  setShowLinkPopover: (v: boolean) => void;
  showColorPicker: boolean;
  setShowColorPicker: (v: boolean) => void;
  showHighlightPicker: boolean;
  setShowHighlightPicker: (v: boolean) => void;
  handleLinkSubmit: (url: string) => void;
  handleImageSubmit: (url: string) => void;
  showImagePopover: boolean;
  setShowImagePopover: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center gap-0.5 flex-wrap">
      {/* Lists */}
      <ToolbarGroup>
        <ToolbarButton onClick={() => editor.chain().focus().toggleBulletList().run()} active={editor.isActive("bulletList")} title={t("Bullet list")}>
          <Ico>
            <line x1="9" y1="6" x2="20" y2="6" />
            <line x1="9" y1="12" x2="20" y2="12" />
            <line x1="9" y1="18" x2="20" y2="18" />
            <circle cx="4.5" cy="6" r="1.5" fill="currentColor" stroke="none" />
            <circle cx="4.5" cy="12" r="1.5" fill="currentColor" stroke="none" />
            <circle cx="4.5" cy="18" r="1.5" fill="currentColor" stroke="none" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleOrderedList().run()} active={editor.isActive("orderedList")} title={t("Numbered list")}>
          <Ico>
            <line x1="10" y1="6" x2="21" y2="6" />
            <line x1="10" y1="12" x2="21" y2="12" />
            <line x1="10" y1="18" x2="21" y2="18" />
            <path d="M4 6h1v4" />
            <path d="M3 10h3" />
            <path d="M3 18h3" />
            <path d="M6 14H4c0 1 .5 1.5 1 2l-1 2" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleTaskList().run()} active={editor.isActive("taskList")} title={t("Task list")}>
          <Ico>
            <rect x="3" y="5" width="6" height="6" rx="1" />
            <path d="m3 17 2 2 4-4" />
            <line x1="13" y1="6" x2="21" y2="6" />
            <line x1="13" y1="12" x2="21" y2="12" />
            <line x1="13" y1="18" x2="21" y2="18" />
          </Ico>
        </ToolbarButton>
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Alignment */}
      <ToolbarGroup>
        <ToolbarButton onClick={() => editor.chain().focus().setTextAlign("left").run()} active={editor.isActive({ textAlign: "left" })} title={t("Align left")}>
          <Ico>
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="3" y1="12" x2="15" y2="12" />
            <line x1="3" y1="18" x2="18" y2="18" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().setTextAlign("center").run()} active={editor.isActive({ textAlign: "center" })} title={t("Align center")}>
          <Ico>
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="6" y1="12" x2="18" y2="12" />
            <line x1="4" y1="18" x2="20" y2="18" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().setTextAlign("right").run()} active={editor.isActive({ textAlign: "right" })} title={t("Align right")}>
          <Ico>
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="9" y1="12" x2="21" y2="12" />
            <line x1="6" y1="18" x2="21" y2="18" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().setTextAlign("justify").run()} active={editor.isActive({ textAlign: "justify" })} title={t("Justify")}>
          <Ico>
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <line x1="3" y1="18" x2="21" y2="18" />
          </Ico>
        </ToolbarButton>
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Text & highlight colors */}
      <ToolbarGroup>
        <div className="relative">
          <ToolbarButton
            onClick={() => {
              setShowColorPicker(!showColorPicker);
              setShowHighlightPicker(false);
              setShowLinkPopover(false);
            }}
            title={t("Text color")}
          >
            <Ico>
              <path d="M4 20h16" />
              <path d="m8.5 6 3.5 10" />
              <path d="M12 16l3.5-10" />
              <path d="M6 16h12" />
            </Ico>
          </ToolbarButton>
          {showColorPicker && (
            <ColorPicker
              currentColor={String(editor.getAttributes("textStyle").color ?? "")}
              onSelect={(c) => {
                if (c) {
                  editor.chain().focus().setColor(c).run();
                } else {
                  editor.chain().focus().unsetColor().run();
                }
              }}
              onClose={() => setShowColorPicker(false)}
            />
          )}
        </div>
        <div className="relative">
          <ToolbarButton
            onClick={() => {
              setShowHighlightPicker(!showHighlightPicker);
              setShowColorPicker(false);
              setShowLinkPopover(false);
            }}
            active={editor.isActive("highlight")}
            title={t("Highlight")}
          >
            <Ico>
              <path d="m9 11-6 6v3h9l3-3" />
              <path d="m22 12-4.6 4.6a2 2 0 0 1-2.8 0l-5.2-5.2a2 2 0 0 1 0-2.8L14 4" />
            </Ico>
          </ToolbarButton>
          {showHighlightPicker && (
            <HighlightColorPicker
              onSelect={(c) => {
                if (c) {
                  editor.chain().focus().toggleHighlight({ color: c }).run();
                } else {
                  editor.chain().focus().unsetHighlight().run();
                }
              }}
              onClose={() => setShowHighlightPicker(false)}
            />
          )}
        </div>
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Block elements */}
      <ToolbarGroup>
        <ToolbarButton onClick={() => editor.chain().focus().toggleBlockquote().run()} active={editor.isActive("blockquote")} title={t("Blockquote")}>
          <Ico>
            <path d="M3 21c3 0 7-1 7-8V5c0-1.25-.756-2.017-2-2H4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2 1 0 1 0 1 1v1c0 1-1 2-2 2s-1 .008-1 1.031V21z" />
            <path d="M15 21c3 0 7-1 7-8V5c0-1.25-.757-2.017-2-2h-4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2h.75c0 2.25.25 4-2.75 4v3z" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().toggleCodeBlock().run()} active={editor.isActive("codeBlock")} title={t("Code block")}>
          <Ico>
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
            <path d="m10 10-2 2 2 2" />
            <path d="m14 10 2 2-2 2" />
          </Ico>
        </ToolbarButton>
        <ToolbarButton onClick={() => editor.chain().focus().setHorizontalRule().run()} title={t("Horizontal rule")}>
          <Ico><path d="M3 12h18" /></Ico>
        </ToolbarButton>
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Link & Image */}
      <ToolbarGroup>
        <div className="relative">
          <ToolbarButton
            onClick={() => {
              if (editor.isActive("link")) {
                editor.chain().focus().unsetLink().run();
              } else {
                setShowLinkPopover(!showLinkPopover);
                setShowColorPicker(false);
                setShowHighlightPicker(false);
              }
            }}
            active={editor.isActive("link")}
            title={editor.isActive("link") ? t("Remove link") : t("Add link")}
          >
            <Ico>
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
            </Ico>
          </ToolbarButton>
          {showLinkPopover && (
            <InputPopover
              initialValue={String(editor.getAttributes("link").href ?? "")}
              onSubmit={handleLinkSubmit}
              onClose={() => setShowLinkPopover(false)}
              inputPlaceholder="https://example.com"
              submitLabel={t("Apply")}
            />
          )}
        </div>
        <ToolbarButton
          onClick={() => {
            setShowImagePopover(!showImagePopover);
            setShowLinkPopover(false);
            setShowColorPicker(false);
            setShowHighlightPicker(false);
          }}
          title={t("Insert image")}
        >
          <Ico>
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
            <circle cx="8.5" cy="8.5" r="1.5" />
            <path d="m21 15-5-5L5 21" />
          </Ico>
        </ToolbarButton>
        {showImagePopover && (
          <InputPopover
            initialValue=""
            onSubmit={handleImageSubmit}
            onClose={() => setShowImagePopover(false)}
            inputPlaceholder={t("https://example.com/image.png")}
            submitLabel={t("Insert")}
          />
        )}
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Table */}
      <ToolbarGroup>
        <ToolbarButton onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()} title={t("Insert table")}>
          <Ico>
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
            <line x1="3" y1="9" x2="21" y2="9" />
            <line x1="3" y1="15" x2="21" y2="15" />
            <line x1="9" y1="3" x2="9" y2="21" />
            <line x1="15" y1="3" x2="15" y2="21" />
          </Ico>
        </ToolbarButton>
        {editor.isActive("table") && (
          <>
            <ToolbarButton onClick={() => editor.chain().focus().addColumnAfter().run()} title={t("Add column")}>
              <span className="text-[9px] font-bold leading-none">+Col</span>
            </ToolbarButton>
            <ToolbarButton onClick={() => editor.chain().focus().addRowAfter().run()} title={t("Add row")}>
              <span className="text-[9px] font-bold leading-none">+Row</span>
            </ToolbarButton>
            <ToolbarButton onClick={() => editor.chain().focus().deleteColumn().run()} title={t("Delete column")}>
              <span className="text-[9px] font-bold leading-none text-danger">-Col</span>
            </ToolbarButton>
            <ToolbarButton onClick={() => editor.chain().focus().deleteRow().run()} title={t("Delete row")}>
              <span className="text-[9px] font-bold leading-none text-danger">-Row</span>
            </ToolbarButton>
            <ToolbarButton onClick={() => editor.chain().focus().deleteTable().run()} title={t("Delete table")}>
              <Ico>
                <path d="M3 6h18" />
                <path d="M8 6V4h8v2" />
                <path d="m19 6-.867 12.142A2 2 0 0 1 16.138 20H7.862a2 2 0 0 1-1.995-1.858L5 6" />
              </Ico>
            </ToolbarButton>
          </>
        )}
      </ToolbarGroup>

      <ToolbarDivider />

      {/* Clear formatting */}
      <ToolbarGroup>
        <ToolbarButton
          onClick={() => editor.chain().focus().clearNodes().unsetAllMarks().run()}
          title={t("Clear formatting")}
        >
          <Ico>
            <path d="M4 7V4h16v3" />
            <path d="M9 20h6" />
            <path d="M12 4v16" />
            <path d="m17 3-10 18" />
          </Ico>
        </ToolbarButton>
      </ToolbarGroup>
    </div>
  );
}

// ─── Main component ─────────────────────────────────────────────────

export function RichTextEditor({ value, onChange, placeholder }: Props) {
  const { __ } = useTranslate();
  const [showLinkPopover, setShowLinkPopover] = useState(false);
  const [showColorPicker, setShowColorPicker] = useState(false);
  const [showHighlightPicker, setShowHighlightPicker] = useState(false);
  const [showImagePopover, setShowImagePopover] = useState(false);

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
      image.configure({
        inline: false,
        allowBase64: false,
      }),
      Table.configure({ resizable: true }),
      tableRow,
      tableCell,
      tableHeader,
      taskList,
      taskItem.configure({ nested: true }),
      typography,
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

  const handleImageSubmit = useCallback((url: string) => {
    if (!editor) return;
    editor.chain().focus().setImage({ src: url }).run();
  }, [editor]);

  if (!editor) {
    return null;
  }

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="relative shrink-0 border-b border-b-border-low bg-level-2/50 overflow-visible">
        <div className="px-3 py-1.5 border-b border-b-border-low/50">
          <ToolbarRow1 editor={editor} __={__} />
        </div>
        <div className="relative px-3 py-1.5 overflow-visible">
          <ToolbarRow2
            editor={editor}
            __={__}
            showLinkPopover={showLinkPopover}
            setShowLinkPopover={setShowLinkPopover}
            showColorPicker={showColorPicker}
            setShowColorPicker={setShowColorPicker}
            showHighlightPicker={showHighlightPicker}
            setShowHighlightPicker={setShowHighlightPicker}
            handleLinkSubmit={handleLinkSubmit}
            handleImageSubmit={handleImageSubmit}
            showImagePopover={showImagePopover}
            setShowImagePopover={setShowImagePopover}
          />
        </div>
      </div>

      {/* Editor */}
      <div className="flex-1 overflow-y-auto">
        <EditorContent editor={editor} className="h-full" />
      </div>
    </div>
  );
}
