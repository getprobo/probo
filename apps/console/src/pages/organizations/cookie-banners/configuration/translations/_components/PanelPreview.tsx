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

interface CategoryPreview {
  name: string;
  description: string;
  isNecessary: boolean;
}

interface PanelPreviewProps {
  panelTitle: string;
  panelDescription: string;
  buttonAcceptAll: string;
  buttonRejectAll: string;
  buttonSave: string;
  categories: CategoryPreview[];
  necessaryCategoryName: string;
}

export function PanelPreview({
  panelTitle,
  panelDescription,
  buttonAcceptAll,
  buttonRejectAll,
  buttonSave,
  categories,
  necessaryCategoryName,
}: PanelPreviewProps) {
  const descriptionParts = panelDescription.split(
    "{{necessary_category}}",
  );
  const hasPlaceholder = descriptionParts.length > 1;

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
        padding: "24px 24px 12px 24px",
      }}
    >
      <p
        style={{
          fontSize: "calc(var(--probo-font-size, 14px) + 2px)",
          fontWeight: 600,
          margin: "0 0 8px",
        }}
      >
        {panelTitle}
      </p>
      <p
        style={{
          color: "var(--probo-text-secondary, #555555)",
          margin: "0 0 20px",
          fontSize: "calc(var(--probo-font-size, 14px) - 1px)",
        }}
      >
        {hasPlaceholder
          ? (
              <>
                {descriptionParts[0]}
                <strong>{necessaryCategoryName}</strong>
                {descriptionParts[1]}
              </>
            )
          : (
              panelDescription
            )}
      </p>

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {categories.map(cat => (
          <div
            key={cat.name}
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              padding: "8px 0",
              borderBottom: "1px solid var(--probo-border, #e0e0e0)",
            }}
          >
            <div style={{ display: "flex", flexDirection: "column", gap: 2, flex: 1, minWidth: 0 }}>
              <span style={{ fontWeight: 500 }}>{cat.name}</span>
              {cat.description && (
                <span
                  style={{
                    fontSize: "calc(var(--probo-font-size, 14px) - 2px)",
                    color: "var(--probo-text-secondary, #555555)",
                  }}
                >
                  {cat.description}
                </span>
              )}
            </div>
            <div
              style={{
                width: 36,
                height: 20,
                borderRadius: 10,
                background:
                  cat.isNecessary
                    ? "var(--probo-accent, #1a1a1a)"
                    : "var(--probo-border, #e0e0e0)",
                position: "relative",
                cursor: "default",
                flexShrink: 0,
                marginLeft: 12,
              }}
            >
              <div
                style={{
                  width: 16,
                  height: 16,
                  borderRadius: "50%",
                  background: "var(--probo-bg, #ffffff)",
                  position: "absolute",
                  top: 2,
                  left: cat.isNecessary ? 18 : 2,
                  transition: "left 0.2s",
                }}
              />
            </div>
          </div>
        ))}
      </div>

      <div
        style={{
          display: "flex",
          gap: 8,
          flexWrap: "wrap",
          marginTop: 20,
          paddingBottom: 12,
        }}
      >
        <button
          type="button"
          style={{
            padding: "8px 10px",
            borderRadius: "var(--probo-btn-radius, 8px)",
            border: "1px solid var(--probo-accent, #1a1a1a)",
            background: "var(--probo-accent, #1a1a1a)",
            color: "var(--probo-accent-text, #ffffff)",
            fontFamily: "inherit",
            fontSize: "var(--probo-font-size, 14px)",
            fontWeight: 500,
            lineHeight: "normal",
            cursor: "pointer",
            whiteSpace: "nowrap",
          }}
        >
          {buttonAcceptAll}
        </button>
        <button
          type="button"
          style={{
            padding: "8px 10px",
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
            whiteSpace: "nowrap",
          }}
        >
          {buttonRejectAll}
        </button>
        <button
          type="button"
          style={{
            padding: "8px 10px",
            borderRadius: "var(--probo-btn-radius, 8px)",
            border: "none",
            background: "transparent",
            color: "var(--probo-accent, #1a1a1a)",
            fontFamily: "inherit",
            fontSize: "var(--probo-font-size, 14px)",
            fontWeight: 500,
            lineHeight: "normal",
            cursor: "pointer",
            whiteSpace: "nowrap",
            textDecoration: "underline",
            flex: 1,
          }}
        >
          {buttonSave}
        </button>
      </div>
    </div>
  );
}
