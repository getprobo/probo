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

import type {
  MappedMarkSpec,
  MappedNodeSpec,
} from "@automerge/prosemirror";
import { Extension, getExtensionField } from "@tiptap/core";
import type { NodeSpec } from "@tiptap/pm/model";

// A top-level text block (paragraph, heading, blockquote, code block) only
// earns a stable Automerge identity at the document root; a list item earns one
// wherever it sits. explicitBlockIdentityPlugin keys off this distinction.
export type AutomergeExplicitIdentity = "topLevelBlock" | "listItem";

// AutomergeNodeMapping extends the upstream block mapping with the two flags the
// frontend needs that the ProseMirror <-> Automerge adapter itself ignores:
// `structural` (leaf or container that must carry isAmgBlock) and
// `explicitIdentity` (which blocks the identity plugin promotes).
export type AutomergeNodeMapping = NonNullable<MappedNodeSpec["automerge"]> & {
  structural?: boolean;
  explicitIdentity?: AutomergeExplicitIdentity;
};

export type AutomergeMarkMapping = NonNullable<MappedMarkSpec["automerge"]>;

declare module "@tiptap/core" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface NodeConfig<Options, Storage> {
    automerge?: () => AutomergeNodeMapping;
  }
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface MarkConfig<Options, Storage> {
    automerge?: () => AutomergeMarkMapping;
  }
}

// AutomergeMappingExtension copies each extension's `automerge()` mapping onto
// its ProseMirror schema spec, so getSchema() carries the mapping and the schema
// adapter can read it without a name switch. It is inert on the base editor
// schema, which omits this extension and keeps the mapping off the spec.
export const AutomergeMappingExtension = Extension.create({
  name: "automergeMapping",

  extendNodeSchema(extension) {
    const mapping = getExtensionField<(() => AutomergeNodeMapping) | undefined>(
      extension,
      "automerge",
    );

    return mapping ? { automerge: mapping() } : {};
  },

  extendMarkSchema(extension) {
    const mapping = getExtensionField<(() => AutomergeMarkMapping) | undefined>(
      extension,
      "automerge",
    );

    return mapping ? { automerge: mapping() } : {};
  },
});

export function readAutomergeNodeMapping(
  spec: NodeSpec,
): AutomergeNodeMapping | undefined {
  return (spec as { automerge?: AutomergeNodeMapping }).automerge;
}

export function parseJSONObject(value: string): Record<string, unknown> {
  const parsed: unknown = JSON.parse(value);
  if (!isJSONObject(parsed)) {
    throw new Error("expected a JSON object");
  }
  return parsed;
}

export function isJSONObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function presentStrings(
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

export function readBooleanAttribute(attributes: unknown, name: string): boolean {
  if (!isJSONObject(attributes)) return false;

  return attributes[name] === true;
}

export function readNumberAttribute(
  attributes: unknown,
  name: string,
  fallback: number,
): number {
  if (!isJSONObject(attributes)) return fallback;
  const value = attributes[name];
  return typeof value === "number" ? value : fallback;
}

export function readStringAttribute(
  attributes: unknown,
  name: string,
  fallback: string,
): string {
  if (!isJSONObject(attributes)) return fallback;
  const value = attributes[name];
  return typeof value === "string" ? value : fallback;
}

export function readNullableStringAttribute(
  attributes: unknown,
  name: string,
): string | null {
  if (!isJSONObject(attributes)) return null;
  const value = attributes[name];
  return typeof value === "string" ? value : null;
}

export function readNumberArrayAttribute(
  attributes: unknown,
  name: string,
): number[] | null {
  if (!isJSONObject(attributes)) return null;
  const value = attributes[name];
  if (!Array.isArray(value)) return null;

  const numbers = value.filter(item => typeof item === "number");
  return numbers.length === value.length ? numbers : null;
}
