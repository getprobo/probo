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

import { Document } from "@tiptap/extension-document";
import { BulletList, ListKeymap, OrderedList, TaskList } from "@tiptap/extension-list";
import { Text } from "@tiptap/extension-text";
import { Dropcursor, UndoRedo } from "@tiptap/extensions";

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
import { BaseRichEditor, type RichEditorProps } from "./BaseRichEditor";
import { CodeBlockExtension } from "./CodeBlockExtension";
import { ContentUploadExtension } from "./ContentUploadExtension";
import { FileExtension } from "./FileExtension";
import { ImageExtension } from "./ImageExtension";
import { LinkExtension } from "./LinkExtension";
import { MarkdownPasteExtension } from "./MarkdownPasteExtension";
import { MentionExtension } from "./MentionExtension";
import { PlaceholderExtension } from "./PlaceholderExtension";
import { SlashCommandExtension } from "./SlashCommandExtension";
import { SuggestionDelete, SuggestionInsert } from "./SuggestionExtension";
import { TaskItemFocusExtension } from "./TaskItemFocusExtension";

export type { RichEditorProps };

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

export const richEditorCommentExtensions = [
  Document,
  ParagraphExtension,
  Text,
  BoldExtension,
  ItalicExtension,
  StrikeExtension,
  CodeExtension,
  LinkExtension,
  BlockquoteExtension,
  BulletList,
  OrderedList,
  ListItemExtension,
  ListKeymap,
  HardBreakExtension,
  UndoRedo,
  MentionExtension,
  PlaceholderExtension,
];

export function RichEditor(props: RichEditorProps) {
  return <BaseRichEditor {...props} extensions={richEditorExtensions} blockMenu />;
}
