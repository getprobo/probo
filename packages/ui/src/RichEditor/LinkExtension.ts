// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { Link } from "@tiptap/extension-link";
import { Plugin, PluginKey } from "@tiptap/pm/state";

export type LinkEditorRequest = {
  id: number;
  href: string;
};

export const linkEditorPluginKey = new PluginKey<LinkEditorRequest | null>(
  "linkEditor",
);

function isAbsoluteHref(href: string): boolean {
  const scheme = /^[a-z][a-z0-9+.-]*:/i.exec(href)?.[0];
  // host:port matches the URI-scheme grammar because "." is allowed in schemes.
  return Boolean(scheme && !/^\d/.test(href.slice(scheme.length)));
}

function isHostLikeHref(href: string): boolean {
  if (
    href === ""
    || href.startsWith("/")
    || href.startsWith("#")
    || href.startsWith("?")
    || href.startsWith(".")
    || isAbsoluteHref(href)
  ) {
    return false;
  }

  const host = href.split("/")[0] ?? "";
  return host.includes(".")
    || host.toLowerCase() === "localhost"
    || /:\d+$/.test(host);
}

function hrefWithScheme(href: string): string {
  return isHostLikeHref(href) ? `https://${href}` : href;
}

function storedLinkHref(anchor: HTMLAnchorElement): string | undefined {
  return anchor.getAttribute("href")?.trim() || undefined;
}

function openableLinkHref(anchor: HTMLAnchorElement): string | undefined {
  const raw = storedLinkHref(anchor);
  if (!raw) {
    return undefined;
  }

  if (isHostLikeHref(raw)) {
    return hrefWithScheme(raw);
  }

  return anchor.href;
}

function hoveredAnchor(event: MouseEvent): HTMLAnchorElement | null {
  const anchor = (event.target as Element | null)?.closest("a");

  return anchor instanceof HTMLAnchorElement ? anchor : null;
}

export function openLink(href: string) {
  window.open(hrefWithScheme(href), "_blank", "noopener,noreferrer");
}

export const LinkExtension = Link.extend({
  addProseMirrorPlugins() {
    const { editor } = this;
    let editingOnPointerDown = false;

    return [
      ...(this.parent?.() ?? []),
      new Plugin<LinkEditorRequest | null>({
        key: linkEditorPluginKey,
        state: {
          init: () => null,
          apply: (tr, request) => {
            const href = tr.getMeta(linkEditorPluginKey) as string | undefined;

            if (href === undefined) {
              return request;
            }

            return { id: (request?.id ?? 0) + 1, href };
          },
        },
        props: {
          handleDOMEvents: {
            mousedown: (view, event) => {
              editingOnPointerDown = view.hasFocus();

              if (
                event.button !== 0
                || editingOnPointerDown
                || !hoveredAnchor(event)
              ) {
                return false;
              }

              // Not editing: swallow the event so following a link never moves
              // the caret into the content. ProseMirror then skips its own
              // click handling, so the link is opened from `click` below.
              event.preventDefault();
              return true;
            },
            click: (_view, event) => {
              const anchor = hoveredAnchor(event);

              if (editingOnPointerDown || !anchor) {
                return false;
              }

              event.preventDefault();
              if (event.detail > 1) {
                return true;
              }

              openLink(openableLinkHref(anchor) ?? anchor.href);
              return true;
            },
          },
          handleClick: (_view, pos, event) => {
            const anchor = hoveredAnchor(event);
            if (!anchor) {
              return false;
            }

            const href = openableLinkHref(anchor);
            if (event.ctrlKey || event.metaKey) {
              if (!href) {
                return false;
              }

              if (event.detail > 1) {
                return true;
              }

              openLink(href);
              return true;
            }

            const stored = storedLinkHref(anchor);
            if (!stored) {
              return false;
            }

            editor
              .chain()
              .setTextSelection(pos)
              .extendMarkRange("link")
              .command(({ tr }) => {
                tr.setMeta(linkEditorPluginKey, stored);
                return true;
              })
              .run();

            return true;
          },
        },
      }),
    ];
  },
}).configure({
  openOnClick: false,
});
