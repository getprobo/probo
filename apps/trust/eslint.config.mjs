import { defineConfig } from "eslint/config";
import { configs } from "@probo/eslint-config";
// import { createTypeScriptImportResolver } from "eslint-import-resolver-typescript";
// import { importX } from "eslint-plugin-import-x";

export default defineConfig([
  ...configs.base,
  ...configs.ts,
  ...configs.imports,
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
