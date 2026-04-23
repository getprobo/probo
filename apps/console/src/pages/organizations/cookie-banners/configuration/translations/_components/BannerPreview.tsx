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

import { Logo } from "@probo/ui";

interface BannerPreviewProps {
  bannerTitle: string;
  bannerDescription: string;
  buttonAcceptAll: string;
  buttonRejectAll: string;
  buttonCustomize: string;
  privacyPolicyLinkText: string;
  showBranding: boolean;
}

function interpolateDescription(
  description: string,
  linkText: string,
): string {
  return description.replaceAll(
    "{{privacy_policy_link}}",
    linkText,
  );
}

export function BannerPreview({
  bannerTitle,
  bannerDescription,
  buttonAcceptAll,
  buttonRejectAll,
  buttonCustomize,
  privacyPolicyLinkText,
  showBranding,
}: BannerPreviewProps) {
  const descriptionParts = bannerDescription.split("{{privacy_policy_link}}");
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
        maxWidth: 450,
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
        {bannerTitle}
      </p>
      <p
        style={{
          color: "var(--probo-text-secondary, #555555)",
          margin: "0 0 20px",
        }}
      >
        {hasPlaceholder
          ? descriptionParts.map((part, i) => (
              <span key={i}>
                {part}
                {i < descriptionParts.length - 1 && (
                  <a
                    href="#"
                    onClick={e => e.preventDefault()}
                    style={{
                      color: "var(--probo-accent, #1a1a1a)",
                      textDecoration: "underline",
                    }}
                  >
                    {privacyPolicyLinkText}
                  </a>
                )}
              </span>
            ))
          : (
              interpolateDescription(bannerDescription, privacyPolicyLinkText)
            )}
      </p>
      <div
        style={{
          display: "flex",
          gap: 8,
          flexWrap: "wrap",
          paddingBottom: "12px",
        }}
      >
        <span>
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
        </span>
        <span>
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
        </span>
        <span>
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
            }}
          >
            {buttonCustomize}
          </button>
        </span>
      </div>
      {showBranding && (
        <div
          style={{
            textAlign: "center",
            fontSize: "calc(var(--probo-font-size, 14px) - 2px)",
            fontWeight: 400,
            color: "var(--probo-text-secondary, #555555)",
          }}
        >
          Privacy by
          {" "}
          <Logo withPicto className="inline h-3.5 align-[-3px]" />
        </div>
      )}
    </div>
  );
}
