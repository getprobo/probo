import { defineConfig } from "eslint/config";

import { baseConfigs } from "./src/baseConfigs.ts";
import { tsConfigs } from "./src/tsConfigs.ts";
import { nodeLanguageOptionsConfig } from "./src/languageOptionsConfigs.ts";
import { stylisticConfigs } from "./src/stylisticConfigs.ts";

export default defineConfig([
  ...baseConfigs,
  ...tsConfigs,
  nodeLanguageOptionsConfig,
  ...stylisticConfigs,
  {
    languageOptions: {
      parserOptions: {
        tsConfigRootDir: import.meta.dirname,
      },
    },
  },
]);
