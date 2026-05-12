// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

export function withViewTransition(fn: () => void) {
    if (!document.startViewTransition) {
        fn();
        return;
    }
    document.startViewTransition(fn);
}

export function downloadFile(url: string | undefined | null, filename: string) {
    if (!url) {
        alert("Cannot download this file, fileUrl is not provided");
        return;
    }
    const link = document.createElement("a");
    link.setAttribute("href", url);
    link.setAttribute("hidden", "hidden");
    link.setAttribute("download", filename);
    link.style.display = "none";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

export function externalLinkProps(url: string): {
    href: string | undefined;
    target: "_blank";
    rel: "noopener noreferrer";
} {
    let safeHref: string | undefined;
    try {
        const parsed = new URL(url);
        if (parsed.protocol === "http:" || parsed.protocol === "https:") {
            safeHref = url;
        } else {
            console.error("Invalid URL protocol. Only HTTP and HTTPS URLs are allowed:", url);
        }
    } catch (error) {
        console.error("Invalid URL format:", url, error);
    }
    return { href: safeHref, target: "_blank", rel: "noopener noreferrer" };
}

export function safeOpenUrl(url: string) {
    try {
        const parsedUrl = new URL(url);
        if (parsedUrl.protocol === "http:" || parsedUrl.protocol === "https:") {
            window.open(url, "_blank", "noopener,noreferrer");
        } else {
            console.error(
                "Invalid URL protocol. Only HTTP and HTTPS URLs are allowed:",
                url,
            );
        }
    } catch (error) {
        console.error("Invalid URL format:", url, error);
    }
}

export function focusSiblingElement(direction = 1) {
    const current = document.activeElement as HTMLElement;

    // Selector for all focusable elements
    const focusableSelector = [
        "a[href]",
        "button:not([disabled])",
        "input:not([disabled])",
        "select:not([disabled])",
        "textarea:not([disabled])",
        '[tabindex]:not([tabindex="-1"])',
        '[contenteditable="true"]',
    ].join(", ");

    // Get all focusable elements in the document
    const focusableElements = Array.from(
        document.querySelectorAll<HTMLElement>(focusableSelector),
    ).filter((el) => {
        // Filter out elements that are not visible or have display: none
        const style = window.getComputedStyle(el);
        return style.display !== "none" && style.visibility !== "hidden";
    });

    const currentIndex = focusableElements.indexOf(current);

    let nextIndex = currentIndex + direction;

    if (nextIndex >= focusableElements.length || nextIndex < 0) {
        return null;
    }

    const nextElement = focusableElements[nextIndex];
    if (nextElement) {
        nextElement.focus();
        return nextElement;
    }

    return null;
}
