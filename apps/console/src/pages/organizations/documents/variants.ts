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

import { tv } from "tailwind-variants/lite";

export const documentDetailsPage = tv({
  slots: {
    root: [
      "grid grid-cols-1 items-start",
      "transition-[grid-template-columns,column-gap,row-gap] duration-150 ease-out",
      "lg:h-full lg:min-h-0 lg:items-stretch lg:overflow-hidden",
    ],
    main: "flex min-w-0 flex-col gap-4 lg:min-h-0",
    backRow: "flex items-center justify-between gap-3",
    title: "min-w-0",
    // Below lg the cards sit under the document, so the row track collapses.
    // At lg the page grid collapses the column instead.
    sideFrame: [
      "grid min-w-0 overflow-hidden",
      "transition-[grid-template-rows] duration-150 ease-out",
      "lg:h-full lg:min-h-0",
    ],
    sideClip: "min-h-0 min-w-0 overflow-hidden lg:h-full lg:min-h-0",
    side: "relative min-w-0 lg:h-full lg:min-h-0",
    viewport: [
      "w-full overscroll-y-contain outline-none lg:h-full",
      "mask-no-repeat",
      "mask-[linear-gradient(to_bottom,transparent,black_min(16px,var(--scroll-area-overflow-y-start)),black_calc(100%-min(16px,var(--scroll-area-overflow-y-end,16px))),transparent)]",
    ],
    cards: "flex w-full !min-w-0 flex-col gap-4 pr-2 [&>*]:shrink-0",
    scrollbar: [
      "flex w-1 flex-col items-center py-1",
      "pointer-events-none opacity-0 transition-opacity",
      "data-hovering:pointer-events-auto data-hovering:opacity-100",
      "data-scrolling:pointer-events-auto data-scrolling:opacity-100 data-scrolling:duration-0",
    ],
    thumb: "w-1 rounded-full bg-sand-a8",
  },
  variants: {
    cards: {
      true: {
        root: "gap-6 lg:grid-cols-[1fr_calc(280px+0.5rem)]",
        sideFrame: "max-lg:grid-rows-[1fr]",
      },
      false: {
        root: "gap-y-0 lg:grid-cols-[1fr_0px] lg:gap-x-0",
        sideFrame: "max-lg:grid-rows-[0fr]",
      },
    },
  },
  defaultVariants: {
    cards: true,
  },
});

export const documentEditor = tv({
  slots: {
    // The row shows while the editor loads, so the frame does not move and
    // the mode select stays reachable. The editor portals its controls into
    // the controls slot.
    toolbar: "flex min-h-10 w-full shrink-0 items-center gap-1 rounded-1 border border-sand-6 bg-sand-1 p-1",
    toolbarControls: "flex min-w-0 flex-1",
    toolbarEnd: "flex shrink-0 items-center gap-2",
    // The select trigger is full width by default, so its own box sizes it to
    // its content instead of the whole end group.
    toolbarMode: "shrink-0",
    // The frame only scrolls. The pages and the review column sit side by
    // side inside it so the cards move with the text.
    frame: "@container relative min-h-80 overflow-hidden lg:min-h-0 lg:flex-1",
    viewport: [
      "w-full overscroll-y-contain outline-none lg:h-full",
      "mask-no-repeat",
      "mask-[linear-gradient(to_bottom,transparent,black_min(16px,var(--scroll-area-overflow-y-start)),black_calc(100%-min(16px,var(--scroll-area-overflow-y-end,16px))),transparent)]",
    ],
    // The pages track is capped at the unscaled page width and shrinks to
    // fit. The review variant decides where it sits and whether a review
    // column follows it.
    content: [
      "grid min-h-full gap-x-(--review-gap) gap-y-4 grid-cols-[minmax(0,var(--page-width))]",
      "[--review-gap:1.5rem] [--review-min:18rem]",
    ],
    // Measured for the fit zoom, so it carries no horizontal padding.
    pages: "min-w-0 pb-6",
    stack: "relative",
    // The stack is zoomed to fit, so the border is divided by the zoom to stay
    // one screen pixel, like the toolbar's.
    page: "pointer-events-none absolute inset-x-0 border-[length:calc(1px/var(--page-scale,1))] border-sand-6 bg-sand-1",
    // The Mermaid sizes are important because rich-editor.css sits outside
    // the cascade layers. They apply on this page only.
    paper: [
      "relative",
      "[&_.mermaid-preview]:p-3! [&_.mermaid-preview]:min-h-16!",
      "[&_.mermaid-preview_svg]:max-h-80! [&_.mermaid-preview_svg]:w-auto!",
    ],
    review: "hidden min-w-0 pr-2 @4xl:block",
    scrollbar: [
      "flex w-1 flex-col items-center py-1",
      "pointer-events-none opacity-0 transition-opacity",
      "data-hovering:pointer-events-auto data-hovering:opacity-100",
      "data-scrolling:pointer-events-auto data-scrolling:opacity-100 data-scrolling:duration-0",
    ],
    thumb: "w-1 rounded-full bg-sand-a8",
  },
  variants: {
    review: {
      // The review column only fits once the frame is wide enough to keep a
      // readable line length beside it. The pages track keeps the full page
      // width as long as the review column can stay at its minimum, so the
      // review column shrinks from its cap first and the pages only scale
      // down after it.
      true: {
        content: [
          "justify-start",
          "@4xl:grid-cols-[minmax(min(var(--page-width),calc(100%_-_var(--review-min)_-_var(--review-gap))),var(--page-width))_minmax(var(--review-min),28rem)]",
        ],
      },
      false: {
        content: "justify-center",
      },
    },
    // Collaborators' carets and selections only show while editing or
    // suggesting. Hiding them in CSS keeps the editor from being rebuilt, and
    // the rules are important because rich-editor.css sits outside the
    // cascade layers.
    presence: {
      true: {},
      false: {
        paper: "[&_.presence-cursor]:hidden! [&_.presence-selection]:bg-transparent!",
      },
    },
  },
  defaultVariants: {
    review: true,
    presence: true,
  },
});

export const documentPresence = tv({
  slots: {
    stack: "flex items-center -space-x-1",
    // The sand padding keeps each avatar apart from the one it overlaps and
    // from its presence ring. The ring is an inline box-shadow, so focus uses
    // an outline.
    avatarTrigger: [
      "flex rounded-full bg-sand-1 p-0.5 outline-none",
      "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sand-8",
    ],
    overflow: [
      "inline-flex size-7 items-center justify-center rounded-full",
      "bg-sand-3 text-1 font-medium text-sand-11 ring-2 ring-sand-1",
    ],
    separator: "my-1.5",
  },
  variants: {
    active: {
      true: {},
      false: { avatarTrigger: "opacity-50 grayscale" },
    },
  },
});

export const documentProperties = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    field: "flex flex-col gap-1",
    value: "px-1.5",
  },
});

export const documentApprovals = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    title: "flex min-w-0 items-center justify-between gap-2",
    titleLead: "flex min-w-0 items-center gap-2",
    titleIcon: "size-4 shrink-0 text-sand-a9",
    rows: "flex w-full flex-col items-stretch gap-2",
    row: "flex min-w-0 items-center justify-between gap-2",
    name: "min-w-0 flex-1 truncate",
    trailing: "flex shrink-0 items-center gap-1",
  },
});

export const documentSignatories = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    title: "flex min-w-0 items-center justify-between gap-2",
    titleLead: "flex min-w-0 items-center gap-2",
    titleIcon: "size-4 shrink-0 text-sand-a9",
    rows: "flex w-full flex-col items-stretch gap-2",
    row: "flex min-w-0 items-center justify-between gap-2",
    name: "min-w-0 flex-1 truncate",
    trailing: "flex shrink-0 items-center gap-1",
  },
});

export const documentControls = tv({
  slots: {
    root: "flex w-full flex-col items-stretch gap-3",
    title: "flex min-w-0 items-center justify-between gap-2",
    names: "flex w-full flex-col items-stretch gap-2",
    name: "flex min-w-0 items-center justify-between gap-2",
    control: "min-w-0 flex-1 truncate !block",
  },
  variants: {
    pending: {
      true: { names: "opacity-60" },
      false: {},
    },
  },
  defaultVariants: {
    pending: false,
  },
});

export const documentReviewList = tv({
  slots: {
    column: "relative",
    card: "absolute inset-x-0 transition-[top] duration-150 ease-out",
  },
  variants: {
    placed: {
      true: {},
      false: { card: "invisible" },
    },
  },
  defaultVariants: {
    placed: false,
  },
});

// Interactive pieces inside a review card sit at `relative z-1`, above the
// card's stretched button.
export const documentReviewItem = tv({
  slots: {
    // The active card drops the interactive hover, so it sets the containing
    // block for the CardButton overlay itself.
    // py-2 tightens the card's p-3 vertically; Tailwind emits padding-block
    // after padding, so it wins without merging.
    root: "group/review relative py-2",
    body: "flex min-w-0 flex-col gap-3",
  },
});

export const documentReviewAuthor = tv({
  slots: {
    root: "flex min-w-0 items-center gap-2",
    identity: "flex min-w-0 flex-1 flex-col items-start",
    name: "max-w-full truncate text-left",
    // The button would inherit the card's larger line height and grow the row
    // past the name, which pulls the centered avatar off the text.
    nameButton: "max-w-full text-left text-1",
    // A line box as tall as the text pulls the time up under the name.
    time: "truncate leading-3",
    // The actions sit in the top-right corner, not centered on the two-line
    // identity.
    trailing: [
      "relative z-1 flex shrink-0 items-center gap-0.5 self-start",
      "opacity-0 transition-opacity",
      "group-hover/review:opacity-100 group-focus-within/review:opacity-100",
      "group-data-active/review:opacity-100",
      // The menu popup is portaled out of the card, so hover and focus leave
      // it while the menu is open.
      "has-data-popup-open:opacity-100",
    ],
  },
});

export const documentReviewComment = tv({
  slots: {
    root: "flex min-w-0 flex-col gap-1",
    // The read-only editor is positioned, so it would sit above the card's
    // stretched button. Clicks fall through to the card except on links.
    // rich-editor.css sizes paragraphs outside any cascade layer, so it beats
    // every utility unless the size is important.
    content: [
      "pointer-events-none [&_.tiptap_p]:text-1!",
      "[&_.tiptap>p:first-child]:mt-0 [&_.tiptap>p:last-child]:mb-0",
      "[&_a]:pointer-events-auto [&_a]:relative [&_a]:z-1",
    ],
    editor: [
      "relative z-1 cursor-text rounded-3 border border-sand-a4 bg-sand-1 px-2 py-1.5",
      "[&_.tiptap_p]:text-1!",
      "[&_.tiptap>p:first-child]:mt-0 [&_.tiptap>p:last-child]:mb-0",
    ],
    save: "relative z-1 flex justify-end gap-2",
  },
});

export const documentReviewThread = tv({
  slots: {
    root: "flex min-w-0 flex-col gap-3",
    more: "relative z-1 self-start",
  },
});

export const documentReviewReplyForm = tv({
  slots: {
    root: [
      "relative z-1 flex cursor-text flex-col gap-2",
      "rounded-3 border border-sand-a3 bg-sand-1 p-2",
    ],
    editor: [
      "min-h-8 [&_.tiptap_p]:text-1! [&_[data-rich-editor-toolbar]]:mt-0",
      "[&_.tiptap>p:first-child]:mt-0 [&_.tiptap>p:last-child]:mb-0",
    ],
    send: "ml-auto",
  },
});

export const documentSuggestionListItem = tv({
  slots: {
    changes: "flex min-w-0 flex-col gap-2",
  },
});

export const documentSuggestionChange = tv({
  slots: {
    root: "flex min-w-0 flex-col",
    text: "line-clamp-4 block break-words whitespace-pre-line",
    deleted: "line-clamp-4 block break-words whitespace-pre-line line-through",
    label: "block truncate",
    prefix: "font-bold",
  },
});

export const documentTitleField = tv({
  slots: {
    root: "min-w-0 flex-1",
    input: [
      "w-full border-0 bg-transparent p-0 text-6 font-medium text-sand-12 outline-none",
      "placeholder:text-sand-a9",
    ],
    heading: "min-w-0",
  },
});
