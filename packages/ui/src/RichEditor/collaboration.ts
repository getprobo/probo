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
} from "@automerge/prosemirror";
import { Extension, type Extensions, getSchema } from "@tiptap/core";
import type {
  Mark,
  Node as ProseMirrorNode,
  Schema,
} from "@tiptap/pm/model";
import { Plugin } from "@tiptap/pm/state";

import { createAutomergeSyncPlugin } from "./AutomergeSyncPlugin";

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
  "image",
  "file",
  "hardBreak",
  "bulletList",
  "orderedList",
  "listItem",
  "taskList",
  "taskItem",
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

  try {
    return supportsNode(JSON.parse(content));
  } catch {
    return false;
  }
}

export function createRichEditorAutomergeDocument(
  content: string,
  extensions: Extensions,
): Automerge.Doc<RichEditorAutomergeDocument> {
  const adapter = createSchemaAdapter(extensions);
  const documentJSON: Record<string, unknown> = content
    ? parseJSONObject(content)
    : { type: "doc", content: [{ type: "paragraph" }] };
  markAutomergeStructuralBlocks(documentJSON);
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

export function reconcileRichEditorAutomergeDocument(
  handle: DocHandle<RichEditorAutomergeDocument>,
  document: ProseMirrorNode,
  extensions: Extensions,
): void {
  let hasStructuralLeaf = false;
  document.descendants((node) => {
    if (
      node.type.name === "horizontalRule"
      || node.type.name === "image"
      || node.type.name === "file"
    ) {
      hasStructuralLeaf = true;
      return false;
    }

    return true;
  });
  if (!hasStructuralLeaf) return;

  const adapter = createSchemaAdapter(extensions, document.type.schema);
  const spans = pmNodeToSpans(adapter, document);
  handle.change((draft) => {
    Automerge.updateSpans(
      draft,
      textPath,
      spans,
      adapter.updateSpansConfig(),
    );
  });
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
        explicitBlockIdentityPlugin(),
        createAutomergeSyncPlugin(adapter, handle, textPath),
      ];
    },
  });
}

export function explicitBlockIdentityPlugin(): Plugin {
  return new Plugin({
    appendTransaction(transactions, _oldState, state) {
      if (!transactions.some(transaction => transaction.docChanged)) {
        return null;
      }

      const transaction = state.tr;
      let changed = false;
      state.doc.descendants((node, position, parent) => {
        if (node.attrs.isAmgBlock !== false) return true;

        const isTopLevelTextBlock = parent === state.doc && (
          node.type.name === "paragraph"
          || node.type.name === "heading"
          || node.type.name === "blockquote"
          || node.type.name === "codeBlock"
        );
        const isListItem = node.type.name === "listItem"
          || node.type.name === "taskItem";
        if (!isTopLevelTextBlock && !isListItem) {
          return true;
        }

        transaction.setNodeMarkup(
          position,
          undefined,
          {
            ...node.attrs,
            isAmgBlock: true,
          },
        );
        changed = true;

        return true;
      });
      if (!changed) return null;

      transaction.setMeta("addToHistory", false);

      return transaction;
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
  for (const mapping of adapter.nodeMappings) {
    if (mapping.blockName === "task-list-item") {
      mapping.outer = adapter.schema.nodes.taskList ?? null;
    }
  }
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
    case "image":
      return {
        block: "image",
        attrParsers: {
          fromAutomerge: block => imageAttrsFromBlock(block.attrs),
          fromProsemirror: node => imageAttrsFromBlock(node.attrs),
        },
      };
    case "file":
      return {
        block: "file",
        attrParsers: {
          fromAutomerge: block => fileAttrsFromBlock(block.attrs),
          fromProsemirror: node => fileAttrsFromBlock(node.attrs),
        },
      };
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
    case "taskItem":
      return {
        block: "task-list-item",
        attrParsers: {
          fromAutomerge: block => ({
            checked: readBooleanAttribute(block.attrs, "checked"),
          }),
          fromProsemirror: node => ({
            checked: readBooleanAttribute(node.attrs, "checked"),
          }),
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

export function markAutomergeStructuralBlocks(
  node: Record<string, unknown>,
): void {
  const type = node.type;
  if (
    type === "horizontalRule"
    || type === "image"
    || type === "file"
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
    if (isJSONObject(child)) markAutomergeStructuralBlocks(child);
  }
}

function imageAttrsFromBlock(attributes: unknown): Record<string, string | number> {
  const attrs = presentStrings(attributes, ["fileId", "src", "alt", "title"]);
  const width = readNumberAttribute(attributes, "width", Number.NaN);
  if (width > 0 && width <= 100) {
    attrs.width = width;
  }

  return attrs;
}

function fileAttrsFromBlock(attributes: unknown): Record<string, string | number> {
  const attrs = presentStrings(attributes, ["fileId", "fileName", "mimeType"]);
  const size = readNumberAttribute(attributes, "size", Number.NaN);
  if (Number.isFinite(size) && size >= 0) {
    attrs.size = size;
  }

  return attrs;
}

function presentStrings(
  attributes: unknown,
  names: string[],
): Record<string, string | number> {
  const attrs: Record<string, string | number> = {};
  for (const name of names) {
    const value = readStringAttribute(attributes, name, "");
    if (value.trim() !== "") {
      attrs[name] = value;
    }
  }

  return attrs;
}

function readBooleanAttribute(attributes: unknown, name: string): boolean {
  if (!isJSONObject(attributes)) return false;

  return attributes[name] === true;
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
