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

import { Extension } from "@tiptap/core";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import { type EditorState, Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet, type EditorView } from "@tiptap/pm/view";

export type PageLayout = {
  pageHeight: number;
  marginTop: number;
  marginBottom: number;
  gap: number;
};

// A block in document order. `top` is measured from the top of the first
// page's content area as if there were no page breaks.
export type PageBlock = {
  pos: number;
  top: number;
  height: number;
};

// A break pushes the block at `pos` down by `spacer` pixels so it starts at
// the top of the next page's content area.
export type PageBreak = {
  pos: number;
  spacer: number;
};

export type PaginationOptions = {
  layout: PageLayout;
  onPages?: (pages: number) => void;
};

const listTypes = new Set(["bulletList", "orderedList", "taskList"]);

// The extension does not change heights itself, but a spacer can stop two
// block margins from collapsing. A difference this small is not a new layout.
const breakTolerance = 2;

const paginationKey = new PluginKey<PaginationState>("document-pagination");

type PaginationState = {
  breaks: PageBreak[];
  decorations: DecorationSet;
};

// computePageBreaks lays blocks out on pages. A block that would cross the
// bottom of a page starts on the next one, unless it already starts at the
// top of a page: a block taller than a page keeps running past the end.
export function computePageBreaks(
  blocks: ReadonlyArray<PageBlock>,
  layout: PageLayout,
): { breaks: PageBreak[]; pages: number } {
  const contentHeight = layout.pageHeight - layout.marginTop - layout.marginBottom;
  const stride = layout.pageHeight + layout.gap;
  const breaks: PageBreak[] = [];
  let offset = 0;
  let pageStart = 0;
  let pages = 1;
  let bottom = 0;

  for (const block of blocks) {
    let top = block.top + offset;
    while (top >= pageStart + contentHeight) {
      pageStart += stride;
      pages += 1;
    }
    if (top < pageStart) {
      // A tall block ended in the margin or gap before this page.
      const spacer = pageStart - top;
      breaks.push({ pos: block.pos, spacer });
      offset += spacer;
      top = pageStart;
    } else if (top > pageStart && top + block.height > pageStart + contentHeight) {
      const spacer = pageStart + stride - top;
      breaks.push({ pos: block.pos, spacer });
      offset += spacer;
      pageStart += stride;
      pages += 1;
      top = pageStart;
    }
    bottom = Math.max(bottom, top + block.height);
  }

  while (bottom > pageStart + contentHeight) {
    pageStart += stride;
    pages += 1;
  }

  return { breaks, pages };
}

// createPaginationExtension lays the editor out on pages with spacer widgets
// at each break. The spacers are decorations, so the document never changes.
export function createPaginationExtension(options: PaginationOptions) {
  return Extension.create({
    name: "documentPagination",
    addProseMirrorPlugins() {
      return [
        new Plugin<PaginationState>({
          key: paginationKey,
          state: {
            init: () => ({ breaks: [], decorations: DecorationSet.empty }),
            apply(transaction, value, _oldState, newState) {
              const next = transaction.getMeta(paginationKey) as PageBreak[] | undefined;
              if (next) {
                return { breaks: next, decorations: spacerDecorations(newState.doc, next) };
              }
              if (!transaction.docChanged) return value;
              return {
                breaks: value.breaks.map(item => ({
                  pos: transaction.mapping.map(item.pos),
                  spacer: item.spacer,
                })),
                decorations: value.decorations.map(transaction.mapping, transaction.doc),
              };
            },
          },
          props: {
            decorations(state) {
              return paginationKey.getState(state)?.decorations;
            },
          },
          view: view => new PaginationView(view, options),
        }),
      ];
    },
  });
}

class PaginationView {
  private frame = 0;
  private readonly view: EditorView;
  private readonly options: PaginationOptions;
  private readonly observer: ResizeObserver;

  constructor(view: EditorView, options: PaginationOptions) {
    this.view = view;
    this.options = options;
    this.observer = new ResizeObserver(() => this.schedule());
    this.observer.observe(view.dom);
    this.schedule();
  }

  update(view: EditorView, previous: EditorState) {
    if (view.state.doc !== previous.doc) this.schedule();
  }

  destroy() {
    if (this.frame !== 0) window.cancelAnimationFrame(this.frame);
    this.observer.disconnect();
  }

  private schedule() {
    if (this.frame !== 0) return;
    this.frame = window.requestAnimationFrame(() => {
      this.frame = 0;
      this.measure();
    });
  }

  private measure() {
    if (!this.view.dom.isConnected) return;
    const current = paginationKey.getState(this.view.state)?.breaks ?? [];
    const blocks = measureBlocks(this.view, current);
    const { breaks, pages } = computePageBreaks(blocks, this.options.layout);
    this.options.onPages?.(pages);
    if (sameBreaks(current, breaks)) return;
    this.view.dispatch(
      this.view.state.tr
        .setMeta(paginationKey, breaks)
        .setMeta("addToHistory", false),
    );
  }
}

// Breaks fall between top-level blocks and between list items, so a long list
// can continue on the next page.
function measureBlocks(view: EditorView, breaks: ReadonlyArray<PageBreak>): PageBlock[] {
  const origin = layoutTop(view.dom);
  const candidates: number[] = [];
  view.state.doc.forEach((node, offset) => {
    if (listTypes.has(node.type.name)) {
      node.forEach((_item, itemOffset) => {
        candidates.push(offset + 1 + itemOffset);
      });
      return;
    }
    candidates.push(offset);
  });

  const blocks: PageBlock[] = [];
  for (const pos of candidates) {
    const dom = view.nodeDOM(pos);
    if (!(dom instanceof HTMLElement)) continue;
    const spacersAbove = breaks.reduce((sum, item) => (item.pos <= pos ? sum + item.spacer : sum), 0);
    blocks.push({
      pos,
      top: layoutTop(dom) - origin - spacersAbove,
      height: dom.offsetHeight,
    });
  }
  return blocks;
}

// Offsets are layout pixels, so CSS zoom on an ancestor does not skew them.
function layoutTop(element: HTMLElement): number {
  let top = 0;
  let current: HTMLElement | null = element;
  while (current) {
    top += current.offsetTop;
    current = current.offsetParent instanceof HTMLElement ? current.offsetParent : null;
  }
  return top;
}

function spacerDecorations(doc: ProseMirrorNode, breaks: ReadonlyArray<PageBreak>): DecorationSet {
  return DecorationSet.create(doc, breaks.map(item => Decoration.widget(item.pos, () => {
    const spacer = document.createElement("div");
    spacer.className = "pagination-spacer";
    spacer.setAttribute("aria-hidden", "true");
    spacer.contentEditable = "false";
    spacer.style.height = `${item.spacer}px`;
    spacer.style.pointerEvents = "none";
    return spacer;
  }, {
    side: -1,
    ignoreSelection: true,
    key: `page-break-${item.pos}-${Math.round(item.spacer)}`,
  })));
}

function sameBreaks(current: ReadonlyArray<PageBreak>, next: ReadonlyArray<PageBreak>): boolean {
  if (current.length !== next.length) return false;
  return current.every((item, index) =>
    item.pos === next[index].pos
    && Math.abs(item.spacer - next[index].spacer) <= breakTolerance,
  );
}
