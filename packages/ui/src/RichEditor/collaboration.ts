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
  Node as ProseMirrorNode,
  Schema,
} from "@tiptap/pm/model";
import { Plugin } from "@tiptap/pm/state";

import {
  isJSONObject,
  parseJSONObject,
  readAutomergeNodeMapping,
} from "./automergeMapping";
import { createAutomergeSyncPlugin } from "./AutomergeSyncPlugin";
import { richEditorCollaborationExtensions } from "./RichEditor";

export type RichEditorAutomergeDocument = {
  body: string;
};

const textPath: Automerge.Prop[] = ["body"];

export function supportsRichEditorCollaboration(content: string): boolean {
  if (!content) return true;

  const { nodes, marks } = collaborationSupportedNames();

  function supportsNode(value: unknown): boolean {
    if (!value || typeof value !== "object") return false;

    const node = value as {
      type?: string;
      marks?: Array<{ type?: string }>;
      content?: unknown[];
    };
    if (!node.type || !nodes.has(node.type)) return false;
    if (node.marks?.some(mark => !mark.type || !marks.has(mark.type))) {
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
    if (readAutomergeNodeMapping(node.type.spec)?.structural && node.type.isLeaf) {
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

        const identity = readAutomergeNodeMapping(node.type.spec)?.explicitIdentity;
        if (!identity) return true;
        if (identity === "topLevelBlock" && parent !== state.doc) return true;

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

type AdapterCacheEntry = {
  base?: SchemaAdapter;
  byTarget: WeakMap<Schema, SchemaAdapter>;
};

const adapterCache = new WeakMap<Extensions, AdapterCacheEntry>();

export function createSchemaAdapter(
  extensions: Extensions,
  targetSchema?: Schema,
): SchemaAdapter {
  let entry = adapterCache.get(extensions);
  if (!entry) {
    entry = { byTarget: new WeakMap() };
    adapterCache.set(extensions, entry);
  }

  if (!targetSchema) {
    entry.base ??= buildSchemaAdapter(extensions);
    return entry.base;
  }

  const cached = entry.byTarget.get(targetSchema);
  if (cached) return cached;

  const adapter = buildSchemaAdapter(extensions, targetSchema);
  entry.byTarget.set(targetSchema, adapter);
  return adapter;
}

function buildSchemaAdapter(
  extensions: Extensions,
  targetSchema?: Schema,
): SchemaAdapter {
  const sourceSchema = getSchema(extensions);
  const nodes: Record<string, MappedNodeSpec> = {};
  const marks: Record<string, MappedMarkSpec> = {};

  sourceSchema.spec.nodes.forEach((name, spec) => {
    nodes[name] = { ...spec };
  });
  sourceSchema.spec.marks.forEach((name, spec) => {
    marks[name] = { ...spec };
  });

  const adapter = new SchemaAdapter({ nodes, marks } satisfies MappedSchemaSpec);
  for (const mapping of adapter.nodeMappings) {
    if (mapping.blockName === "task-list-item") {
      mapping.outer = adapter.schema.nodes.taskList ?? null;
    }
  }
  // Read-only aliases for documents that still carry camelCase Automerge mark
  // names from an earlier sync path. New writes always use kebab-case via the
  // suggestion marks' automerge mappings; do not add further write aliases here.
  const legacySuggestionMarks: Record<string, string> = {
    "suggestion-insert": "suggestionInsert",
    "suggestion-delete": "suggestionDelete",
  };
  for (const mapping of [...adapter.markMappings]) {
    const legacy = legacySuggestionMarks[mapping.automergeMarkName];
    if (!legacy) continue;
    adapter.markMappings.push({ ...mapping, automergeMarkName: legacy });
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

export function markAutomergeStructuralBlocks(
  node: Record<string, unknown>,
): void {
  markStructural(node, structuralBlockNames());
}

function markStructural(
  node: Record<string, unknown>,
  structural: ReadonlySet<string>,
): void {
  const type = node.type;
  if (typeof type === "string" && structural.has(type)) {
    const attrs = isJSONObject(node.attrs) ? node.attrs : {};
    attrs.isAmgBlock = true;
    node.attrs = attrs;
  }

  if (!Array.isArray(node.content)) return;
  for (const child of node.content) {
    if (isJSONObject(child)) markStructural(child, structural);
  }
}

let structuralNames: Set<string> | null = null;

function structuralBlockNames(): ReadonlySet<string> {
  if (structuralNames) return structuralNames;

  const names = new Set<string>();
  getSchema(richEditorCollaborationExtensions).spec.nodes.forEach((name, spec) => {
    if (readAutomergeNodeMapping(spec)?.structural) names.add(name);
  });
  structuralNames = names;

  return names;
}

let supportedNames: { nodes: Set<string>; marks: Set<string> } | null = null;

function collaborationSupportedNames(): {
  nodes: ReadonlySet<string>;
  marks: ReadonlySet<string>;
} {
  if (supportedNames) return supportedNames;

  const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
  const nodes = new Set<string>(["doc", "text"]);
  for (const mapping of adapter.nodeMappings) {
    nodes.add(mapping.content.name);
    if (mapping.outer) nodes.add(mapping.outer.name);
  }
  const marks = new Set<string>(
    adapter.markMappings.map(mapping => mapping.prosemirrorMark.name),
  );
  supportedNames = { nodes, marks };

  return supportedNames;
}
