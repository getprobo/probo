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
          ? parts.map((part, i) => (
              <span key={i}>
                {part}
                {i < parts.length - 1 && <strong>{categoryName}</strong>}
              </span>
            ))
          : placeholderText}
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
