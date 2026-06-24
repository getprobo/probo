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
        // Pin an explicit version: eslint-plugin-react@7.37.5 predates ESLint 10
        // and its "detect" path calls the removed context.getFilename(), which
        // crashes under ESLint 10.
        version: "19.2",
      },
    },
  },
  reactHooks.configs.flat.recommended,
];
