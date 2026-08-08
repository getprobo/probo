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
import {
  type DocHandle,
  type MappedMarkSpec,
  type MappedNodeSpec,
  type MappedSchemaSpec,
  pmDocFromSpans,
  pmNodeToSpans,
  SchemaAdapter,
  syncPlugin,
} from "@automerge/prosemirror";
import { Extension, type Extensions, getSchema } from "@tiptap/core";
import type { Mark, Schema } from "@tiptap/pm/model";

export type RichEditorAutomergeDocument = {
  body: string;
};

const textPath: Automerge.Prop[] = ["body"];

const supportedNodeNames = new Set([
  "doc",
  "paragraph",
  "text",
  "heading",
  "blockquote",
  "codeBlock",
  "horizontalRule",
  "hardBreak",
  "bulletList",
  "orderedList",
  "listItem",
  "table",
  "tableRow",
  "tableCell",
  "tableHeader",
]);

const supportedMarkNames = new Set([
  "bold",
  "italic",
  "strike",
  "underline",
  "code",
  "link",
]);

export function supportsRichEditorCollaboration(content: string): boolean {
  if (!content) return true;

  const document = JSON.parse(content) as {
    type?: string;
    marks?: Array<{ type?: string }>;
    content?: unknown[];
  };

  function supportsNode(value: unknown): boolean {
    if (!value || typeof value !== "object") return false;

    const node = value as {
      type?: string;
      marks?: Array<{ type?: string }>;
      content?: unknown[];
    };
    if (!node.type || !supportedNodeNames.has(node.type)) return false;
    if (node.marks?.some(mark => !mark.type || !supportedMarkNames.has(mark.type))) {
      return false;
    }
    return node.content?.every(supportsNode) ?? true;
  }

  return supportsNode(document);
}

export function createRichEditorAutomergeDocument(
  content: string,
  extensions: Extensions,
): Automerge.Doc<RichEditorAutomergeDocument> {
  const adapter = createSchemaAdapter(extensions);
  const documentJSON: Record<string, unknown> = content
    ? parseJSONObject(content)
    : { type: "doc", content: [{ type: "paragraph" }] };
  markTableStructure(documentJSON);
  const pmDocument = adapter.schema.nodeFromJSON(documentJSON);
  const spans = pmNodeToSpans(adapter, pmDocument);
  const document = Automerge.from<RichEditorAutomergeDocument>({ body: "" });

  return Automerge.change(document, (draft) => {
    Automerge.updateSpans(draft, textPath, spans, adapter.updateSpansConfig());
  });
}

export function richEditorAutomergeContent(
  handle: DocHandle<RichEditorAutomergeDocument>,
  extensions: Extensions,
): object {
  const adapter = createSchemaAdapter(extensions);
  const content: unknown = pmDocFromSpans(
    adapter,
    Automerge.spans(handle.doc(), textPath),
  ).toJSON();
  if (!isJSONObject(content)) {
    throw new Error("Automerge produced invalid ProseMirror content");
  }
  return content;
}

export function createRichEditorCollaborationExtension(
  handle: DocHandle<RichEditorAutomergeDocument>,
  extensions: Extensions,
): Extension {
  return Extension.create({
    name: "automergeCollaboration",
    priority: 1_000,

    addProseMirrorPlugins() {
      const adapter = createSchemaAdapter(extensions, this.editor.schema);
      return [
        syncPlugin({
          adapter,
          handle,
          path: textPath,
        }),
      ];
    },
  });
}

export function createSchemaAdapter(
  extensions: Extensions,
  targetSchema?: Schema,
): SchemaAdapter {
  const sourceSchema = getSchema(extensions);
  const nodes: Record<string, MappedNodeSpec> = {};
  const marks: Record<string, MappedMarkSpec> = {};

  sourceSchema.spec.nodes.forEach((name, spec) => {
    nodes[name] = {
      ...spec,
      automerge: automergeNodeMapping(name),
    };
  });
  sourceSchema.spec.marks.forEach((name, spec) => {
    marks[name] = {
      ...spec,
      automerge: automergeMarkMapping(name),
    };
  });

  const adapter = new SchemaAdapter({ nodes, marks } satisfies MappedSchemaSpec);
  if (!targetSchema) return adapter;

  adapter.schema = targetSchema;
  adapter.unknownBlock = targetSchema.nodes.automergeUnknownBlock;
  adapter.nodeMappings = adapter.nodeMappings.map(mapping => ({
    ...mapping,
    outer: mapping.outer ? targetSchema.nodes[mapping.outer.name] : null,
    content: targetSchema.nodes[mapping.content.name],
  }));
  adapter.markMappings = adapter.markMappings.map(mapping => ({
    ...mapping,
    prosemirrorMark: targetSchema.marks[mapping.prosemirrorMark.name],
  }));

  return adapter;
}

function automergeNodeMapping(name: string): MappedNodeSpec["automerge"] {
  switch (name) {
    case "paragraph":
      return { block: "paragraph" };
    case "heading":
      return {
        block: "heading",
        attrParsers: {
          fromAutomerge: block => ({
            level: readNumberAttribute(block.attrs, "level", 1),
          }),
          fromProsemirror: node => ({
            level: readNumberAttribute(node.attrs, "level", 1),
          }),
        },
      };
    case "blockquote":
      return { block: "blockquote" };
    case "codeBlock":
      return {
        block: "code-block",
        attrParsers: {
          fromAutomerge: block => ({
            language: readNullableStringAttribute(block.attrs, "language"),
          }),
          fromProsemirror: node => ({
            language: readNullableStringAttribute(node.attrs, "language"),
          }),
        },
      };
    case "horizontalRule":
      return { block: "horizontal-rule" };
    case "hardBreak":
      return { block: "hard-break", isEmbed: true };
    case "listItem":
      return {
        block: {
          within: {
            orderedList: "ordered-list-item",
            bulletList: "unordered-list-item",
          },
        },
      };
    case "table":
      return { block: "table" };
    case "tableRow":
      return { block: "table-row" };
    case "tableCell":
    case "tableHeader":
      return {
        block: name === "tableCell" ? "table-cell" : "table-header",
        attrParsers: {
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
        },
      };
    case "automergeUnknownBlock":
      return { unknownBlock: true };
    default:
      return undefined;
  }
}

function automergeMarkMapping(name: string): MappedMarkSpec["automerge"] {
  switch (name) {
    case "bold":
      return { markName: "strong" };
    case "italic":
      return { markName: "em" };
    case "strike":
    case "underline":
    case "code":
      return { markName: name };
    case "link":
      return {
        markName: "link",
        parsers: {
          fromAutomerge: (value) => {
            if (typeof value !== "string") return { href: "", title: null };
            const parsed = parseJSONObject(value);
            return {
              href: readStringAttribute(parsed, "href", ""),
              title: readNullableStringAttribute(parsed, "title"),
            };
          },
          fromProsemirror: (mark: Mark) => JSON.stringify({
            href: readStringAttribute(mark.attrs, "href", ""),
            title: readNullableStringAttribute(mark.attrs, "title"),
          }),
        },
      };
    default:
      return undefined;
  }
}

function parseJSONObject(value: string): Record<string, unknown> {
  const parsed: unknown = JSON.parse(value);
  if (!isJSONObject(parsed)) {
    throw new Error("expected a JSON object");
  }
  return parsed;
}

function isJSONObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function markTableStructure(node: Record<string, unknown>): void {
  const type = node.type;
  if (
    type === "horizontalRule"
    || type === "table"
    || type === "tableRow"
    || type === "tableCell"
    || type === "tableHeader"
  ) {
    const attrs = isJSONObject(node.attrs) ? node.attrs : {};
    attrs.isAmgBlock = true;
    node.attrs = attrs;
  }

  if (!Array.isArray(node.content)) return;
  for (const child of node.content) {
    if (isJSONObject(child)) markTableStructure(child);
  }
}

function readNumberAttribute(
  attributes: unknown,
  name: string,
  fallback: number,
): number {
  if (!isJSONObject(attributes)) return fallback;
  const value = attributes[name];
  return typeof value === "number" ? value : fallback;
}

function readStringAttribute(
  attributes: unknown,
  name: string,
  fallback: string,
): string {
  if (!isJSONObject(attributes)) return fallback;
  const value = attributes[name];
  return typeof value === "string" ? value : fallback;
}

function readNullableStringAttribute(
  attributes: unknown,
  name: string,
): string | null {
  if (!isJSONObject(attributes)) return null;
  const value = attributes[name];
  return typeof value === "string" ? value : null;
}

function readNumberArrayAttribute(
  attributes: unknown,
  name: string,
): number[] | null {
  if (!isJSONObject(attributes)) return null;
  const value = attributes[name];
  if (!Array.isArray(value)) return null;

  const numbers = value.filter(item => typeof item === "number");
  return numbers.length === value.length ? numbers : null;
}
