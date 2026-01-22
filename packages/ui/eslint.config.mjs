import { configs } from "@probo/eslint-config";
import { defineConfig } from "eslint/config";

export default defineConfig([
  ...configs.base,
  ...configs.ts,
  ...configs.imports,
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
]);
