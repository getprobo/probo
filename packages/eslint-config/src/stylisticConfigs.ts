import stylistic from "@stylistic/eslint-plugin";
import type { Linter } from "eslint";

export const stylisticConfigs = [
  stylistic.configs.customize({
    arrowParens: false, // Will actually set it to "as-needed"
    blockSpacing: true,
    braceStyle: "1tbs",
    commaDangle: "always-multiline",
    indent: 2,
    quotes: "double",
    quoteProps: "consistent-as-needed",
    semi: true,
    jsx: true,
    severity: "error",
  }),
  {
    rules: {
      "@stylistic/max-len": [
        "warn",
        { code: 120, ignoreUrls: true, ignoreStrings: true },
      ],
    },
  } satisfies Linter.Config,
];
