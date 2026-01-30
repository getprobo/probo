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