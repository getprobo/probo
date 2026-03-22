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

import type { ThemeConfig } from "../headless/types";

export function themeToCSSVars(theme: ThemeConfig): string {
  return `
    --probo-primary: ${theme.primary_color};
    --probo-primary-text: ${theme.primary_text_color};
    --probo-secondary: ${theme.secondary_color};
    --probo-secondary-text: ${theme.secondary_text_color};
    --probo-bg: ${theme.background_color};
    --probo-text: ${theme.text_color};
    --probo-text-secondary: ${theme.secondary_text_body_color};
    --probo-border: ${theme.border_color};
    --probo-font: ${theme.font_family};
    --probo-radius: ${theme.border_radius}px;
  `;
}
