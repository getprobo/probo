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

function LockIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
      <path d="M7 11V7a5 5 0 0 1 10 0v4" />
    </svg>
  );
}

export function PlaceholderPreview({
  placeholderText,
  placeholderButton,
  categoryName,
}: PlaceholderPreviewProps) {
  const parts = placeholderText.split("{{category}}");
  const hasPlaceholder = parts.length > 1;

  return (
    <div
      style={{
        background: "var(--probo-bg, #ffffff)",
        color: "var(--probo-text-secondary, #555555)",
        borderRadius: "var(--probo-radius, 12px)",
        border: "1px dashed var(--probo-border, #e0e0e0)",
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
        gap: 12,
        textAlign: "center",
        minHeight: 120,
        boxSizing: "border-box",
      }}
    >
      <span style={{ color: "var(--probo-text-secondary, #555555)" }}>
        <LockIcon />
      </span>
      <p style={{ margin: 0 }}>
        {hasPlaceholder
          ? (
              <>
                {parts[0]}
                <strong>{categoryName}</strong>
                {parts[1]}
              </>
            )
          : placeholderText.replace("{{category}}", categoryName)}
      </p>
      <button
        type="button"
        style={{
          background: "none",
          border: "none",
          color: "var(--probo-accent, #1a1a1a)",
          textDecoration: "underline",
          cursor: "pointer",
          fontFamily: "inherit",
          fontSize: "inherit",
          padding: 0,
        }}
      >
        {placeholderButton}
      </button>
    </div>
  );
}
