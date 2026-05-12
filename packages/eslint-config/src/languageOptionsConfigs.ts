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

import globals from "globals";
import type { FlatConfig } from "typescript-eslint";

export const browserLanguageOptionsConfigs: FlatConfig.ConfigArray = [
  {
    languageOptions: {
      // Same as the ones we use in our tsconfig compilerOptions.lib for browser
      ecmaVersion: 2022,
      globals: {
        ...globals.browser,
        ...globals.es2022,
      },
      sourceType: "module",
      parserOptions: {
        projectService: true,
        ecmaFeatures: {
          impliedStrict: true,
        },
      },
    },
  },
];

export const nodeLanguageOptionsConfigs: FlatConfig.ConfigArray = [
  {
    languageOptions: {
      // Same as the ones we use in our tsconfig compilerOptions.lib for node
      ecmaVersion: 2023,
      globals: {
        ...globals.node,
        ...globals.es2023,
      },
      sourceType: "module",
      parserOptions: {
        projectService: true,
        ecmaFeatures: {
          impliedStrict: true,
        },
      },
    },
  },
];
