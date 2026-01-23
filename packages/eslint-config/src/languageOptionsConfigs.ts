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
