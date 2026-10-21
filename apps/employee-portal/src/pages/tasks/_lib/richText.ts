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
  text?: string;
  content?: JSONContent[];
};

export const taskCommentMaxLength = 5000;

function isJSONContent(value: unknown): value is JSONContent {
  return typeof value === "object" && value != null;
}

function parseDoc(content: string): JSONContent | null {
  if (!content.trim()) {
    return null;
  }

  try {
    const parsed: unknown = JSON.parse(content);
    if (isJSONContent(parsed)) {
      return parsed;
    }
  } catch {
    return null;
  }

  return null;
}

function textOf(node: JSONContent): string {
  let value = node.text ?? "";
  for (const child of node.content ?? []) {
    value += textOf(child);
  }
  return value;
}

export function isRichTextEmpty(content: string): boolean {
  const parsed = parseDoc(content);
  if (parsed == null) {
    return true;
  }

  return textOf(parsed).trim() === "";
}

export function richTextLength(content: string): number {
  const parsed = parseDoc(content);
  if (parsed == null) {
    return 0;
  }

  return [...textOf(parsed)].length;
}
