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
import { type ReactNode, useCallback, useRef, useState } from "react";
import { Converter as ShowdownConverter } from "showdown";
import TurndownService from "turndown";

type EditorMode = "visual" | "html" | "markdown";

const turndown = new TurndownService({
  headingStyle: "atx",
  codeBlockStyle: "fenced",
  bulletListMarker: "-",
});

turndown.addRule("taskListItems", {
  filter: (node) => {
    return (
      node.nodeName === "LI"
      && node.getAttribute("data-type") === "taskItem"
    );
  },
  replacement: (content, node) => {
    const checked = node.getAttribute("data-checked") === "true";
    return `${checked ? "- [x]" : "- [ ]"} ${content.trim()}\n`;
  },
});

const showdown = new ShowdownConverter({
  tables: true,
  tasklists: true,
  strikethrough: true,
  ghCodeBlocks: true,
  simpleLineBreaks: false,
});

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
  children: ReactNode;
}) {
  return (
    <button
      type="button"
      title={title}
      onMouseDown={e => e.preventDefault()}
      onClick={onClick}
      disabled={disabled}
      className={clsx(
        "size-8 flex items-center justify-center rounded-md transition-colors",
        disabled && "opacity-30 cursor-not-allowed",
        !disabled && "cursor-pointer",
        !disabled && active && "bg-primary/10 text-primary",
        !disabled && !active && "text-txt-secondary hover:bg-subtle hover:text-txt-primary",
      )}
    >
      {children}
    </button>
  );
}

function ToolbarSep() {
  return <div className="w-px h-5 bg-border-low/50 shrink-0" />;
}

function ToolbarGrp({ children }: { children: ReactNode }) {
  return (
    <div className="flex items-center gap-px rounded-lg bg-level-1 shadow-sm ring-1 ring-border-low/40 p-0.5">
      {children}
    </div>
  );
}

// ─── SVG icon helper ────────────────────────────────────────────────

function Ico({ children, size = 16 }: { children: ReactNode; size?: number }) {
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
  const presets = [
    "#000000", "#374151", "#6b7280", "#9ca3af",
    "#dc2626", "#ea580c", "#ca8a04", "#16a34a",
    "#0d9488", "#2563eb", "#4f46e5", "#9333ea",
    "#db2777", "#92400e", "#0369a1", "#7c3aed",
  ];

  return (
    <div className="absolute left-0 top-full mt-1 z-50 bg-level-1 border border-border-low rounded-xl shadow-mid p-2.5 w-52">
      <div className="text-xs font-medium text-txt-tertiary px-1 py-1 mb-1">
        {__("Text color")}
      </div>
      <div className="grid grid-cols-8 gap-1 mb-2">
        {presets.map(c => (
          <button
            key={c}
            type="button"
            onMouseDown={e => e.preventDefault()}
            onClick={() => {
              onSelect(c);
              onClose();
            }}
            className={clsx(
              "size-5.5 rounded-full transition-all hover:scale-110 hover:ring-2 hover:ring-primary/30",
              currentColor === c && "ring-2 ring-primary ring-offset-1 ring-offset-level-1",
            )}
            style={{ backgroundColor: c }}
          />
        ))}
      </div>
      <div className="flex items-center gap-2 pt-2 border-t border-border-low">
        <label className="flex items-center gap-2 flex-1 cursor-pointer group">
          <input
            type="color"
            value={currentColor || "#000000"}
            onMouseDown={e => e.stopPropagation()}
            onChange={(e) => {
              onSelect(e.target.value);
            }}
            className="size-6 rounded cursor-pointer border border-border-low p-0 bg-transparent [&::-webkit-color-swatch-wrapper]:p-0.5 [&::-webkit-color-swatch]:rounded"
          />
          <span className="text-xs text-txt-secondary group-hover:text-txt-primary">
            {__("Custom")}
          </span>
        </label>
        <button
          type="button"
          onMouseDown={e => e.preventDefault()}
          onClick={() => {
            onSelect("");
            onClose();
          }}
          className="text-xs text-txt-secondary hover:text-txt-primary px-2 py-1 hover:bg-subtle rounded-md transition-colors"
        >
          {__("Reset")}
        </button>
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
  const presets = [
    "#fef08a", "#bbf7d0", "#bfdbfe", "#e9d5ff",
    "#fce7f3", "#fed7aa", "#fecaca", "#e5e7eb",
    "#fde68a", "#a7f3d0", "#93c5fd", "#c4b5fd",
    "#f9a8d4", "#fdba74", "#fca5a5", "#d1d5db",
  ];

  return (
    <div className="absolute left-0 top-full mt-1 z-50 bg-level-1 border border-border-low rounded-xl shadow-mid p-2.5 w-52">
      <div className="text-xs font-medium text-txt-tertiary px-1 py-1 mb-1">
        {__("Highlight")}
      </div>
      <div className="grid grid-cols-8 gap-1 mb-2">
        {presets.map(c => (
          <button
            key={c}
            type="button"
            onMouseDown={e => e.preventDefault()}
            onClick={() => {
              onSelect(c);
              onClose();
            }}
            className="size-5.5 rounded transition-all hover:scale-110 hover:ring-2 hover:ring-primary/30"
            style={{ backgroundColor: c }}
          />
        ))}
      </div>
      <div className="flex items-center gap-2 pt-2 border-t border-border-low">
        <label className="flex items-center gap-2 flex-1 cursor-pointer group">
          <input
            type="color"
            defaultValue="#fef08a"
            onMouseDown={e => e.stopPropagation()}
            onChange={(e) => {
              onSelect(e.target.value);
            }}
            className="size-6 rounded cursor-pointer border border-border-low p-0 bg-transparent [&::-webkit-color-swatch-wrapper]:p-0.5 [&::-webkit-color-swatch]:rounded"
          />
          <span className="text-xs text-txt-secondary group-hover:text-txt-primary">
            {__("Custom")}
          </span>
        </label>
        <button
          type="button"
          onMouseDown={e => e.preventDefault()}
          onClick={() => {
            onSelect("");
            onClose();
          }}
          className="text-xs text-txt-secondary hover:text-txt-primary px-2 py-1 hover:bg-subtle rounded-md transition-colors"
        >
          {__("Remove")}
        </button>
      </div>
    </div>
  );
}

// ─── Markdown toolbar ───────────────────────────────────────────────

function insertMd(
  ref: React.RefObject<HTMLTextAreaElement | null>,
  source: string,
  setSource: (v: string) => void,
  onHtmlChange: ((v: string) => void) | undefined,
  before: string,
  after: string,
  placeholder: string,
) {
  const ta = ref.current;
  if (!ta) return;
  const start = ta.selectionStart;
  const end = ta.selectionEnd;
  const selected = source.slice(start, end);
  const insert = selected || placeholder;
  const next = source.slice(0, start) + before + insert + after + source.slice(end);
  setSource(next);
  onHtmlChange?.(showdown.makeHtml(next));

  requestAnimationFrame(() => {
    ta.focus();
    const cursorStart = start + before.length;
    const cursorEnd = cursorStart + insert.length;
    ta.setSelectionRange(cursorStart, cursorEnd);
  });
}

function insertMdLine(
  ref: React.RefObject<HTMLTextAreaElement | null>,
  source: string,
  setSource: (v: string) => void,
  onHtmlChange: ((v: string) => void) | undefined,
  prefix: string,
) {
  const ta = ref.current;
  if (!ta) return;
  const start = ta.selectionStart;
  const lineStart = source.lastIndexOf("\n", start - 1) + 1;
  const needsNewline = lineStart > 0 && source[lineStart - 1] !== "\n" ? "" : "";
  const next = source.slice(0, lineStart) + needsNewline + prefix + source.slice(lineStart);
  setSource(next);
  onHtmlChange?.(showdown.makeHtml(next));

  requestAnimationFrame(() => {
    ta.focus();
    ta.setSelectionRange(lineStart + prefix.length, lineStart + prefix.length);
  });
}

function insertMdBlock(
  ref: React.RefObject<HTMLTextAreaElement | null>,
  source: string,
  setSource: (v: string) => void,
  onHtmlChange: ((v: string) => void) | undefined,
  block: string,
) {
  const ta = ref.current;
  if (!ta) return;
  const pos = ta.selectionStart;
  const pre = pos > 0 && source[pos - 1] !== "\n" ? "\n" : "";
  const next = source.slice(0, pos) + pre + block + source.slice(pos);
  setSource(next);
  onHtmlChange?.(showdown.makeHtml(next));

  requestAnimationFrame(() => {
    ta.focus();
    const newPos = pos + pre.length + block.length;
    ta.setSelectionRange(newPos, newPos);
  });
}

function MarkdownToolbar({
  __: t,
  textareaRef,
  source,
  setSource,
  onHtmlChange,
}: {
  __: (s: string) => string;
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  source: string;
  setSource: (v: string) => void;
  onHtmlChange: ((v: string) => void) | undefined;
}) {
  const [showMdColorPicker, setShowMdColorPicker] = useState(false);
  const [showMdHighlightPicker, setShowMdHighlightPicker] = useState(false);

  const wrap = (before: string, after: string, ph: string) =>
    insertMd(textareaRef, source, setSource, onHtmlChange, before, after, ph);
  const line = (prefix: string) =>
    insertMdLine(textareaRef, source, setSource, onHtmlChange, prefix);
  const block = (text: string) =>
    insertMdBlock(textareaRef, source, setSource, onHtmlChange, text);

  const execCmd = (cmd: string) => {
    const ta = textareaRef.current;
    if (!ta) return;
    ta.focus();
    document.execCommand(cmd);
  };

  const clearFormatting = () => {
    const ta = textareaRef.current;
    if (!ta) return;
    const start = ta.selectionStart;
    const end = ta.selectionEnd;
    if (start === end) return;
    const selected = source.slice(start, end);
    const cleaned = selected
      .replace(/(\*{1,3}|_{1,3}|~~|`{1,3})/g, "")
      .replace(/<\/?(u|sup|sub|b|i|s|em|strong|mark|span)[^>]*>/gi, "");
    const next = source.slice(0, start) + cleaned + source.slice(end);
    setSource(next);
    onHtmlChange?.(showdown.makeHtml(next));
    requestAnimationFrame(() => {
      ta.focus();
      ta.setSelectionRange(start, start + cleaned.length);
    });
  };

  return (
    <div className="flex flex-col overflow-visible">
      <div className="flex items-center gap-1.5 flex-wrap px-3 py-2">
        {/* Undo / Redo */}
        <ToolbarGrp>
          <ToolbarButton onClick={() => execCmd("undo")} title={t("Undo")}>
            <Ico>
              <path d="M3 7v6h6" />
              <path d="M21 17a9 9 0 0 0-9-9 9 9 0 0 0-6 2.3L3 13" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => execCmd("redo")} title={t("Redo")}>
            <Ico>
              <path d="M21 7v6h-6" />
              <path d="M3 17a9 9 0 0 1 9-9 9 9 0 0 1 6 2.3L21 13" />
            </Ico>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Headings */}
        <ToolbarGrp>
          <ToolbarButton onClick={() => line("# ")} title={t("Heading 1")}>
            <span className="text-xs font-bold leading-none">H1</span>
          </ToolbarButton>
          <ToolbarButton onClick={() => line("## ")} title={t("Heading 2")}>
            <span className="text-xs font-bold leading-none">H2</span>
          </ToolbarButton>
          <ToolbarButton onClick={() => line("### ")} title={t("Heading 3")}>
            <span className="text-xs font-bold leading-none">H3</span>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Text formatting */}
        <ToolbarGrp>
          <ToolbarButton onClick={() => wrap("**", "**", "bold")} title={t("Bold")}>
            <Ico><path d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6zM6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z" /></Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("*", "*", "italic")} title={t("Italic")}>
            <Ico>
              <path d="M19 4h-9" />
              <path d="M14 20H5" />
              <path d="M15 4L9 20" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("<u>", "</u>", "underline")} title={t("Underline")}>
            <Ico>
              <path d="M6 3v7a6 6 0 0 0 6 6 6 6 0 0 0 6-6V3" />
              <path d="M4 21h16" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("~~", "~~", "strikethrough")} title={t("Strikethrough")}>
            <Ico>
              <path d="M16 4H9a3 3 0 0 0-2.83 4" />
              <path d="M14 12a4 4 0 0 1 0 8H6" />
              <path d="M4 12h16" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("`", "`", "code")} title={t("Inline code")}>
            <Ico>
              <path d="m16 18 6-6-6-6" />
              <path d="m8 6-6 6 6 6" />
            </Ico>
          </ToolbarButton>
          <ToolbarSep />
          <ToolbarButton onClick={() => wrap("<sup>", "</sup>", "text")} title={t("Superscript")}>
            <Ico>
              <path d="m4 19 8-8" />
              <path d="m12 19-8-8" />
              <path d="M20 12h-4c0-1.5.44-2 1.5-2.5S20 8.33 20 7c0-1.06-.75-2-2-2a1.98 1.98 0 0 0-1.62.86" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("<sub>", "</sub>", "text")} title={t("Subscript")}>
            <Ico>
              <path d="m4 5 8 8" />
              <path d="m12 5-8 8" />
              <path d="M20 19h-4c0-1.5.44-2 1.5-2.5S20 15.33 20 14c0-1.06-.75-2-2-2a1.98 1.98 0 0 0-1.62.86" />
            </Ico>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Colors */}
        <ToolbarGrp>
          <div className="relative">
            <ToolbarButton
              onClick={() => {
                setShowMdColorPicker(!showMdColorPicker);
                setShowMdHighlightPicker(false);
              }}
              title={t("Text color")}
            >
              <div className="flex flex-col items-center gap-0.5">
                <svg
                  width={14}
                  height={14}
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M6 20h12" />
                  <path d="M12 4 6.5 16h2.1L10 13h4l1.4 3h2.1z" />
                </svg>
                <div className="h-1 w-4 rounded-sm bg-current" />
              </div>
            </ToolbarButton>
            {showMdColorPicker && (
              <ColorPicker
                currentColor=""
                onSelect={(c) => {
                  if (c) {
                    wrap(`<span style="color:${c}">`, "</span>", "text");
                  }
                  setShowMdColorPicker(false);
                }}
                onClose={() => setShowMdColorPicker(false)}
              />
            )}
          </div>
          <div className="relative">
            <ToolbarButton
              onClick={() => {
                setShowMdHighlightPicker(!showMdHighlightPicker);
                setShowMdColorPicker(false);
              }}
              title={t("Highlight")}
            >
              <svg
                width={16}
                height={16}
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="m19 4-8.8 8.8a1 1 0 0 0-.2.4L9 17l3.8-1a1 1 0 0 0 .4-.2L22 7" />
                <path d="m19 4 2 2-1.5 1.5" />
                <rect x="2" y="20" width="20" height="3" rx="1" fill="#fef08a" stroke="none" />
              </svg>
            </ToolbarButton>
            {showMdHighlightPicker && (
              <HighlightColorPicker
                onSelect={(c) => {
                  if (c) {
                    wrap(`<mark style="background-color:${c}">`, "</mark>", "text");
                  }
                  setShowMdHighlightPicker(false);
                }}
                onClose={() => setShowMdHighlightPicker(false)}
              />
            )}
          </div>
        </ToolbarGrp>

        {/* Alignment */}
        <ToolbarGrp>
          <ToolbarButton onClick={() => wrap("<div style=\"text-align:left\">\n", "\n</div>", "text")} title={t("Align left")}>
            <Ico>
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="15" y2="12" />
              <line x1="3" y1="18" x2="18" y2="18" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("<div style=\"text-align:center\">\n", "\n</div>", "text")} title={t("Align center")}>
            <Ico>
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="6" y1="12" x2="18" y2="12" />
              <line x1="4" y1="18" x2="20" y2="18" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("<div style=\"text-align:right\">\n", "\n</div>", "text")} title={t("Align right")}>
            <Ico>
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="9" y1="12" x2="21" y2="12" />
              <line x1="6" y1="18" x2="21" y2="18" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("<div style=\"text-align:justify\">\n", "\n</div>", "text")} title={t("Justify")}>
            <Ico>
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
            </Ico>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Lists */}
        <ToolbarGrp>
          <ToolbarButton onClick={() => line("- ")} title={t("Bullet list")}>
            <Ico>
              <line x1="9" y1="6" x2="20" y2="6" />
              <line x1="9" y1="12" x2="20" y2="12" />
              <line x1="9" y1="18" x2="20" y2="18" />
              <circle cx="4.5" cy="6" r="1.5" fill="currentColor" stroke="none" />
              <circle cx="4.5" cy="12" r="1.5" fill="currentColor" stroke="none" />
              <circle cx="4.5" cy="18" r="1.5" fill="currentColor" stroke="none" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => line("1. ")} title={t("Numbered list")}>
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
          <ToolbarButton onClick={() => line("- [ ] ")} title={t("Task list")}>
            <Ico>
              <rect x="3" y="5" width="6" height="6" rx="1" />
              <path d="m3 17 2 2 4-4" />
              <line x1="13" y1="6" x2="21" y2="6" />
              <line x1="13" y1="12" x2="21" y2="12" />
              <line x1="13" y1="18" x2="21" y2="18" />
            </Ico>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Blocks */}
        <ToolbarGrp>
          <ToolbarButton onClick={() => line("> ")} title={t("Blockquote")}>
            <Ico>
              <path d="M3 21c3 0 7-1 7-8V5c0-1.25-.756-2.017-2-2H4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2 1 0 1 0 1 1v1c0 1-1 2-2 2s-1 .008-1 1.031V21z" />
              <path d="M15 21c3 0 7-1 7-8V5c0-1.25-.757-2.017-2-2h-4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2h.75c0 2.25.25 4-2.75 4v3z" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("```\n", "\n```", "code")} title={t("Code block")}>
            <Ico>
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <path d="m10 10-2 2 2 2" />
              <path d="m14 10 2 2-2 2" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => block("\n---\n")} title={t("Horizontal rule")}>
            <Ico><path d="M3 12h18" /></Ico>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Insert */}
        <ToolbarGrp>
          <ToolbarButton onClick={() => wrap("[", "](url)", "link text")} title={t("Add link")}>
            <Ico>
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
            </Ico>
          </ToolbarButton>
          <ToolbarButton onClick={() => wrap("![alt](", ")", "image-url")} title={t("Insert image")}>
            <Ico>
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <path d="m21 15-5-5L5 21" />
            </Ico>
          </ToolbarButton>
          <ToolbarSep />
          <ToolbarButton
            onClick={() => block("\n| Header | Header | Header |\n| ------ | ------ | ------ |\n| Cell   | Cell   | Cell   |\n| Cell   | Cell   | Cell   |\n")}
            title={t("Insert table")}
          >
            <Ico>
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <line x1="3" y1="9" x2="21" y2="9" />
              <line x1="3" y1="15" x2="21" y2="15" />
              <line x1="9" y1="3" x2="9" y2="21" />
              <line x1="15" y1="3" x2="15" y2="21" />
            </Ico>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Clear */}
        <ToolbarGrp>
          <ToolbarButton onClick={clearFormatting} title={t("Clear formatting")}>
            <Ico>
              <path d="M4 7V4h16v3" />
              <path d="M9 20h6" />
              <path d="M12 4v16" />
              <path d="m17 3-10 18" />
            </Ico>
          </ToolbarButton>
        </ToolbarGrp>
      </div>
    </div>
  );
}

// ─── Toolbar rows ───────────────────────────────────────────────────

function Toolbar({
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
    <div className="flex flex-col overflow-visible">
      {/* Main toolbar */}
      <div className="flex items-center gap-1.5 flex-wrap px-3 py-2">
        {/* Undo / Redo */}
        <ToolbarGrp>
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
        </ToolbarGrp>

        {/* Headings */}
        <ToolbarGrp>
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
          <ToolbarSep />
          <ToolbarButton onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} active={editor.isActive("heading", { level: 1 })} title={t("Heading 1")}>
            <span className="text-xs font-bold leading-none">H1</span>
          </ToolbarButton>
          <ToolbarButton onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} active={editor.isActive("heading", { level: 2 })} title={t("Heading 2")}>
            <span className="text-xs font-bold leading-none">H2</span>
          </ToolbarButton>
          <ToolbarButton onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()} active={editor.isActive("heading", { level: 3 })} title={t("Heading 3")}>
            <span className="text-xs font-bold leading-none">H3</span>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Text formatting */}
        <ToolbarGrp>
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
          <ToolbarSep />
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
        </ToolbarGrp>

        {/* Colors */}
        <ToolbarGrp>
          <div className="relative">
            <ToolbarButton
              onClick={() => {
                setShowColorPicker(!showColorPicker);
                setShowHighlightPicker(false);
                setShowLinkPopover(false);
              }}
              title={t("Text color")}
            >
              <div className="flex flex-col items-center gap-0.5">
                <svg
                  width={14}
                  height={14}
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M6 20h12" />
                  <path d="M12 4 6.5 16h2.1L10 13h4l1.4 3h2.1z" />
                </svg>
                <div
                  className="h-1 w-4 rounded-sm"
                  style={{ backgroundColor: String(editor.getAttributes("textStyle").color || "currentColor") }}
                />
              </div>
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
              <svg
                width={16}
                height={16}
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="m19 4-8.8 8.8a1 1 0 0 0-.2.4L9 17l3.8-1a1 1 0 0 0 .4-.2L22 7" />
                <path d="m19 4 2 2-1.5 1.5" />
                <rect x="2" y="20" width="20" height="3" rx="1" fill="#fef08a" stroke="none" />
              </svg>
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
        </ToolbarGrp>

        {/* Lists */}
        <ToolbarGrp>
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
        </ToolbarGrp>

        {/* Alignment */}
        <ToolbarGrp>
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
        </ToolbarGrp>

        {/* Blocks */}
        <ToolbarGrp>
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
        </ToolbarGrp>

        {/* Insert */}
        <ToolbarGrp>
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
          <div className="relative">
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
          </div>
          <ToolbarSep />
          <ToolbarButton onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()} title={t("Insert table")}>
            <Ico>
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <line x1="3" y1="9" x2="21" y2="9" />
              <line x1="3" y1="15" x2="21" y2="15" />
              <line x1="9" y1="3" x2="9" y2="21" />
              <line x1="15" y1="3" x2="15" y2="21" />
            </Ico>
          </ToolbarButton>
        </ToolbarGrp>

        {/* Clear */}
        <ToolbarGrp>
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
        </ToolbarGrp>
      </div>

      {/* Table context bar */}
      {editor.isActive("table") && (
        <div className="flex items-center gap-1.5 flex-wrap px-3 py-1.5 border-t border-t-border-low/50 bg-primary/[0.03]">
          <span className="text-[10px] font-medium text-primary select-none">
            {t("Table")}
          </span>
          <ToolbarGrp>
            <ToolbarButton onClick={() => editor.chain().focus().addColumnAfter().run()} title={t("Add column")}>
              <Ico>
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <line x1="12" y1="8" x2="12" y2="16" />
                <line x1="8" y1="12" x2="16" y2="12" />
              </Ico>
            </ToolbarButton>
            <ToolbarButton onClick={() => editor.chain().focus().addRowAfter().run()} title={t("Add row")}>
              <Ico>
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <line x1="3" y1="12" x2="21" y2="12" />
                <line x1="12" y1="12" x2="12" y2="21" />
              </Ico>
            </ToolbarButton>
          </ToolbarGrp>
          <ToolbarGrp>
            <ToolbarButton onClick={() => editor.chain().focus().deleteColumn().run()} title={t("Delete column")}>
              <Ico>
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <line x1="9" y1="3" x2="9" y2="21" />
                <path d="m14 9 4 4" />
                <path d="m18 9-4 4" />
              </Ico>
            </ToolbarButton>
            <ToolbarButton onClick={() => editor.chain().focus().deleteRow().run()} title={t("Delete row")}>
              <Ico>
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <line x1="3" y1="9" x2="21" y2="9" />
                <path d="m9 14 4 4" />
                <path d="m13 14-4 4" />
              </Ico>
            </ToolbarButton>
          </ToolbarGrp>
          <ToolbarGrp>
            <ToolbarButton onClick={() => editor.chain().focus().deleteTable().run()} title={t("Delete table")}>
              <Ico>
                <path d="M3 6h18" />
                <path d="M8 6V4h8v2" />
                <path d="m19 6-.867 12.142A2 2 0 0 1 16.138 20H7.862a2 2 0 0 1-1.995-1.858L5 6" />
              </Ico>
            </ToolbarButton>
          </ToolbarGrp>
        </div>
      )}
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
  const [mode, setMode] = useState<EditorMode>("visual");
  const [htmlSource, setHtmlSource] = useState(value ?? "");
  const [markdownSource, setMarkdownSource] = useState("");
  const mdTextareaRef = useRef<HTMLTextAreaElement>(null);

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
        autolink: true,
        HTMLAttributes: {
          class: "text-primary underline cursor-pointer",
          target: "_blank",
          rel: "noopener noreferrer",
        },
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
    let href = url.trim();
    if (href && !/^https?:\/\//i.test(href)) {
      href = `https://${href}`;
    }
    editor.chain().focus().extendMarkRange("link").setLink({ href }).run();
  }, [editor]);

  const handleImageSubmit = useCallback((url: string) => {
    if (!editor) return;
    editor.chain().focus().setImage({ src: url }).run();
  }, [editor]);

  if (!editor) {
    return null;
  }

  const switchMode = (target: EditorMode) => {
    if (target === mode) return;

    // Collect current HTML from whatever mode we're in
    let currentHtml: string;
    if (mode === "visual") {
      currentHtml = editor.getHTML();
    } else if (mode === "html") {
      currentHtml = htmlSource;
    } else {
      currentHtml = showdown.makeHtml(markdownSource);
    }

    // Push into the target mode
    if (target === "visual") {
      editor.commands.setContent(currentHtml);
      onChange?.(currentHtml);
    } else if (target === "html") {
      setHtmlSource(currentHtml);
    } else {
      setMarkdownSource(turndown.turndown(currentHtml));
    }

    setMode(target);
  };

  const handleHtmlSourceChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setHtmlSource(e.target.value);
    onChange?.(e.target.value);
  };

  const handleMarkdownSourceChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const md = e.target.value;
    setMarkdownSource(md);
    onChange?.(showdown.makeHtml(md));
  };

  return (
    <div className="flex flex-col h-full">
      {/* Mode toggle bar */}
      <div className="shrink-0 flex items-center justify-between border-b border-b-border-low bg-level-2/50 px-3 py-1">
        <div className="flex items-center gap-1 rounded-lg bg-level-1 p-0.5 shadow-sm ring-1 ring-border-low/40">
          <button
            type="button"
            onClick={() => switchMode("visual")}
            className={clsx(
              "flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer",
              mode === "visual"
                ? "bg-primary/10 text-primary"
                : "text-txt-secondary hover:text-txt-primary hover:bg-subtle",
            )}
          >
            <Ico size={12}>
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
              <circle cx="12" cy="12" r="3" />
            </Ico>
            {__("Visual")}
          </button>
          <button
            type="button"
            onClick={() => switchMode("markdown")}
            className={clsx(
              "flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer",
              mode === "markdown"
                ? "bg-primary/10 text-primary"
                : "text-txt-secondary hover:text-txt-primary hover:bg-subtle",
            )}
          >
            <Ico size={12}>
              <path d="M2 16V8l4 4 4-4v8" />
              <path d="M18 8v8" />
              <path d="M22 12l-4 4-4-4" />
            </Ico>
            {__("Markdown")}
          </button>
          <button
            type="button"
            onClick={() => switchMode("html")}
            className={clsx(
              "flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer",
              mode === "html"
                ? "bg-primary/10 text-primary"
                : "text-txt-secondary hover:text-txt-primary hover:bg-subtle",
            )}
          >
            <Ico size={12}>
              <polyline points="16 18 22 12 16 6" />
              <polyline points="8 6 2 12 8 18" />
            </Ico>
            {__("HTML")}
          </button>
        </div>
      </div>

      {mode === "visual" && (
        <>
          {/* Toolbar */}
          <div className="relative shrink-0 border-b border-b-border-low bg-level-2/50 overflow-visible">
            <Toolbar
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

          {/* Editor */}
          <div className="flex-1 overflow-y-auto">
            <EditorContent editor={editor} className="h-full" />
          </div>
        </>
      )}
      {mode === "markdown" && (
        <>
          <div className="relative shrink-0 border-b border-b-border-low bg-level-2/50 overflow-visible">
            <MarkdownToolbar
              __={__}
              textareaRef={mdTextareaRef}
              source={markdownSource}
              setSource={setMarkdownSource}
              onHtmlChange={onChange}
            />
          </div>
          <div className="flex-1 overflow-hidden">
            <textarea
              ref={mdTextareaRef}
              className="w-full h-full resize-none bg-level-1 text-txt-primary font-mono text-sm px-6 py-4 focus:outline-none"
              value={markdownSource}
              onChange={handleMarkdownSourceChange}
              spellCheck={false}
              placeholder={placeholder ?? __("Write markdown here...")}
            />
          </div>
        </>
      )}
      {mode === "html" && (
        <div className="flex-1 overflow-hidden">
          <textarea
            className="w-full h-full resize-none bg-level-1 text-txt-primary font-mono text-sm px-6 py-4 focus:outline-none"
            value={htmlSource}
            onChange={handleHtmlSourceChange}
            spellCheck={false}
            placeholder={placeholder}
          />
        </div>
      )}
    </div>
  );
}
