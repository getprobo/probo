// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import DOMPurify from "dompurify";
import type { Token } from "marked";
import type {
  Mark as ProseMirrorMark,
  Node as ProseMirrorNode,
  Schema,
} from "@tiptap/pm/model";

const blockOpenTagPattern =
  /^<(ul|ol|table|blockquote|pre|div|h[1-6]|hr)[\s>\/]/i;

const blockTags = new Set([
  "p", "div", "ul", "ol", "blockquote", "pre", "table",
  "h1", "h2", "h3", "h4", "h5", "h6", "hr",
]);

export function safeLinkHref(href: string): string {
  href = href.trim();
  if (!href) return "#";
  if (href[0] === "#") return href;
  if (href.startsWith("/")) {
    if (href.length > 1 && (href[1] === "/" || href[1] === "\\")) return "#";
    return href;
  }
  try {
    const url = new URL(href);
    const scheme = url.protocol.slice(0, -1).toLowerCase();
    switch (scheme) {
      case "http":
      case "https":
      case "mailto":
      case "tel":
        return href;
      default:
        return "#";
    }
  } catch {
    return href;
  }
}

export function cellHasBlockHTML(tokens: Token[]): boolean {
  return tokens.some(
    t => t.type === "html" && blockOpenTagPattern.test(t.raw),
  );
}

export function convertHTMLBlockContent(
  html: string,
  schema: Schema,
): ProseMirrorNode[] {
  const sanitized = DOMPurify.sanitize(html).trim();
  if (!sanitized) return [];
  const doc = new DOMParser().parseFromString(sanitized, "text/html");
  const c = new HTMLBlockConverter(schema);
  const nodes: ProseMirrorNode[] = [];
  for (const child of Array.from(doc.body.childNodes)) {
    nodes.push(...c.convertBlockNode(child));
  }
  return nodes;
}

class HTMLBlockConverter {
  private schema: Schema;
  private marks: ProseMirrorMark[];

  constructor(schema: Schema) {
    this.schema = schema;
    this.marks = [];
  }

  private withMark(
    mark: ProseMirrorMark,
    fn: () => ProseMirrorNode[],
  ): ProseMirrorNode[] {
    this.marks.push(mark);
    const result = fn();
    this.marks.pop();
    return result;
  }

  private currentMarks(): readonly ProseMirrorMark[] {
    let result: readonly ProseMirrorMark[] = [];
    for (const mark of this.marks) {
      result = mark.addToSet(result as ProseMirrorMark[]);
    }
    return result;
  }

  convertBlockNode(node: Node): ProseMirrorNode[] {
    if (node.nodeType === Node.TEXT_NODE) {
      const text = node.textContent?.trim();
      if (!text) return [];
      return [
        this.schema.nodes.paragraph.create(null, this.schema.text(text)),
      ];
    }
    if (node.nodeType !== Node.ELEMENT_NODE) return [];

    const el = node as Element;
    switch (el.tagName.toLowerCase()) {
      case "ul":
        return this.convertList(el, false);
      case "ol":
        return this.convertList(el, true);
      case "blockquote": {
        const children = this.convertBlockChildren(el);
        if (children.length === 0) return [];
        return [this.schema.nodes.blockquote.create(null, children)];
      }
      case "pre": {
        const codeEl = el.querySelector("code");
        const text = codeEl?.textContent ?? el.textContent ?? "";
        const lang =
          codeEl?.className?.match(/language-(\w+)/)?.[1] ?? null;
        return [
          this.schema.nodes.codeBlock.create(
            { language: lang },
            text ? this.schema.text(text) : undefined,
          ),
        ];
      }
      case "hr":
        return [this.schema.nodes.horizontalRule.create()];
      case "h1":
      case "h2":
      case "h3":
      case "h4":
      case "h5":
      case "h6": {
        const level = parseInt(el.tagName[1]);
        const inlines = this.convertInlineChildren(el);
        return [
          this.schema.nodes.heading.create(
            { level },
            inlines.length > 0 ? inlines : undefined,
          ),
        ];
      }
      default: {
        const inlines = this.convertInlineChildren(el);
        if (inlines.length === 0) return [];
        return [this.schema.nodes.paragraph.create(null, inlines)];
      }
    }
  }

  private convertBlockChildren(parent: Node): ProseMirrorNode[] {
    const nodes: ProseMirrorNode[] = [];
    for (const child of Array.from(parent.childNodes)) {
      nodes.push(...this.convertBlockNode(child));
    }
    return nodes;
  }

  private convertList(
    el: Element,
    ordered: boolean,
  ): ProseMirrorNode[] {
    const items: ProseMirrorNode[] = [];
    for (const child of Array.from(el.children)) {
      if (child.tagName.toLowerCase() !== "li") continue;
      const content = this.convertListItem(child);
      items.push(
        this.schema.nodes.listItem.create(
          null,
          content.length > 0
            ? content
            : this.schema.nodes.paragraph.create(),
        ),
      );
    }
    if (items.length === 0) return [];

    if (ordered) {
      const start = parseInt(el.getAttribute("start") ?? "1") || 1;
      return [this.schema.nodes.orderedList.create({ start }, items)];
    }
    return [this.schema.nodes.bulletList.create(null, items)];
  }

  private convertListItem(el: Element): ProseMirrorNode[] {
    const hasBlock = Array.from(el.children).some(child =>
      blockTags.has(child.tagName.toLowerCase()),
    );

    if (hasBlock) {
      return this.convertBlockChildren(el);
    }

    const inlines = this.convertInlineChildren(el);
    if (inlines.length === 0) return [];
    return [this.schema.nodes.paragraph.create(null, inlines)];
  }

  private convertInlineChildren(parent: Node): ProseMirrorNode[] {
    const nodes: ProseMirrorNode[] = [];
    for (const child of Array.from(parent.childNodes)) {
      nodes.push(...this.convertInlineNode(child));
    }
    return nodes;
  }

  private convertInlineNode(node: Node): ProseMirrorNode[] {
    if (node.nodeType === Node.TEXT_NODE) {
      const text = node.textContent ?? "";
      if (!text) return [];
      const marks = this.currentMarks();
      return [this.schema.text(text, marks.length > 0 ? marks : undefined)];
    }
    if (node.nodeType !== Node.ELEMENT_NODE) return [];

    const el = node as Element;
    switch (el.tagName.toLowerCase()) {
      case "strong":
      case "b":
        return this.withMark(this.schema.marks.bold.create(), () =>
          this.convertInlineChildren(el),
        );
      case "em":
      case "i":
        return this.withMark(this.schema.marks.italic.create(), () =>
          this.convertInlineChildren(el),
        );
      case "s":
      case "del":
      case "strike":
        return this.withMark(this.schema.marks.strike.create(), () =>
          this.convertInlineChildren(el),
        );
      case "u":
        return this.withMark(this.schema.marks.underline.create(), () =>
          this.convertInlineChildren(el),
        );
      case "code":
        return this.withMark(this.schema.marks.code.create(), () =>
          this.convertInlineChildren(el),
        );
      case "a": {
        const href = el.getAttribute("href") ?? "#";
        const attrs: Record<string, unknown> = {
          href: safeLinkHref(href),
        };
        const title = el.getAttribute("title");
        if (title) attrs.title = title;
        return this.withMark(this.schema.marks.link.create(attrs), () =>
          this.convertInlineChildren(el),
        );
      }
      case "br":
        return [this.schema.nodes.hardBreak.create()];
      case "img": {
        const src = el.getAttribute("src");
        if (!src) return [];
        const attrs: Record<string, unknown> = { src };
        const alt = el.getAttribute("alt");
        const title = el.getAttribute("title");
        if (alt) attrs.alt = alt;
        if (title) attrs.title = title;
        return [this.schema.nodes.image.create(attrs)];
      }
      default:
        return this.convertInlineChildren(el);
    }
  }
}
