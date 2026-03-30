// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useEffect } from "react";

export function useFavicon(faviconUrl?: string | null) {
  useEffect(() => {
    if (!faviconUrl) return;

    let favicon: HTMLLinkElement;
    const existingFavicon = document.getElementById("favicon") as (HTMLLinkElement | null);

    if (existingFavicon) {
      favicon = existingFavicon
      favicon.href = faviconUrl;
    } else {
      favicon = document.createElement("link");
      favicon.id = "favicon";
      favicon.rel = "icon";
      favicon.href = faviconUrl;
      document.head.appendChild(favicon);
    }

    return () => {
      favicon.href = "/favicons/favicon.ico";
    }
  });
}