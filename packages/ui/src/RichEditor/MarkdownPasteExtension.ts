// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { Extension } from "@tiptap/core";
import {
  Fragment,
  type Node as ProseMirrorNode,
  type Schema,
  Slice,
} from "@tiptap/pm/model";
import { Plugin, PluginKey } from "@tiptap/pm/state";

const markdownPasteKey = new PluginKey("markdownPaste");

const codeBlockOpenPattern = /^```(\w*)$/;
const tableSeparatorPattern
  = /^\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)*\|?\s*$/m;
const markdownLinePattern
  = /(?:^```\w*$|^#{1,6}\s|^>\s|^[-+*]\s|^\d+\.\s|^(?:---|___|\*\*\*)\s*$)/m;

function parseCells(line: string): string[] {
  return line
    .replace(/^\|/, "")
    .replace(/\|\s*$/, "")
    .split("|")
    .map(cell => cell.trim());
}

function applyMarks(
  schema: Schema,
  markNames: string[],
  inner: string,
): ProseMirrorNode[] {
  const marks = markNames.map(name => schema.marks[name].create());
  return parseInlineContent(schema, inner).map((node) => {
    let combined = node.marks;
    for (const mark of marks) {
      combined = mark.addToSet(combined);
    }
    return node.mark(combined);
  });
}

function parseInlineContent(
  schema: Schema,
  text: string,
): ProseMirrorNode[] {
  const nodes: ProseMirrorNode[] = [];
  const pattern
    = /`([^`]+)`|\*\*\*(.+?)\*\*\*|\*\*(.+?)\*\*|~~(.+?)~~|<u>(.+?)<\/u>|\[([^\]]+)\]\(([^)]+)\)|\*(.+?)\*/g;
  let lastIndex = 0;
  let match;

  while ((match = pattern.exec(text)) !== null) {
    if (match.index > lastIndex) {
      nodes.push(schema.text(text.slice(lastIndex, match.index)));
    }

    if (match[1] !== undefined) {
      nodes.push(schema.text(match[1], [schema.marks.code.create()]));
    } else if (match[2] !== undefined) {
      nodes.push(...applyMarks(schema, ["bold", "italic"], match[2]));
    } else if (match[3] !== undefined) {
      nodes.push(...applyMarks(schema, ["bold"], match[3]));
    } else if (match[4] !== undefined) {
      nodes.push(...applyMarks(schema, ["strike"], match[4]));
    } else if (match[5] !== undefined) {
      nodes.push(...applyMarks(schema, ["underline"], match[5]));
    } else if (match[6] !== undefined) {
      const linkMark = schema.marks.link.create({ href: match[7] });
      for (const node of parseInlineContent(schema, match[6])) {
        nodes.push(node.mark(linkMark.addToSet(node.marks)));
      }
    } else if (match[8] !== undefined) {
      nodes.push(...applyMarks(schema, ["italic"], match[8]));
    }

    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < text.length) {
    nodes.push(schema.text(text.slice(lastIndex)));
  }

  return nodes;
}

function buildTable(
  schema: Schema,
  headerLine: string,
  dataLines: string[],
): ProseMirrorNode {
  const headers = parseCells(headerLine);
  const columnCount = headers.length;

  const headerCells = headers.map((h) => {
    const content = h ? parseInlineContent(schema, h) : [];
    return schema.nodes.tableHeader.create(
      null,
      schema.nodes.paragraph.create(
        null,
        content.length > 0 ? content : undefined,
      ),
    );
  });

  const tableRows: ProseMirrorNode[] = [
    schema.nodes.tableRow.create(null, headerCells),
  ];

  for (const line of dataLines) {
    const cells = parseCells(line);
    const rowCells: ProseMirrorNode[] = [];
    for (let c = 0; c < columnCount; c++) {
      const value = cells[c]?.trim() ?? "";
      const content = value ? parseInlineContent(schema, value) : [];
      rowCells.push(
        schema.nodes.tableCell.create(
          null,
          schema.nodes.paragraph.create(
            null,
            content.length > 0 ? content : undefined,
          ),
        ),
      );
    }
    tableRows.push(schema.nodes.tableRow.create(null, rowCells));
  }

  return schema.nodes.table.create(null, tableRows);
}

function hasMarkdown(text: string): boolean {
  return markdownLinePattern.test(text) || tableSeparatorPattern.test(text);
}

function parseMarkdown(
  text: string,
  schema: Schema,
): ProseMirrorNode[] {
  const lines = text.split("\n");
  const nodes: ProseMirrorNode[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Code blocks
    const codeMatch = line.match(codeBlockOpenPattern);
    if (codeMatch) {
      const language = codeMatch[1] || null;
      const contentLines: string[] = [];
      i++;
      while (i < lines.length && !lines[i].startsWith("```")) {
        contentLines.push(lines[i]);
        i++;
      }
      i++;

      const codeContent = contentLines.join("\n").replace(/\n$/, "");
      nodes.push(
        schema.nodes.codeBlock.create(
          { language },
          codeContent ? schema.text(codeContent) : undefined,
        ),
      );
      continue;
    }

    // Tables
    if (
      line.includes("|")
      && i + 1 < lines.length
      && tableSeparatorPattern.test(lines[i + 1].trim())
    ) {
      const headerLine = line;
      i += 2;
      const dataLines: string[] = [];
      while (i < lines.length) {
        const row = lines[i].trim();
        if (!row || !row.includes("|")) break;
        dataLines.push(lines[i]);
        i++;
      }
      nodes.push(buildTable(schema, headerLine, dataLines));
      continue;
    }

    // Horizontal rules (before bullet lists since --- / *** could overlap)
    const trimmed = line.trim();
    if (trimmed === "---" || trimmed === "___" || trimmed === "***") {
      nodes.push(schema.nodes.horizontalRule.create());
      i++;
      continue;
    }

    // Headings
    const headingMatch = line.match(/^(#{1,6})\s(.+)$/);
    if (headingMatch) {
      const level = headingMatch[1].length;
      const content = parseInlineContent(schema, headingMatch[2]);
      nodes.push(
        schema.nodes.heading.create(
          { level },
          content.length > 0 ? content : undefined,
        ),
      );
      i++;
      continue;
    }

    // Blockquotes (group consecutive lines)
    if (line.startsWith("> ")) {
      const paragraphs: ProseMirrorNode[] = [];
      while (i < lines.length) {
        const bqMatch = lines[i].match(/^>\s(.*)$/);
        if (!bqMatch) break;
        const bqText = bqMatch[1];
        const content = bqText
          ? parseInlineContent(schema, bqText)
          : [];
        paragraphs.push(
          schema.nodes.paragraph.create(
            null,
            content.length > 0 ? content : undefined,
          ),
        );
        i++;
      }
      nodes.push(schema.nodes.blockquote.create(null, paragraphs));
      continue;
    }

    // Bullet lists (group consecutive items)
    if (
      line.startsWith("- ")
      || line.startsWith("+ ")
      || line.startsWith("* ")
    ) {
      const items: ProseMirrorNode[] = [];
      while (i < lines.length) {
        const blMatch = lines[i].match(/^[-+*]\s(.*)$/);
        if (!blMatch) break;
        const blText = blMatch[1];
        const content = blText
          ? parseInlineContent(schema, blText)
          : [];
        items.push(
          schema.nodes.listItem.create(
            null,
            schema.nodes.paragraph.create(
              null,
              content.length > 0 ? content : undefined,
            ),
          ),
        );
        i++;
      }
      nodes.push(schema.nodes.bulletList.create(null, items));
      continue;
    }

    // Ordered lists (group consecutive items)
    const olMatch = line.match(/^(\d+)\.\s(.*)$/);
    if (olMatch) {
      const start = +olMatch[1];
      const items: ProseMirrorNode[] = [];
      while (i < lines.length) {
        const itemMatch = lines[i].match(/^\d+\.\s(.*)$/);
        if (!itemMatch) break;
        const olText = itemMatch[1];
        const content = olText
          ? parseInlineContent(schema, olText)
          : [];
        items.push(
          schema.nodes.listItem.create(
            null,
            schema.nodes.paragraph.create(
              null,
              content.length > 0 ? content : undefined,
            ),
          ),
        );
        i++;
      }
      nodes.push(schema.nodes.orderedList.create({ start }, items));
      continue;
    }

    // Plain paragraph
    if (line.trim()) {
      const content = parseInlineContent(schema, line);
      nodes.push(
        schema.nodes.paragraph.create(
          null,
          content.length > 0 ? content : undefined,
        ),
      );
    }
    i++;
  }

  return nodes;
}

export const MarkdownPasteExtension = Extension.create({
  name: "markdownPaste",

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: markdownPasteKey,
        props: {
          handlePaste(view, event) {
            const text = event.clipboardData?.getData("text/plain") ?? "";
            if (!hasMarkdown(text)) {
              return false;
            }

            const schema = view.state.schema;
            const nodes = parseMarkdown(text, schema);
            if (nodes.length === 0) return false;

            const slice = new Slice(Fragment.from(nodes), 0, 0);
            const tr = view.state.tr.replaceSelection(slice);
            view.dispatch(tr);
            return true;
          },
        },
      }),
    ];
  },
});
