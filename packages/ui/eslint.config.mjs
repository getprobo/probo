import { configs } from "@probo/eslint-config";
import { defineConfig } from "eslint/config";

export default defineConfig([
  configs.base,
  configs.ts,
  configs.imports,
  configs.react,
  configs.stylistic,
  {
    extends: [configs.languageOptions.browser],
    ignores: ["./tailwind.config.js"],
  },
  {
    extends: [configs.languageOptions.node],
    files: ["./tailwind.config.js"],
    languageOptions: {
      sourceType: "commonjs",
    },
  },
]);
