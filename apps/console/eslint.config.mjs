import { defineConfig } from "eslint/config";
import { configs } from "@probo/eslint-config";

export default defineConfig([
  ...configs.base,
  ...configs.ts,
  ...configs.react,
  configs.languageOptions.browser,
  ...configs.stylistic,
  {
    languageOptions: {
      parserOptions: {
        tsConfigRootDir: import.meta.dirname,
      },
    },
  },
]);
