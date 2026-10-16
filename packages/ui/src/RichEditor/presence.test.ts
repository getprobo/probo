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

import { Schema } from "@tiptap/pm/model";
import { describe, expect, it } from "vitest";

import {
  decorationsForPresences,
  presenceHighlightColor,
  type RichEditorPresence,
} from "./presence";

const schema = new Schema({
  nodes: {
    doc: { content: "block+" },
    paragraph: { content: "inline*", group: "block" },
    text: { group: "inline" },
  },
});

function presence(overrides: Partial<RichEditorPresence> = {}): RichEditorPresence {
  return {
    connectionID: "conn-1",
    identityID: "collaborator-1",
    displayName: "Ada",
    active: true,
    anchorPosition: 1,
    headPosition: 1,
    ...overrides,
  };
}

describe("presence highlight color", () => {
  it("returns a translucent hsl color, not hex-alpha on hsl()", () => {
    const color = presenceHighlightColor("collaborator-1");

    expect(color).toMatch(/^hsl\(\d+ 70% 45% \/ 0\.28\)$/);
    expect(color).not.toMatch(/\)33$/);
  });
});

describe("presence decorations", () => {
  it("skips caret decorations for inactive peers", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("hello")]),
    ]);

    const decorations = decorationsForPresences(doc, [
      presence({
        active: false,
        connectionID: "idle",
        anchorPosition: 3,
        headPosition: 3,
      }),
      presence({
        active: true,
        connectionID: "active",
        identityID: "collaborator-2",
        anchorPosition: 2,
        headPosition: 2,
      }),
    ]);

    expect(decorations.find().length).toBe(1);
  });
});
