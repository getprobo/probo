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
