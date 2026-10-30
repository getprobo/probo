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

import { type Node as PmNode, Schema } from "@tiptap/pm/model";
import { EditorState, type Transaction } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { describe, expect, it } from "vitest";

import {
  anyAttachmentAccept,
  attachmentKind,
  fileAccept,
  imageAccept,
  maxAttachmentSize,
  prepareAttachment,
} from "./attachments";
import { uploadAttachments } from "./uploadAttachments";

describe("anyAttachmentAccept", () => {
  it("offers images and other files from one picker", () => {
    const accepted = anyAttachmentAccept.split(",");

    expect(accepted).toEqual(expect.arrayContaining([
      ".png",
      ".pdf",
      "image/png",
      "application/pdf",
    ]));
    expect(fileAccept.split(",")).not.toContain("image/png");
    expect(imageAccept.split(",")).not.toContain("application/pdf");
  });
});

describe("attachmentKind", () => {
  it("accepts a csv file reported as excel", () => {
    const file = new File(["a,b"], "report.csv", { type: "application/vnd.ms-excel" });

    expect(attachmentKind(file)).toBe("file");
  });

  it("accepts an xls file reported as excel", () => {
    const file = new File(["x"], "sheet.xls", { type: "application/vnd.ms-excel" });

    expect(attachmentKind(file)).toBe("file");
  });
});

describe("prepareAttachment", () => {
  it("normalizes an alias when the extension has one canonical type", () => {
    const jpeg = prepareAttachment(new File(["x"], "scan.jpg", { type: "image/pjpeg" }));
    const jpegAlias = prepareAttachment(new File(["x"], "scan.jpg", { type: "image/jpg" }));
    const png = prepareAttachment(new File(["x"], "scan.png", { type: "image/jpg" }));

    expect(jpeg.type).toBe("image/jpeg");
    expect(attachmentKind(jpeg)).toBe("image");
    expect(jpegAlias.type).toBe("image/jpeg");
    expect(png.type).toBe("image/png");
    expect(attachmentKind(png)).toBe("image");
  });

  it("keeps a reported mime when the extension has several types", () => {
    const file = prepareAttachment(
      new File(["a,b"], "report.csv", { type: "application/vnd.ms-excel" }),
    );

    expect(file.type).toBe("application/vnd.ms-excel");
    expect(attachmentKind(file)).toBe("file");
  });

  it("uses the first allowed mime when the reported type is not allowed", () => {
    const csv = prepareAttachment(
      new File(["a,b"], "report.csv", { type: "application/octet-stream" }),
    );
    const json = prepareAttachment(
      new File(["{}"], "data.json", { type: "application/octet-stream" }),
    );
    const yaml = prepareAttachment(
      new File(["a: 1"], "data.yaml", { type: "application/octet-stream" }),
    );
    const log = prepareAttachment(
      new File(["x"], "app.log", { type: "application/octet-stream" }),
    );

    expect(csv.type).toBe("application/vnd.ms-excel");
    expect(attachmentKind(csv)).toBe("file");
    expect(json.type).toBe("application/json");
    expect(attachmentKind(json)).toBe("file");
    expect(yaml.type).toBe("application/yaml");
    expect(attachmentKind(yaml)).toBe("file");
    expect(log.type).toBe("text/plain");
    expect(attachmentKind(log)).toBe("file");
  });
});

const editorSchema = new Schema({
  nodes: {
    doc: { content: "block+" },
    paragraph: { content: "inline*", group: "block" },
    text: { group: "inline" },
    bulletList: { content: "listItem+", group: "block" },
    listItem: { content: "paragraph block*" },
    blockquote: { content: "block+", group: "block" },
    image: {
      group: "block",
      atom: true,
      attrs: {
        fileId: { default: null },
        src: { default: null },
        alt: { default: null },
        title: { default: null },
      },
    },
    file: {
      group: "block",
      atom: true,
      attrs: {
        fileId: { default: null },
        fileName: { default: "" },
        mimeType: { default: "" },
        size: { default: 0 },
      },
    },
    contentUpload: {
      group: "block",
      atom: true,
      attrs: {
        id: { default: null },
        fileName: { default: "" },
      },
    },
  },
});

function textPos(doc: PmNode, text: string, offset: number) {
  let found = -1;
  doc.descendants((node, pos) => {
    if (found >= 0 || !node.isText || node.text !== text) {
      return found < 0;
    }

    found = pos + offset;

    return false;
  });

  return found;
}

function editorView(doc: PmNode) {
  const view = {
    isDestroyed: false,
    editable: true,
    state: EditorState.create({ schema: editorSchema, doc }),
    dispatch(tr: Transaction) {
      view.state = view.state.apply(tr);
    },
  };

  return view as unknown as EditorView;
}

describe("uploadAttachments", () => {
  it("reports an oversized file separately from an unsupported type", async () => {
    const oversized = new File(["x"], "scan.pdf", { type: "application/pdf" });
    Object.defineProperty(oversized, "size", { value: maxAttachmentSize + 1 });
    const unknown = new File(["x"], "notes.exe", { type: "application/octet-stream" });
    const reasons: string[] = [];

    await uploadAttachments(
      { isDestroyed: false, editable: true } as EditorView,
      [oversized, unknown],
      0,
      () => Promise.reject(new Error("unused")),
      error => reasons.push(error.reason),
    );

    expect(reasons).toEqual(["tooLarge", "unsupported"]);
  });

  it("inserts a file at the caret inside a list item", async () => {
    const doc = editorSchema.node("doc", null, [
      editorSchema.node("bulletList", null, [
        editorSchema.node("listItem", null, [
          editorSchema.node("paragraph", null, [editorSchema.text("ab")]),
        ]),
        editorSchema.node("listItem", null, [
          editorSchema.node("paragraph", null, [editorSchema.text("cd")]),
        ]),
      ]),
    ]);
    const view = editorView(doc);

    await uploadAttachments(
      view,
      [new File(["x"], "scan.png", { type: "image/png" })],
      textPos(doc, "ab", 1),
      () => Promise.resolve({ id: "file-1", fileName: "scan.png", mimeType: "image/png", size: 1 }),
      () => {},
    );

    expect(view.state.doc.toString()).toBe(
      "doc(bulletList(listItem(paragraph(\"a\"), image, paragraph(\"b\")), listItem(paragraph(\"cd\"))))",
    );
  });

  it("inserts a file at the caret inside a blockquote", async () => {
    const doc = editorSchema.node("doc", null, [
      editorSchema.node("blockquote", null, [
        editorSchema.node("paragraph", null, [editorSchema.text("hello")]),
      ]),
    ]);
    const view = editorView(doc);

    await uploadAttachments(
      view,
      [new File(["x"], "notes.pdf", { type: "application/pdf" })],
      textPos(doc, "hello", 2),
      () => Promise.resolve({ id: "file-2", fileName: "notes.pdf", mimeType: "application/pdf", size: 1 }),
      () => {},
    );

    expect(view.state.doc.toString()).toBe(
      "doc(blockquote(paragraph(\"he\"), file, paragraph(\"llo\")))",
    );
  });

  it("does not insert a file when the placeholder was removed", async () => {
    const doc = editorSchema.node("doc", null, [
      editorSchema.node("paragraph", null, [editorSchema.text("ab")]),
    ]);
    const view = editorView(doc);
    let resolveUpload: (file: {
      id: string;
      fileName: string;
      mimeType: string;
      size: number;
    }) => void = () => {};

    const pending = uploadAttachments(
      view,
      [new File(["x"], "scan.png", { type: "image/png" })],
      textPos(doc, "ab", 1),
      () => new Promise((resolve) => {
        resolveUpload = resolve;
      }),
      () => {},
    );

    let removed = false;
    view.state.doc.descendants((node, pos) => {
      if (removed || node.type.name !== "contentUpload") {
        return !removed;
      }

      view.dispatch(view.state.tr.delete(pos, pos + node.nodeSize));
      removed = true;

      return false;
    });
    resolveUpload({
      id: "file-removed",
      fileName: "scan.png",
      mimeType: "image/png",
      size: 1,
    });
    await pending;

    expect(removed).toBe(true);
    expect(view.state.doc.toString()).toBe("doc(paragraph(\"a\"), paragraph(\"b\"))");
  });
});
