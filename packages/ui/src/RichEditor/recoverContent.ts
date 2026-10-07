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

import type { JSONContent } from "@tiptap/react";

const maxDocumentDecodeDepth = 4;

const literalBlockTypes = new Set(["paragraph", "codeBlock"]);

const recoverableBlockTypes = new Set([
  "paragraph",
  "heading",
  "blockquote",
  "codeBlock",
  "horizontalRule",
  "bulletList",
  "orderedList",
  "table",
  "image",
]);

function isContentNode(value: unknown): value is JSONContent {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  if (!("content" in value) || value.content == null) {
    return true;
  }

  return Array.isArray(value.content) && value.content.every(isContentNode);
}

function parseDocValue(value: unknown, depth = 0): JSONContent | null {
  if (depth > maxDocumentDecodeDepth) {
    return null;
  }

  if (typeof value === "string") {
    const trimmed = value.trim();
    if (trimmed === "") {
      return null;
    }

    try {
      return parseDocValue(JSON.parse(trimmed) as unknown, depth + 1);
    } catch {
      return null;
    }
  }

  if (!isContentNode(value) || value.type !== "doc") {
    return null;
  }

  return value;
}

function inlineText(node: JSONContent): string {
  if (node.type === "text") {
    return node.text ?? "";
  }

  if (node.type === "hardBreak") {
    return "\n";
  }

  return (node.content ?? []).map(inlineText).join("");
}

function literalDocumentText(node: JSONContent): string | null {
  const blocks = node.content ?? [];
  if (blocks.length === 0) {
    return null;
  }

  const lines: string[] = [];
  for (const block of blocks) {
    if (block.type == null || !literalBlockTypes.has(block.type)) {
      return null;
    }

    lines.push(inlineText(block));
  }

  return lines.join("\n").trim();
}

function isRecoverableDoc(node: JSONContent): boolean {
  const blocks = node.content ?? [];
  if (blocks.length === 0) {
    return false;
  }

  return blocks.every(
    block => block.type != null && recoverableBlockTypes.has(block.type),
  );
}

function unwrapLiteralDocument(node: JSONContent, depth = 0): JSONContent {
  if (depth >= maxDocumentDecodeDepth) {
    return node;
  }

  const text = literalDocumentText(node);
  if (text == null || text === "" || (text[0] !== "{" && text[0] !== "\"")) {
    return node;
  }

  const inner = parseDocValue(text);
  if (inner == null || !isRecoverableDoc(inner)) {
    return node;
  }

  return unwrapLiteralDocument(inner, depth + 1);
}

// recoverRichContent parses editor JSON, including a document that was
// stored as a JSON string or as paragraphs of that JSON. Returns null when
// the value is not a ProseMirror document.
export function recoverRichContent(content: string): JSONContent | null {
  const parsed = parseDocValue(content);
  if (parsed == null) {
    return null;
  }

  return unwrapLiteralDocument(parsed);
}
