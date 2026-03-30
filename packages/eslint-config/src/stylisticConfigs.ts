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

import stylistic from "@stylistic/eslint-plugin";
import type { FlatConfig } from "typescript-eslint";

export const stylisticConfigs: FlatConfig.ConfigArray = [
  stylistic.configs.customize({
    arrowParens: false, // Will actually set it to "as-needed"
    blockSpacing: true,
    braceStyle: "1tbs",
    commaDangle: "always-multiline",
    indent: 2,
    quotes: "double",
    quoteProps: "consistent-as-needed",
    semi: true,
    jsx: true,
    severity: "error",
  }),
  {
    rules: {
      "@stylistic/max-len": [
        "warn",
        {
          code: 120,
          tabWidth: 2,
          ignoreStrings: true,
          ignoreTemplateLiterals: true,
          ignoreComments: true,
          ignoreUrls: true,
        },
      ],
    },
  },
];
