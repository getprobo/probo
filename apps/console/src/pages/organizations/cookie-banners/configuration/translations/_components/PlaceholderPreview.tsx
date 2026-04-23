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

interface PlaceholderPreviewProps {
  placeholderText: string;
  placeholderButton: string;
  categoryName: string;
}

export function PlaceholderPreview({
  placeholderText,
  placeholderButton,
  categoryName,
}: PlaceholderPreviewProps) {
  const displayText = placeholderText.replace("{{category}}", categoryName);

  return (
    <div
      style={{
        background: "var(--probo-bg, #ffffff)",
        color: "var(--probo-text, #1a1a1a)",
        borderRadius: "var(--probo-radius, 12px)",
        boxShadow:
          "var(--probo-shadow, 0 4px 24px rgba(0, 0, 0, 0.12))",
        fontFamily:
          "var(--probo-font-family, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif)",
        fontSize: "var(--probo-font-size, 14px)",
        lineHeight: 1.5,
        maxWidth: 380,
        width: "100%",
        padding: "24px",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: 16,
        textAlign: "center",
      }}
    >
      <div
        style={{
          width: 48,
          height: 48,
          borderRadius: "50%",
          background:
            "color-mix(in srgb, var(--probo-text, #1a1a1a) 8%, var(--probo-bg, #ffffff))",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontSize: 20,
        }}
      >
        🍪
      </div>
      <p
        style={{
          color: "var(--probo-text-secondary, #555555)",
          margin: 0,
        }}
      >
        {displayText}
      </p>
      <button
        type="button"
        style={{
          padding: "8px 16px",
          borderRadius: "var(--probo-btn-radius, 8px)",
          border: "1px solid var(--probo-border, #e0e0e0)",
          background:
            "color-mix(in srgb, var(--probo-text, #1a1a1a) 8%, var(--probo-bg, #ffffff))",
          color: "var(--probo-text, #1a1a1a)",
          fontFamily: "inherit",
          fontSize: "var(--probo-font-size, 14px)",
          fontWeight: 500,
          lineHeight: "normal",
          cursor: "pointer",
        }}
      >
        {placeholderButton}
      </button>
    </div>
  );
}
