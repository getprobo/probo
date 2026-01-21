import { defineConfig } from "eslint/config";

import { baseConfigs } from "#baseConfigs";
import { tsConfigs } from "#tsConfigs";
import { nodeLanguageOptionsConfig } from "#languageOptionsConfigs";
import { stylisticConfigs } from "#stylisticConfigs";

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
