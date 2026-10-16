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

import * as Automerge from "@automerge/automerge";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";

const collaborationDebugStorageKey = "probo:collaboration-debug";
const collaborationDebugPrefix = "[probo:collaboration]";

export function collaborationDebug(
  event: string,
  details: Record<string, unknown>,
): void {
  if (!collaborationDebugEnabled()) return;
  console.info(collaborationDebugPrefix, event, JSON.stringify(details));
}

export function summarizeAutomergeSpans(
  document: Automerge.Doc<{ body: string }>,
): Array<Record<string, unknown>> {
  return Automerge.spans(document, ["body"]).map((span, index) => {
    if (span.type === "text") {
      return {
        index,
        kind: "text",
        length: span.value.length,
        marks: Object.keys(span.marks ?? {}).sort(),
      };
    }

    return {
      index,
      kind: "block",
      type: automergeString(span.value.type),
      parents: automergeStringArray(span.value.parents),
      isEmbed: span.value.isEmbed,
      attrs: Object.keys(
        isRecord(span.value.attrs) ? span.value.attrs : {},
      ).sort(),
    };
  });
}

export function summarizeProseMirrorDocument(
  document: ProseMirrorNode,
): Array<Record<string, unknown>> {
  const nodes: Array<Record<string, unknown>> = [];
  document.descendants((node, position, parent) => {
    nodes.push({
      position,
      type: node.type.name,
      parent: parent?.type.name ?? null,
      isText: node.isText,
      textLength: node.isText ? node.text?.length ?? 0 : 0,
      isAmgBlock: node.attrs.isAmgBlock ?? null,
    });
    return true;
  });
  return nodes;
}

function collaborationDebugEnabled(): boolean {
  return typeof window !== "undefined"
    && window.localStorage.getItem(collaborationDebugStorageKey) === "1";
}

function automergeString(value: unknown): string {
  if (typeof value === "string") return value;
  if (Automerge.isImmutableString(value)) return value.val;
  return "<non-string>";
}

function automergeStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.map(automergeString);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
