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

import type { Icon } from "@phosphor-icons/react";
import {
  CheckSquareIcon,
  CodeBlockIcon,
  GridFourIcon,
  ImageIcon,
  ListBulletsIcon,
  ListNumbersIcon,
  MinusIcon,
  PaperclipIcon,
  QuotesIcon,
  TextHFourIcon,
  TextHOneIcon,
  TextHThreeIcon,
  TextHTwoIcon,
  TextTIcon,
  TreeStructureIcon,
} from "@phosphor-icons/react";
import type { EditorState, Transaction } from "@tiptap/pm/state";
import type { Editor } from "@tiptap/react";

import {
  codeCommand,
  headingCommand,
  listCommand,
  paragraphCommand,
  suggestCommand,
  suggestDivider,
  wrapCommand,
} from "../suggestionBlock";
import { applySuggestingStructure } from "../SuggestionExtension";
import { suggestInsertTable } from "../suggestionStructure";

type ChainCommands = ReturnType<Editor["chain"]>;

export type BlockTypeId
  = | "paragraph"
    | "heading1"
    | "heading2"
    | "heading3"
    | "heading4"
    | "bulletList"
    | "orderedList"
    | "taskList"
    | "codeBlock"
    | "blockquote";

export type InsertId = "mermaid" | "divider" | "table" | "image" | "file";

type BlockCommand<Id extends string> = {
  id: Id;
  label: string;
  icon: Icon;
  action: (chain: ChainCommands) => ChainCommands;
  suggest?: (state: EditorState, id: string) => Transaction | null;
};

type BlockPicker = {
  id: InsertId;
  label: string;
  icon: Icon;
  pick: "image" | "file";
};

export type BlockTypeItem = BlockCommand<BlockTypeId>;
export type InsertItem = BlockCommand<InsertId> | BlockPicker;
export type BlockItem = BlockTypeItem | InsertItem;

export const blockTypeItems: BlockTypeItem[] = [
  { id: "paragraph", label: "Text", icon: TextTIcon, action: chain => chain.setParagraph(), suggest: (state, id) => suggestCommand(state, id, paragraphCommand()) },
  { id: "heading1", label: "Heading 1", icon: TextHOneIcon, action: chain => chain.toggleHeading({ level: 1 }), suggest: (state, id) => suggestCommand(state, id, headingCommand(1)) },
  { id: "heading2", label: "Heading 2", icon: TextHTwoIcon, action: chain => chain.toggleHeading({ level: 2 }), suggest: (state, id) => suggestCommand(state, id, headingCommand(2)) },
  { id: "heading3", label: "Heading 3", icon: TextHThreeIcon, action: chain => chain.toggleHeading({ level: 3 }), suggest: (state, id) => suggestCommand(state, id, headingCommand(3)) },
  { id: "heading4", label: "Heading 4", icon: TextHFourIcon, action: chain => chain.toggleHeading({ level: 4 }), suggest: (state, id) => suggestCommand(state, id, headingCommand(4)) },
  { id: "bulletList", label: "Bullet List", icon: ListBulletsIcon, action: chain => chain.toggleBulletList(), suggest: (state, id) => suggestCommand(state, id, listCommand("bulletList")) },
  { id: "orderedList", label: "Ordered List", icon: ListNumbersIcon, action: chain => chain.toggleOrderedList(), suggest: (state, id) => suggestCommand(state, id, listCommand("orderedList")) },
  { id: "taskList", label: "Task List", icon: CheckSquareIcon, action: chain => chain.toggleTaskList(), suggest: (state, id) => suggestCommand(state, id, listCommand("taskList")) },
  { id: "codeBlock", label: "Code Block", icon: CodeBlockIcon, action: chain => chain.toggleCodeBlock(), suggest: (state, id) => suggestCommand(state, id, codeCommand()) },
  { id: "blockquote", label: "Blockquote", icon: QuotesIcon, action: chain => chain.toggleBlockquote(), suggest: (state, id) => suggestCommand(state, id, wrapCommand("blockquote")) },
];

export const insertItems: InsertItem[] = [
  { id: "mermaid", label: "Mermaid Diagram", icon: TreeStructureIcon, action: chain => chain.setCodeBlock({ language: "mermaid" }), suggest: (state, id) => suggestCommand(state, id, codeCommand("mermaid")) },
  { id: "divider", label: "Divider", icon: MinusIcon, action: chain => chain.setHorizontalRule(), suggest: suggestDivider },
  { id: "table", label: "Table", icon: GridFourIcon, action: chain => chain.insertTable(), suggest: (state, id) => suggestInsertTable(state, id, 3, 3, true) },
  { id: "image", label: "Image", icon: ImageIcon, pick: "image" },
  { id: "file", label: "File", icon: PaperclipIcon, pick: "file" },
];

export const blockItems: BlockItem[] = [...blockTypeItems, ...insertItems];

export function isPickerItem(item: BlockItem): item is BlockPicker {
  return "pick" in item;
}

export function runBlockItem(
  editor: Editor,
  item: BlockItem,
  onPickFiles?: (kind: "image" | "file") => void,
) {
  try {
    if (isPickerItem(item)) {
      onPickFiles?.(item.pick);
    } else if (!item.suggest || !applySuggestingStructure(editor, item.suggest, { fresh: true })) {
      item.action(editor.chain().focus()).run();
    }
  } catch {
    // Block may no longer be in the document
  }
}
