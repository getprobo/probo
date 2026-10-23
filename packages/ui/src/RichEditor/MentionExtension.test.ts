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

import { getSchema, getText, getTextSerializersFromSchema } from "@tiptap/core";
import { Node } from "@tiptap/pm/model";
import { StarterKit } from "@tiptap/starter-kit";
import { describe, expect, it } from "vitest";

import { MentionExtension } from "./MentionExtension";

const doc = {
  type: "doc",
  content: [
    {
      type: "paragraph",
      content: [
        { type: "text", text: "Ping " },
        { type: "mention", attrs: { id: "profile-1", label: "Jane Doe" } },
      ],
    },
  ],
};

describe("MentionExtension", () => {
  it("keeps the mentioned id and label", () => {
    const schema = getSchema([StarterKit, MentionExtension]);
    const node = Node.fromJSON(schema, doc);
    const mention = node.firstChild?.lastChild;

    expect(mention?.type.name).toBe("mention");
    expect(mention?.attrs.id).toBe("profile-1");
    expect(mention?.attrs.label).toBe("Jane Doe");
  });

  it("renders as @label in plain text", () => {
    const schema = getSchema([StarterKit, MentionExtension]);
    const node = Node.fromJSON(schema, doc);

    const text = getText(node, {
      textSerializers: getTextSerializersFromSchema(schema),
    });

    expect(text).toBe("Ping @Jane Doe");
  });
});
