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

import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { recoverRichContent } from "../src/RichEditor/recoverContent.ts";

const diagramRequest = {
  type: "doc",
  content: [
    {
      type: "paragraph",
      content: [{ type: "text", text: "Add three Mermaid diagrams" }],
    },
    {
      type: "bulletList",
      content: [
        {
          type: "listItem",
          content: [
            {
              type: "paragraph",
              content: [{ type: "text", text: "Network" }],
            },
          ],
        },
        {
          type: "listItem",
          content: [
            {
              type: "paragraph",
              content: [{ type: "text", text: "Data flow" }],
            },
          ],
        },
        {
          type: "listItem",
          content: [
            {
              type: "paragraph",
              content: [{ type: "text", text: "SDLC" }],
            },
          ],
        },
      ],
    },
  ],
};

function plainTextDocument(text: string): string {
  return JSON.stringify({
    type: "doc",
    content: text.split("\n").map(line => ({
      type: "paragraph",
      ...(line ? { content: [{ type: "text", text: line }] } : {}),
    })),
  });
}

describe("recoverRichContent", () => {
  it("renders a document stored as paragraphs of json", () => {
    const pretty = JSON.stringify(diagramRequest, null, 2);
    const recovered = recoverRichContent(plainTextDocument(pretty));

    assert.ok(recovered);
    assert.equal(recovered.content?.[0]?.content?.[0]?.text, "Add three Mermaid diagrams");
    assert.equal(recovered.content?.[1]?.type, "bulletList");
    assert.equal(
      recovered.content?.[1]?.content?.[1]?.content?.[0]?.content?.[0]?.text,
      "Data flow",
    );
    assert.equal(
      recovered.content?.[0]?.content?.[0]?.text?.includes("bulletList"),
      false,
    );
  });

  it("unwraps a json string around the document", () => {
    const recovered = recoverRichContent(JSON.stringify(JSON.stringify(diagramRequest)));

    assert.equal(recovered?.content?.[1]?.type, "bulletList");
    assert.equal(
      recovered?.content?.[1]?.content?.[0]?.content?.[0]?.content?.[0]?.text,
      "Network",
    );
    assert.equal(
      recovered?.content?.[1]?.content?.[2]?.content?.[0]?.content?.[0]?.text,
      "SDLC",
    );
  });

  it("keeps a real bullet list", () => {
    const stored = JSON.stringify(diagramRequest);
    const recovered = recoverRichContent(stored);

    assert.equal(recovered?.content?.[1]?.type, "bulletList");
    assert.equal(
      recovered?.content?.[1]?.content?.[0]?.content?.[0]?.content?.[0]?.text,
      "Network",
    );
  });

  it("keeps prose", () => {
    const recovered = recoverRichContent(plainTextDocument("Keep this sentence"));

    assert.equal(recovered?.content?.[0]?.content?.[0]?.text, "Keep this sentence");
  });

  it("returns null for plain text", () => {
    assert.equal(recoverRichContent("not json"), null);
  });
});
