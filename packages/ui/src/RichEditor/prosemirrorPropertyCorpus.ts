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

export type PropertyCorpusEntry = { name: string; doc: unknown };

const MARKS = ["bold", "italic", "strike", "underline"] as const;
const WORDS = [
  "alpha",
  "beta",
  "control",
  "delta",
  "evidence",
  "finding",
  "gamma",
  "policy",
  "risk",
  "task",
  "😀",
  "e\u0301",
] as const;

function random(seed: number): () => number {
  let state = seed >>> 0;

  return () => {
    state += 0x6D2B79F5;
    let value = state;
    value = Math.imul(value ^ (value >>> 15), value | 1);
    value ^= value + Math.imul(value ^ (value >>> 7), value | 61);

    return ((value ^ (value >>> 14)) >>> 0) / 4_294_967_296;
  };
}

function integer(next: () => number, minimum: number, maximum: number): number {
  return minimum + Math.floor(next() * (maximum - minimum + 1));
}

function choose<T>(next: () => number, values: readonly T[]): T {
  return values[integer(next, 0, values.length - 1)];
}

function randomText(next: () => number): string {
  return Array.from(
    { length: integer(next, 1, 4) },
    () => choose(next, WORDS),
  ).join(" ");
}

function randomMarks(next: () => number): unknown[] | undefined {
  if (next() < 0.35) return undefined;

  if (next() < 0.15) {
    return [{ type: "code" }];
  }

  const marks: unknown[] = [];
  for (const mark of MARKS) {
    if (next() < 0.28) marks.push({ type: mark });
  }
  if (next() < 0.2) {
    marks.push({
      type: "link",
      attrs: {
        href: `https://example.com/${integer(next, 1, 99)}`,
        title: next() < 0.5 ? null : `Link ${integer(next, 1, 9)}`,
      },
    });
  }

  return marks.length > 0 ? marks : undefined;
}

function randomInlineContent(next: () => number): unknown[] {
  const content: unknown[] = [];
  const runs = integer(next, 1, 4);

  for (let index = 0; index < runs; index++) {
    const marks = randomMarks(next);
    content.push({
      type: "text",
      text: randomText(next),
      ...(marks ? { marks } : {}),
    });

    if (index < runs - 1 && next() < 0.15) {
      content.push({ type: "hardBreak" });
    }
  }

  return content;
}

function paragraph(next: () => number): unknown {
  if (next() < 0.12) return { type: "paragraph" };

  return { type: "paragraph", content: randomInlineContent(next) };
}

function list(
  next: () => number,
  type: "bulletList" | "orderedList",
  allowNested: boolean,
): unknown {
  const items = Array.from({ length: integer(next, 1, 4) }, () => {
    const content: unknown[] = [paragraph(next)];
    if (allowNested && next() < 0.3) {
      content.push(
        list(
          next,
          next() < 0.5 ? "bulletList" : "orderedList",
          false,
        ),
      );
    }

    return { type: "listItem", content };
  });

  return { type, content: items };
}

function table(next: () => number): unknown {
  const rows = integer(next, 1, 3);
  const columns = integer(next, 1, 3);

  return {
    type: "table",
    content: Array.from({ length: rows }, (_, row) => ({
      type: "tableRow",
      content: Array.from({ length: columns }, () => ({
        type: row === 0 && next() < 0.35 ? "tableHeader" : "tableCell",
        attrs: {
          colspan: 1,
          rowspan: 1,
          colwidth: null,
        },
        content: [paragraph(next)],
      })),
    })),
  };
}

function block(next: () => number): unknown {
  switch (integer(next, 0, 7)) {
    case 0:
      return paragraph(next);
    case 1:
      return {
        type: "heading",
        attrs: { level: integer(next, 1, 6) },
        content: randomInlineContent(next),
      };
    case 2:
      return {
        type: "blockquote",
        content: Array.from(
          { length: integer(next, 1, 3) },
          () => paragraph(next),
        ),
      };
    case 3:
      return {
        type: "codeBlock",
        attrs: { language: next() < 0.35 ? "mermaid" : null },
        content: [{ type: "text", text: randomText(next) }],
      };
    case 4:
      return { type: "horizontalRule" };
    case 5:
      return list(next, "bulletList", true);
    case 6:
      return list(next, "orderedList", true);
    default:
      return table(next);
  }
}

function documentForSeed(seed: number): unknown {
  const next = random(seed);

  return {
    type: "doc",
    content: Array.from(
      { length: integer(next, 1, 7) },
      () => block(next),
    ),
  };
}

export function generatedPropertyCorpus(count = 256): PropertyCorpusEntry[] {
  return Array.from({ length: count }, (_, index) => {
    const seed = (0xA017E2D5 + Math.imul(index + 1, 0x9E3779B1)) >>> 0;

    return {
      name: `property-seed-${seed.toString(16).padStart(8, "0")}`,
      doc: documentForSeed(seed),
    };
  });
}
