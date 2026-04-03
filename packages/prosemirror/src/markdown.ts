// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { lexer, type Token, type Tokens } from "marked";
import type {
  Mark as ProseMirrorMark,
  Node as ProseMirrorNode,
  Schema,
} from "@tiptap/pm/model";

import {
  cellHasBlockHTML,
  convertHTMLBlockContent,
  safeLinkHref,
} from "./htmlBlock";

const tableSeparatorPattern =
  /^\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)*\|?\s*$/m;
const markdownLinePattern =
  /(?:^```\w*$|^#{1,6}\s|^>\s|^[-+*]\s|^\d+\.\s|^(?:---|___|\*\*\*)\s*$)/m;

function unescapeHtml(html: string): string {
  return html
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&amp;/g, "&");
}

type HtmlTagInfo = {
  closing: boolean;
  markName: string;
  attrs?: Record<string, unknown>;
};

const openTagPattern = /^<(u|b|i|s|em|strong|del|code|a)(\s[^>]*)?\s*\/?>$/i;
const closeTagPattern = /^<\/(u|b|i|s|em|strong|del|code|a)\s*>$/i;

function tagToMarkName(tag: string): string | null {
  switch (tag.toLowerCase()) {
    case "u":
      return "underline";
    case "b":
    case "strong":
      return "bold";
    case "i":
    case "em":
      return "italic";
    case "s":
    case "del":
      return "strike";
    case "code":
      return "code";
    case "a":
      return "link";
    default:
      return null;
  }
}

function parseHtmlTag(html: string): HtmlTagInfo | null {
  const closeMatch = html.match(closeTagPattern);
  if (closeMatch) {
    const markName = tagToMarkName(closeMatch[1]);
    if (markName) return { closing: true, markName };
    return null;
  }

  const openMatch = html.match(openTagPattern);
  if (openMatch) {
    const markName = tagToMarkName(openMatch[1]);
    if (!markName) return null;

    const info: HtmlTagInfo = { closing: false, markName };

    if (openMatch[1].toLowerCase() === "a" && openMatch[2]) {
      const hrefMatch = openMatch[2].match(/href=["']([^"']*)["']/i);
      if (hrefMatch) {
        info.attrs = { href: safeLinkHref(hrefMatch[1]) };
      }
    }

    return info;
  }

  return null;
}

class Converter {
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

  private makeTextNode(text: string): ProseMirrorNode[] {
    const unescaped = unescapeHtml(text);
    if (!unescaped) return [];
    const marks = this.currentMarks();
    return [this.schema.text(unescaped, marks.length > 0 ? marks : undefined)];
  }

  convertBlockTokens(tokens: Token[]): ProseMirrorNode[] {
    const nodes: ProseMirrorNode[] = [];
    for (const token of tokens) {
      nodes.push(...this.convertBlockToken(token));
    }
    return nodes;
  }

  private convertBlockToken(token: Token): ProseMirrorNode[] {
    switch (token.type) {
      case "heading":
        return this.convertHeading(token as Tokens.Heading);
      case "paragraph":
        return this.convertParagraph(token as Tokens.Paragraph);
      case "blockquote":
        return this.convertBlockquote(token as Tokens.Blockquote);
      case "code":
        return this.convertCodeBlock(token as Tokens.Code);
      case "list":
        return this.convertList(token as Tokens.List);
      case "hr":
        return [this.schema.nodes.horizontalRule.create()];
      case "table":
        return this.convertTable(token as Tokens.Table);
      case "text":
        return this.convertBlockText(token as Tokens.Text);
      case "html":
      case "space":
      case "def":
        return [];
      default:
        return [];
    }
  }

  private convertHeading(token: Tokens.Heading): ProseMirrorNode[] {
    const content = this.convertInlineTokens(token.tokens);
    return [
      this.schema.nodes.heading.create(
        { level: token.depth },
        content.length > 0 ? content : undefined,
      ),
    ];
  }

  private convertParagraph(token: Tokens.Paragraph): ProseMirrorNode[] {
    const content = this.convertInlineTokens(token.tokens);
    return [
      this.schema.nodes.paragraph.create(
        null,
        content.length > 0 ? content : undefined,
      ),
    ];
  }

  private convertBlockquote(token: Tokens.Blockquote): ProseMirrorNode[] {
    const children = this.convertBlockTokens(token.tokens);
    return [
      this.schema.nodes.blockquote.create(
        null,
        children.length > 0 ? children : undefined,
      ),
    ];
  }

  private convertCodeBlock(token: Tokens.Code): ProseMirrorNode[] {
    const language = token.lang || null;
    return [
      this.schema.nodes.codeBlock.create(
        { language },
        token.text ? this.schema.text(token.text) : undefined,
      ),
    ];
  }

  private convertList(token: Tokens.List): ProseMirrorNode[] {
    const children = token.items.flatMap(item => this.convertListItem(item));

    if (token.ordered) {
      const start = typeof token.start === "number" ? token.start : 1;
      return [
        this.schema.nodes.orderedList.create(
          { start },
          children.length > 0 ? children : undefined,
        ),
      ];
    }

    return [
      this.schema.nodes.bulletList.create(
        null,
        children.length > 0 ? children : undefined,
      ),
    ];
  }

  private convertListItem(token: Tokens.ListItem): ProseMirrorNode[] {
    const children: ProseMirrorNode[] = [];

    for (const child of token.tokens) {
      if (child.type === "text") {
        const textToken = child as Tokens.Text;
        const inlineContent = textToken.tokens
          ? this.convertInlineTokens(textToken.tokens)
          : this.makeTextNode(textToken.text);
        children.push(
          this.schema.nodes.paragraph.create(
            null,
            inlineContent.length > 0 ? inlineContent : undefined,
          ),
        );
      } else {
        children.push(...this.convertBlockToken(child));
      }
    }

    if (children.length === 0) {
      children.push(this.schema.nodes.paragraph.create());
    }

    return [this.schema.nodes.listItem.create(null, children)];
  }

  private convertBlockText(token: Tokens.Text): ProseMirrorNode[] {
    const content = token.tokens
      ? this.convertInlineTokens(token.tokens)
      : this.makeTextNode(token.text);
    return [
      this.schema.nodes.paragraph.create(
        null,
        content.length > 0 ? content : undefined,
      ),
    ];
  }

  private convertTable(token: Tokens.Table): ProseMirrorNode[] {
    const rows: ProseMirrorNode[] = [];

    const headerCells = token.header.map(cell =>
      this.convertTableCell(cell, true),
    );
    rows.push(this.schema.nodes.tableRow.create(null, headerCells));

    for (const row of token.rows) {
      const dataCells = row.map(cell => this.convertTableCell(cell, false));
      rows.push(this.schema.nodes.tableRow.create(null, dataCells));
    }

    return [this.schema.nodes.table.create(null, rows)];
  }

  private convertTableCell(
    cell: Tokens.TableCell,
    isHeader: boolean,
  ): ProseMirrorNode {
    const nodeType = isHeader
      ? this.schema.nodes.tableHeader
      : this.schema.nodes.tableCell;

    if (cellHasBlockHTML(cell.tokens)) {
      const raw = cell.tokens.map(t => t.raw).join("");
      const blockContent = convertHTMLBlockContent(raw, this.schema);
      if (blockContent.length > 0) {
        return nodeType.create(null, blockContent);
      }
    }

    const content = this.convertInlineTokens(cell.tokens);
    return nodeType.create(
      null,
      this.schema.nodes.paragraph.create(
        null,
        content.length > 0 ? content : undefined,
      ),
    );
  }

  convertInlineTokens(tokens: Token[]): ProseMirrorNode[] {
    const nodes: ProseMirrorNode[] = [];

    for (const token of tokens) {
      switch (token.type) {
        case "text": {
          const textToken = token as Tokens.Text;
          if (textToken.tokens && textToken.tokens.length > 0) {
            nodes.push(...this.convertInlineTokens(textToken.tokens));
          } else {
            nodes.push(...this.makeTextNode(textToken.text));
          }
          break;
        }

        case "strong":
          nodes.push(
            ...this.withMark(this.schema.marks.bold.create(), () =>
              this.convertInlineTokens((token as Tokens.Strong).tokens),
            ),
          );
          break;

        case "em":
          nodes.push(
            ...this.withMark(this.schema.marks.italic.create(), () =>
              this.convertInlineTokens((token as Tokens.Em).tokens),
            ),
          );
          break;

        case "del":
          nodes.push(
            ...this.withMark(this.schema.marks.strike.create(), () =>
              this.convertInlineTokens((token as Tokens.Del).tokens),
            ),
          );
          break;

        case "codespan": {
          const codeText = unescapeHtml((token as Tokens.Codespan).text);
          if (codeText) {
            const codeMark = this.schema.marks.code.create();
            const marks = codeMark.addToSet(
              this.currentMarks() as ProseMirrorMark[],
            );
            nodes.push(this.schema.text(codeText, marks));
          }
          break;
        }

        case "link": {
          const linkToken = token as Tokens.Link;
          const attrs: Record<string, unknown> = {
            href: safeLinkHref(linkToken.href),
          };
          if (linkToken.title) {
            attrs.title = linkToken.title;
          }
          nodes.push(
            ...this.withMark(this.schema.marks.link.create(attrs), () =>
              this.convertInlineTokens(linkToken.tokens),
            ),
          );
          break;
        }

        case "image": {
          const imgToken = token as Tokens.Image;
          const attrs: Record<string, unknown> = { src: imgToken.href };
          if (imgToken.title) attrs.title = imgToken.title;
          if (imgToken.text) attrs.alt = unescapeHtml(imgToken.text);
          nodes.push(this.schema.nodes.image.create(attrs));
          break;
        }

        case "br":
          nodes.push(this.schema.nodes.hardBreak.create());
          break;

        case "escape":
          nodes.push(...this.makeTextNode((token as Tokens.Escape).text));
          break;

        case "html": {
          const tagInfo = parseHtmlTag((token as Tokens.Tag).text);
          if (tagInfo) {
            if (tagInfo.closing) {
              for (let j = this.marks.length - 1; j >= 0; j--) {
                if (this.marks[j].type.name === tagInfo.markName) {
                  this.marks.splice(j, 1);
                  break;
                }
              }
            } else {
              const markType = this.schema.marks[tagInfo.markName];
              if (markType) {
                this.marks.push(markType.create(tagInfo.attrs || null));
              }
            }
          }
          break;
        }

        default:
          break;
      }
    }

    return nodes;
  }
}

export function hasMarkdown(text: string): boolean {
  return markdownLinePattern.test(text) || tableSeparatorPattern.test(text);
}

export function parseInlineContent(
  schema: Schema,
  text: string,
): ProseMirrorNode[] {
  const tokens = lexer(text);
  const converter = new Converter(schema);

  for (const token of tokens) {
    if (token.type === "paragraph") {
      return converter.convertInlineTokens(
        (token as Tokens.Paragraph).tokens,
      );
    }
  }

  return [];
}

export function parseMarkdown(
  text: string,
  schema: Schema,
): ProseMirrorNode[] {
  const tokens = lexer(text);
  const converter = new Converter(schema);
  return converter.convertBlockTokens(tokens);
}
