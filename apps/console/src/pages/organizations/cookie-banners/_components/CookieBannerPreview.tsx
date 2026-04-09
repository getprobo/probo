import type { BannerCategory, ThemeConfig } from "@probo/cookie-banner/api";
import { defaultTheme } from "@probo/cookie-banner/api";
import { renderBanner } from "@probo/cookie-banner/banner";
import { getStrings } from "@probo/cookie-banner/i18n";
import { useEffect, useRef } from "react";

interface CookieBannerPreviewProps {
  title: string;
  description: string;
  acceptAllLabel: string;
  rejectAllLabel: string;
  savePreferencesLabel: string;
  categories: BannerCategory[];
  theme?: Partial<ThemeConfig>;
}

const noop = () => {};

export function CookieBannerPreview({
  title,
  description,
  acceptAllLabel,
  rejectAllLabel,
  savePreferencesLabel,
  categories,
  theme,
}: CookieBannerPreviewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const shadowRef = useRef<ShadowRoot | null>(null);

  useEffect(() => {
    if (containerRef.current && !shadowRef.current) {
      shadowRef.current = containerRef.current.attachShadow({ mode: "open" });
    }
  }, []);

  useEffect(() => {
    const shadow = shadowRef.current;
    if (!shadow) return;

    const resolvedTheme = { ...defaultTheme, ...theme };

    renderBanner(
      shadow,
      {
        id: "preview",
        title,
        description,
        accept_all_label: acceptAllLabel,
        reject_all_label: rejectAllLabel,
        save_preferences_label: savePreferencesLabel,
        privacy_policy_url: "",
        consent_expiry_days: 365,
        version: 1,
        categories,
        theme: resolvedTheme,
      },
      {},
      { onAcceptAll: noop, onRejectAll: noop, onCustomize: noop },
      getStrings("en"),
      { preview: true, theme: resolvedTheme },
    );

    // Override fixed positioning so the banner renders inline in the preview
    const overrideStyle = document.createElement("style");
    overrideStyle.textContent = `
      .probo-banner-overlay {
        position: relative !important;
      }
    `;
    shadow.appendChild(overrideStyle);
  }, [
    title,
    description,
    acceptAllLabel,
    rejectAllLabel,
    savePreferencesLabel,
    categories,
    theme,
  ]);

  return <div ref={containerRef} />;
}
