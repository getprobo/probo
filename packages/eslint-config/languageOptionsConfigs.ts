import { browser, es2022, es2023, node } from "globals"

export const browserLanguageOptionsConfig = {
  languageOptions: {
    // Same as the ones we use in our tsconfig compilerOptions.lib for browser
    ecmaVersion: 2022,
    globals: {
      ...browser,
      ...es2022,
    },
    sourceType: "module",
    parserOptions: {
      ecmaFeatures: {
        impliedStrict: true,
      },
    },
  },
};

export const nodeLanguageOptionsConfig = {
  languageOptions: {
    // Same as the ones we use in our tsconfig compilerOptions.lib for node
    ecmaVersion: 2023,
    globals: {
      ...node,
      ...es2023,
    },
    sourceType: "module",
    parserOptions: {
      ecmaFeatures: {
        impliedStrict: true,
      },
    },
  },
};
