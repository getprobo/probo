import { createTypeScriptImportResolver } from "eslint-import-resolver-typescript";
import { importX } from "eslint-plugin-import-x";

export const importsConfigs = [
  importX.flatConfigs.recommended,
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
              pattern: "/**",
              group: "external",
              position: "after",
            },
          ],
        },
      ],
    },
  },
  {
    ...importX.flatConfigs.typescript,
    settings: {
      "import-x/resolver-next": createTypeScriptImportResolver({
        alwaysTryTypes: true,
        project: "tsconfig.json",
      }),
    },
  },
];
