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

import type { Node as PmNode } from "@tiptap/pm/model";

export type DomNode = {
  nodeType: number;
  nodeName: string;
  nodeValue: string | null;
  tagName?: string;
  style?: { whiteSpace?: string };
  childNodes: DomNode[];
  firstChild: DomNode | null;
  nextSibling: DomNode | null;
  previousSibling: DomNode | null;
  getAttribute?: (name: string) => string | null;
  matches?: (selector: string) => boolean;
  contains?: (node: DomNode) => boolean;
  textContent?: string | null;
};

export function text(value: string): DomNode {
  return {
    nodeType: 3,
    nodeName: "#text",
    nodeValue: value,
    childNodes: [],
    firstChild: null,
    nextSibling: null,
    previousSibling: null,
    textContent: value,
  };
}

export function element(tag: string, attrs: Record<string, string>, children: DomNode[]): DomNode {
  link(children);

  return {
    nodeType: 1,
    nodeName: tag.toUpperCase(),
    tagName: tag.toUpperCase(),
    nodeValue: null,
    style: {},
    childNodes: children,
    firstChild: children[0] ?? null,
    nextSibling: null,
    previousSibling: null,
    contains: () => false,
    get textContent() {
      return children.map(child => child.textContent ?? "").join("");
    },
    getAttribute: name => attrs[name] ?? null,
    matches: (selector) => {
      if (selector === "*") {
        return true;
      }
      const tagName = selector.match(/^[a-zA-Z0-9-]+/)?.[0];
      if (tagName && tagName.toLowerCase() !== tag.toLowerCase()) {
        return false;
      }
      const attr = selector.match(/\[([^\]=]+)(?:=(?:"([^"]*)"|'([^']*)'))?\]/);
      if (!attr) {
        return tagName !== undefined;
      }
      const actual = attrs[attr[1]];
      if (actual === undefined) {
        return false;
      }
      const expected = attr[2] ?? attr[3];

      return expected === undefined || actual === expected;
    },
  };
}

export function nodeNames(node: PmNode): string[] {
  const names: string[] = [];
  node.descendants((child) => {
    names.push(child.type.name);
  });

  return names;
}

function link(children: DomNode[]) {
  for (let i = 0; i < children.length; i++) {
    children[i].nextSibling = children[i + 1] ?? null;
    children[i].previousSibling = children[i - 1] ?? null;
  }
}
