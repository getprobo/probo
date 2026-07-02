// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
