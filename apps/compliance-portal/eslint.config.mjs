import { configs } from "@probo/eslint-config";
import { defineConfig } from "eslint/config";

export default defineConfig([
  configs.base,
  configs.ts,
  configs.imports,
  configs.react,
  configs.relay,
  configs.stylistic,
  configs.languageOptions.browser,
]);
