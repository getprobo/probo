import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import type { FlatConfig } from "typescript-eslint";

export const reactConfigs: FlatConfig.ConfigArray = [
  {
    ...react.configs.flat.recommended,
    ...react.configs.flat["jsx-runtime"],
    settings: {
      react: {
        version: "detect",
      },
    },
  },
  reactHooks.configs.flat.recommended,
];
