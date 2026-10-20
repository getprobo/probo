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

import { getSchema } from "@tiptap/core";
import { DOMParser, type Node as PmNode } from "@tiptap/pm/model";
import { StarterKit } from "@tiptap/starter-kit";
import { describe, expect, it } from "vitest";

import { attachmentPath } from "./attachments";
import { FileExtension } from "./FileExtension";
import { element, text } from "./htmlDom";

describe("FileExtension", () => {
  it("trims a pasted file id", () => {
    const files = fileNodes(parse(element("a", { "data-file-id": "  file-1  " }, [text("policy.pdf")])));

    expect(files).toHaveLength(1);
    expect(files[0]?.attrs.fileId).toBe("file-1");
    expect(files[0]?.attrs.fileName).toBe("policy.pdf");
  });

  it("drops a pasted file anchor with a blank id", () => {
    expect(fileNodes(parse(element("a", { "data-file-id": "" }, [text("policy.pdf")])))).toHaveLength(0);
    expect(fileNodes(parse(element("a", { "data-file-id": "   " }, [text("policy.pdf")])))).toHaveLength(0);
  });

  it("recognizes a rendered attachment href", () => {
    const files = fileNodes(parse(element("a", {
      href: "/api/files/v1/file-1",
      class: "document-file",
    }, [text("policy.pdf")])));

    expect(files).toHaveLength(1);
    expect(files[0]?.attrs.fileId).toBe("file-1");
    expect(files[0]?.attrs.fileName).toBe("policy.pdf");
  });

  it("recognizes an absolute attachment href", () => {
    const files = fileNodes(parse(element("a", {
      href: "https://app.example/api/files/v1/file-1",
    }, [text("policy.pdf")])));

    expect(files).toHaveLength(1);
    expect(files[0]?.attrs.fileId).toBe("file-1");
  });

  it("leaves an ordinary link as a hyperlink", () => {
    const doc = parse(element("a", { href: "https://example.com/policy.pdf" }, [text("policy.pdf")]));

    expect(fileNodes(doc)).toHaveLength(0);
    expect(hasLinkMark(doc)).toBe(true);
  });

  it("parses a rendered file anchor as a file node", () => {
    const doc = parse(element("a", {
      "href": attachmentPath("file-1"),
      "data-file-id": "file-1",
    }, [text("policy.pdf")]));
    const files = fileNodes(doc);

    expect(files).toHaveLength(1);
    expect(files[0]?.attrs.fileId).toBe("file-1");
    expect(files[0]?.attrs.fileName).toBe("policy.pdf");
    expect(hasLinkMark(doc)).toBe(false);
  });
});

function parse(anchor: ReturnType<typeof element>): PmNode {
  const schema = getSchema([StarterKit, FileExtension]);

  return DOMParser.fromSchema(schema).parse(element("body", {}, [anchor]) as unknown as Node);
}

function fileNodes(doc: PmNode): PmNode[] {
  const files: PmNode[] = [];
  doc.descendants((node) => {
    if (node.type.name === "file") {
      files.push(node);
    }
  });

  return files;
}

function hasLinkMark(doc: PmNode): boolean {
  let linked = false;
  doc.descendants((node) => {
    if (node.marks.some(mark => mark.type.name === "link")) {
      linked = true;
    }
  });

  return linked;
}
