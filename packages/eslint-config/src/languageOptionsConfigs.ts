import type { Linter } from "eslint";
import * as globals from "globals";

export const browserLanguageOptionsConfig = {
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
} satisfies Pick<Linter.Config, "languageOptions">;

export const nodeLanguageOptionsConfig = {
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
} satisfies Pick<Linter.Config, "languageOptions">;
