import { defineConfig } from "eslint/config";
import { configs } from "@probo/eslint-config";

export default defineConfig([
  ...configs.base,
  ...configs.ts,
  ...configs.react,
  ...configs.stylistic,
  {
    ignores: ["./tailwind.config.js"],
    ...configs.languageOptions.browser,
  },
  {
    files: ["./tailwind.config.js"],
    languageOptions: {
      ...configs.languageOptions.node.languageOptions,
      sourceType: "commonjs",
    },
  },
  {
    languageOptions: {
      parserOptions: {
        tsConfigRootDir: import.meta.dirname,
      },
    },
  },
]);
