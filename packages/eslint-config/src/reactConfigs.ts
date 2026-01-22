import react from "eslint-plugin-react";
import * as reactHooks from "eslint-plugin-react-hooks";

export const reactConfigs = [
  {
    ...react.configs.flat.recommended,
    ...react.configs.flat["jsx-runtime"],
    ...reactHooks.configs.flat.recommended,
    settings: {
      react: {
        version: "detect",
      },
    },
  },
];
