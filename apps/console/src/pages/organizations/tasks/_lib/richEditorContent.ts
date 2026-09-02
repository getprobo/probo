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

type JSONContent = {
  type?: string;
  text?: string;
  content?: JSONContent[];
};

function isJSONContent(value: unknown): value is JSONContent {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  if (!("content" in value) || value.content === undefined) {
    return true;
  }

  return Array.isArray(value.content) && value.content.every(isJSONContent);
}

function isDocContent(value: unknown): value is JSONContent {
  return isJSONContent(value) && value.type === "doc";
}

function jsonContentHasText(node: JSONContent): boolean {
  if (typeof node.text === "string" && node.text.trim().length > 0) {
    return true;
  }

  return node.content?.some(jsonContentHasText) ?? false;
}

function isEmptyBlock(node: JSONContent): boolean {
  switch (node.type) {
    case "horizontalRule":
    case "image":
      return false;
    case "hardBreak":
      return true;
    case "paragraph":
    case "heading":
    case "codeBlock":
      return !jsonContentHasText(node);
    case "blockquote":
    case "bulletList":
    case "orderedList":
    case "listItem":
    case "table":
    case "tableRow":
    case "tableHeader":
    case "tableCell":
      return (node.content ?? []).every(isEmptyBlock);
    default:
      return !jsonContentHasText(node);
  }
}

function isEmptyDoc(node: JSONContent): boolean {
  const content = node.content ?? [];
  if (content.length === 0) {
    return true;
  }

  return content.every(isEmptyBlock);
}

function parseDoc(content: string): JSONContent | null {
  if (!content.trim()) {
    return null;
  }

  try {
    const parsed: unknown = JSON.parse(content);
    if (isDocContent(parsed)) {
      return parsed;
    }
  } catch {
    return null;
  }

  return null;
}

export function isRichEditorContentEmpty(content: string): boolean {
  const parsed = parseDoc(content);
  if (parsed == null) {
    return true;
  }

  return isEmptyDoc(parsed);
}

function jsonContentTextLength(node: JSONContent): number {
  let length = typeof node.text === "string" ? [...node.text].length : 0;

  for (const child of node.content ?? []) {
    length += jsonContentTextLength(child);
  }

  return length;
}

export function richEditorContentTextLength(content: string): number {
  const parsed = parseDoc(content);
  if (parsed == null) {
    return 0;
  }

  return jsonContentTextLength(parsed);
}
