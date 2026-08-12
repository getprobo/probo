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

import type { BannerConfig, BannerLayout, BannerText, Presentation } from "./types";

// Strict, GDPR-safe layout used only as a defensive fallback: a config without a
// `layout` means the (self-hosted) Probo backend predates probod v0.246.0, the
// first release that sends it. We keep the banner functional and compliant
// rather than crashing, and surface the mismatch so operators know to update.
const STRICT_LAYOUT: BannerLayout = {
  presentation: "OPT_IN",
  initial_state: "banner",
  reopen_state: "panel",
  default_non_necessary_granted: false,
  buttons: { accept_all: true, reject_all: true, customize: true, save: true },
  settings_link: "default",
};

let warnedLayout = false;

// resolveLayout returns the presentation policy the server sends. The SDK and
// backend are released together, so `layout` is always present in practice; the
// only way it can be missing is a self-hosted Probo backend older than
// v0.246.0, which we flag once and treat as strict opt-in.
export function resolveLayout(config: BannerConfig): BannerLayout {
  if (config.layout) {
    return config.layout;
  }

  if (!warnedLayout) {
    warnedLayout = true;
    console.error(
      "[probo] banner config has no `layout`: your self-hosted Probo backend (probod) is older than this SDK. Update probod to v0.246.0 or later. Falling back to strict opt-in.",
    );
  }

  return STRICT_LAYOUT;
}

// The banner-card text keys each presentation uses. Every presentation falls
// back to the base keys when a customer translation omits the presentation-
// specific wording.
const PRESENTATION_TEXT_KEYS: Record<
  Presentation,
  { title: string; description: string; primary: string; secondary?: string }
> = {
  OPT_IN: {
    title: "banner_title",
    description: "banner_description",
    primary: "button_accept_all",
    secondary: "button_reject_all",
  },
  OPT_OUT: {
    title: "banner_title_opt_out",
    description: "banner_description_opt_out",
    primary: "button_acknowledge",
    secondary: "button_opt_out",
  },
  NOTICE: {
    title: "banner_title_notice",
    description: "banner_description_notice",
    primary: "button_dismiss",
  },
};

// resolveBannerText builds the typed banner-card wording for the active
// presentation from the flat text keys, so callers never touch presentation-
// specific key names. The themed renderers use their own keys directly; this is
// the resolver headless integrators use to render the correct copy.
export function resolveBannerText(config: BannerConfig): BannerText {
  const presentation = resolveLayout(config).presentation;
  const keys = PRESENTATION_TEXT_KEYS[presentation] ?? PRESENTATION_TEXT_KEYS.OPT_IN;
  const texts = config.texts ?? {};

  const pick = (variantKey: string, baseKey: string): string =>
    texts[variantKey] || texts[baseKey] || "";

  const result: BannerText = {
    title: pick(keys.title, "banner_title"),
    description: pick(keys.description, "banner_description"),
    primaryButton: pick(keys.primary, "button_accept_all"),
  };

  if (keys.secondary) {
    const secondary = pick(keys.secondary, "button_reject_all");
    if (secondary) {
      result.secondaryButton = secondary;
    }
  }

  return result;
}
