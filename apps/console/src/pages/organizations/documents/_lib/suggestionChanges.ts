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

import type { SuggestionStyleName } from "@probo/ui/src/RichEditor/suggestionStyle";
import { graphql, readInlineData } from "relay-runtime";

import type {
  DocumentSuggestionBlockKind,
  DocumentSuggestionFormatAction,
  DocumentSuggestionFormatMark,
  DocumentSuggestionNodeAction,
  DocumentSuggestionNodeKind,
  suggestionChanges_suggestion$data,
  suggestionChanges_suggestion$key,
} from "#/__generated__/core/suggestionChanges_suggestion.graphql";

import type { BlockKind, BlockLabel, StructureNode, SuggestionChange } from "./reviewAnchors";

const suggestionChangesFragment = graphql`
  fragment suggestionChanges_suggestion on DocumentSuggestion @inline {
    changes {
      kind
      deleted
      inserted
      mark
      formatAction
      href
      previousHref
      before {
        kind
        level
      }
      after {
        kind
        level
      }
      node
      nodeAction
    }
    changeCount
  }
`;

export type RecordedSuggestionChange = NonNullable<suggestionChanges_suggestion$data["changes"]>[number];

type RecordedBlockLabel = NonNullable<RecordedSuggestionChange["before"]>;

// The changes recorded when a suggestion was decided. None are recorded while
// it is open.
export function recordedSuggestionChanges(suggestionKey: suggestionChanges_suggestion$key): {
  changes: SuggestionChange[];
  changeCount: number;
} {
  const suggestion = readInlineData(suggestionChangesFragment, suggestionKey);
  const changes = suggestionChanges(suggestion.changes ?? []);
  return { changes, changeCount: suggestion.changeCount ?? changes.length };
}

// A change the client cannot draw, such as one of a kind added to the API
// later, is left out.
export function suggestionChanges(changes: readonly RecordedSuggestionChange[]): SuggestionChange[] {
  return changes.flatMap((change) => {
    const mapped = suggestionChange(change);
    return mapped == null ? [] : [mapped];
  });
}

export function suggestionChange(change: RecordedSuggestionChange): SuggestionChange | null {
  switch (change.kind) {
    case "TEXT":
      return { kind: "text", deleted: change.deleted ?? "", inserted: change.inserted ?? "" };
    case "FORMAT": {
      const style = change.mark == null ? null : formatStyle(change.mark);
      const action = change.formatAction == null ? null : formatAction(change.formatAction);
      if (style == null || action == null) {
        return null;
      }
      return {
        kind: "format",
        style,
        action,
        href: change.href ?? null,
        previousHref: change.previousHref ?? null,
      };
    }
    case "BLOCK":
      if (change.before == null || change.after == null) {
        return null;
      }
      return { kind: "block", before: blockLabel(change.before), after: blockLabel(change.after) };
    case "NODE": {
      const action = change.nodeAction == null ? null : nodeAction(change.nodeAction);
      if (action == null) {
        return null;
      }
      return {
        kind: "node",
        node: change.node == null ? "other" : structureNode(change.node),
        action,
      };
    }
    default:
      return null;
  }
}

function formatStyle(mark: DocumentSuggestionFormatMark): SuggestionStyleName | null {
  switch (mark) {
    case "BOLD":
      return "bold";
    case "ITALIC":
      return "italic";
    case "UNDERLINE":
      return "underline";
    case "STRIKE":
      return "strike";
    case "CODE":
      return "code";
    case "LINK":
      return "link";
    default:
      return null;
  }
}

function formatAction(action: DocumentSuggestionFormatAction): "add" | "remove" | null {
  switch (action) {
    case "ADD":
      return "add";
    case "REMOVE":
      return "remove";
    default:
      return null;
  }
}

function nodeAction(action: DocumentSuggestionNodeAction): "insert" | "delete" | null {
  switch (action) {
    case "INSERT":
      return "insert";
    case "DELETE":
      return "delete";
    default:
      return null;
  }
}

function blockLabel(label: RecordedBlockLabel): BlockLabel {
  return { kind: blockKind(label.kind), level: label.level ?? null };
}

function blockKind(kind: DocumentSuggestionBlockKind): BlockKind {
  switch (kind) {
    case "PARAGRAPH":
      return "paragraph";
    case "HEADING":
      return "heading";
    case "QUOTE":
      return "quote";
    case "CODE":
      return "code";
    case "BULLET_LIST":
      return "bulletList";
    case "ORDERED_LIST":
      return "orderedList";
    case "TASK_LIST":
      return "taskList";
    default:
      return "other";
  }
}

function structureNode(node: DocumentSuggestionNodeKind): StructureNode {
  switch (node) {
    case "TABLE":
      return "table";
    case "ROW":
      return "row";
    case "COLUMN":
      return "column";
    case "IMAGE":
      return "image";
    case "DIVIDER":
      return "divider";
    case "FILE":
      return "file";
    case "LINE":
      return "line";
    default:
      return "other";
  }
}
