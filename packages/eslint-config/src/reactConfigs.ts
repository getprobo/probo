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

import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import type { FlatConfig } from "typescript-eslint";

export const reactConfigs: FlatConfig.ConfigArray = [
  react.configs.flat.recommended,
  react.configs.flat["jsx-runtime"],
  {
    rules: {
      ...react.configs.flat.recommended.rules,
      ...react.configs.flat["jsx-runtime"].rules,
      // onScrollEnd is a valid React 19 event handler, but eslint-plugin-react
      // doesn't support it yet. See: https://github.com/jsx-eslint/eslint-plugin-react/pull/3958
      "react/no-unknown-property": ["error", { ignore: ["onScrollEnd"] }],
    },
  },
  {
    settings: {
      react: {
        version: "detect",
      },
    },
  },
  reactHooks.configs.flat.recommended,
];
