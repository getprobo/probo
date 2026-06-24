// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import type { StorybookConfig } from "@storybook/react-vite";

// The v2 kit is theme-isolated from v1 (v2 wipes Tailwind's palette), so it has
// its own Storybook build scanning only src/v2 and loading the v2 theme.
const config: StorybookConfig = {
  stories: ["../src/v2/**/*.stories.@(js|jsx|mjs|ts|tsx)"],
  addons: [import.meta.resolve("@chromatic-com/storybook")],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
};
export default config;
