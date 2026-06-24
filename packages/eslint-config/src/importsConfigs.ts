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

import { createTypeScriptImportResolver } from "eslint-import-resolver-typescript";
import { importX } from "eslint-plugin-import-x";
import type { FlatConfig } from "typescript-eslint";

export const importsConfigs: FlatConfig.ConfigArray = [
  importX.flatConfigs.recommended,
  importX.flatConfigs.typescript,
  {
    rules: {
      "import-x/order": [
        "error",
        {
          "groups": ["builtin", "external", "parent", "sibling", "index"],
          "distinctGroup": true,
          "newlines-between": "always",
          "pathGroups": [
            {
              // Aliases
              pattern: "#/**",
              group: "external",
              position: "after",
            },
          ],
          "named": true,
          "alphabetize": {
            order: "asc",
            orderImportKind: "asc",
            caseInsensitive: true,
          },
        },
      ],
    },
    settings: {
      "import-x/resolver-next": createTypeScriptImportResolver({
        alwaysTryTypes: true,
        // Linting runs from the repo root, so resolve every workspace's
        // tsconfig rather than a single cwd-relative one. The `#/*` path
        // aliases live in the per-app `tsconfig.app.json`, so match all
        // tsconfig variants, not just the references-only `tsconfig.json`.
        project: ["apps/*/tsconfig*.json", "packages/*/tsconfig*.json"],
        noWarnOnMultipleProjects: true,
      }),
    },
  },
];
