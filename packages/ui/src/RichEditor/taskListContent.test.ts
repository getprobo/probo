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

import { parseMarkdown } from "@probo/prosemirror";
import { getSchema } from "@tiptap/core";
import { Document } from "@tiptap/extension-document";
import { BulletList, ListItem, OrderedList, TaskItem, TaskList } from "@tiptap/extension-list";
import { Paragraph } from "@tiptap/extension-paragraph";
import { Text } from "@tiptap/extension-text";
import { describe, expect, it } from "vitest";

import { promoteCheckboxLists } from "./taskListContent";

const schema = getSchema([
  Document,
  Paragraph,
  Text,
  BulletList,
  OrderedList,
  ListItem,
  TaskList,
  TaskItem.configure({ nested: true }),
]);

describe("promoteCheckboxLists", () => {
  it("turns checkbox syntax in a bullet list into task items", () => {
    const promoted = promoteCheckboxLists({
      type: "doc",
      content: [
        {
          type: "bulletList",
          content: [
            paragraphItem("[ ] open"),
            paragraphItem("[x] done"),
            paragraphItem("[X] also"),
            paragraphItem("plain"),
          ],
        },
      ],
    });

    expect(promoted).toEqual({
      type: "doc",
      content: [
        {
          type: "taskList",
          content: [
            taskItem(false, "open"),
            taskItem(true, "done"),
            taskItem(true, "also"),
          ],
        },
        {
          type: "bulletList",
          content: [paragraphItem("plain")],
        },
      ],
    });
  });

  it("leaves an ordinary bullet list unchanged", () => {
    const doc = {
      type: "doc",
      content: [
        {
          type: "bulletList",
          content: [paragraphItem("plain")],
        },
      ],
    };

    expect(promoteCheckboxLists(doc)).toEqual(doc);
  });
});

describe("parseMarkdown task lists", () => {
  it("parses checkbox syntax as a task list", () => {
    const nodes = parseMarkdown("- [ ] open\n- [x] done\n", schema);

    expect(nodes).toHaveLength(1);
    expect(nodes[0]?.type.name).toBe("taskList");
    expect(nodes[0]?.child(0).type.name).toBe("taskItem");
    expect(nodes[0]?.child(0).attrs.checked).toBe(false);
    expect(nodes[0]?.child(0).textContent).toBe("open");
    expect(nodes[0]?.child(1).attrs.checked).toBe(true);
    expect(nodes[0]?.child(1).textContent).toBe("done");
  });

  it("keeps a plain bullet list", () => {
    const nodes = parseMarkdown("- plain\n", schema);

    expect(nodes[0]?.type.name).toBe("bulletList");
    expect(nodes[0]?.child(0).type.name).toBe("listItem");
  });
});

function paragraphItem(text: string) {
  return {
    type: "listItem",
    content: [
      {
        type: "paragraph",
        content: [{ type: "text", text }],
      },
    ],
  };
}

function taskItem(checked: boolean, text: string) {
  return {
    type: "taskItem",
    attrs: { checked },
    content: [
      {
        type: "paragraph",
        content: [{ type: "text", text }],
      },
    ],
  };
}
