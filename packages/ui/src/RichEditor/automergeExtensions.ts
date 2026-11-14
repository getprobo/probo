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

import { Blockquote } from "@tiptap/extension-blockquote";
import { Bold } from "@tiptap/extension-bold";
import { Code } from "@tiptap/extension-code";
import { HardBreak } from "@tiptap/extension-hard-break";
import { Heading } from "@tiptap/extension-heading";
import { HorizontalRule } from "@tiptap/extension-horizontal-rule";
import { Italic } from "@tiptap/extension-italic";
import { ListItem, TaskItem } from "@tiptap/extension-list";
import { Paragraph } from "@tiptap/extension-paragraph";
import { Strike } from "@tiptap/extension-strike";
import { Table, TableCell, TableHeader, TableRow } from "@tiptap/extension-table";
import { Underline } from "@tiptap/extension-underline";

import {
  type AutomergeNodeMapping,
  readBooleanAttribute,
  readNumberArrayAttribute,
  readNumberAttribute,
} from "./automergeMapping";

export const ParagraphExtension = Paragraph.extend({
  automerge: () => ({ block: "paragraph", explicitIdentity: "topLevelBlock" }),
});

export const HeadingExtension = Heading.extend({
  automerge: () => ({
    block: "heading",
    explicitIdentity: "topLevelBlock",
    attrParsers: {
      fromAutomerge: block => ({
        level: readNumberAttribute(block.attrs, "level", 1),
      }),
      fromProsemirror: node => ({
        level: readNumberAttribute(node.attrs, "level", 1),
      }),
    },
  }),
});

export const BlockquoteExtension = Blockquote.extend({
  automerge: () => ({ block: "blockquote", explicitIdentity: "topLevelBlock" }),
});

export const HorizontalRuleExtension = HorizontalRule.extend({
  automerge: () => ({ block: "horizontal-rule", structural: true }),
});

export const HardBreakExtension = HardBreak.extend({
  automerge: () => ({ block: "hard-break", isEmbed: true }),
});

export const ListItemExtension = ListItem.extend({
  automerge: () => ({
    block: {
      within: {
        orderedList: "ordered-list-item",
        bulletList: "unordered-list-item",
      },
    },
    explicitIdentity: "listItem",
  }),
});

export const TaskItemExtension = TaskItem.extend({
  automerge: () => ({
    block: "task-list-item",
    explicitIdentity: "listItem",
    attrParsers: {
      fromAutomerge: block => ({
        checked: readBooleanAttribute(block.attrs, "checked"),
      }),
      fromProsemirror: node => ({
        checked: readBooleanAttribute(node.attrs, "checked"),
      }),
    },
  }),
}).configure({ nested: true });

export const TableExtension = Table.extend({
  automerge: () => ({ block: "table", structural: true }),
}).configure({ resizable: true });

export const TableRowExtension = TableRow.extend({
  automerge: () => ({ block: "table-row", structural: true }),
});

function tableCellAttrParsers(): AutomergeNodeMapping["attrParsers"] {
  return {
    fromAutomerge: block => ({
      colspan: readNumberAttribute(block.attrs, "colspan", 1),
      rowspan: readNumberAttribute(block.attrs, "rowspan", 1),
      colwidth: readNumberArrayAttribute(block.attrs, "colwidth"),
    }),
    fromProsemirror: node => ({
      colspan: readNumberAttribute(node.attrs, "colspan", 1),
      rowspan: readNumberAttribute(node.attrs, "rowspan", 1),
      colwidth: readNumberArrayAttribute(node.attrs, "colwidth"),
    }),
  };
}

export const TableCellExtension = TableCell.extend({
  automerge: () => ({
    block: "table-cell",
    structural: true,
    attrParsers: tableCellAttrParsers(),
  }),
});

export const TableHeaderExtension = TableHeader.extend({
  automerge: () => ({
    block: "table-header",
    structural: true,
    attrParsers: tableCellAttrParsers(),
  }),
});

export const BoldExtension = Bold.extend({
  automerge: () => ({ markName: "strong" }),
});

export const ItalicExtension = Italic.extend({
  automerge: () => ({ markName: "em" }),
});

export const StrikeExtension = Strike.extend({
  automerge: () => ({ markName: "strike" }),
});

export const UnderlineExtension = Underline.extend({
  automerge: () => ({ markName: "underline" }),
});

export const CodeExtension = Code.extend({
  // Code replaces the other character styles, and still shares its span with
  // suggestion marks so a suggested code change can be accepted or rejected.
  excludes: "bold italic strike underline link",
  automerge: () => ({ markName: "code" }),
});
