import { defineConfig } from "eslint/config";

import { baseConfigs } from "#baseConfigs";
import { importsConfigs } from "#importsConfigs";
import { nodeLanguageOptionsConfigs } from "#languageOptionsConfigs";
import { stylisticConfigs } from "#stylisticConfigs";
import { tsConfigs } from "#tsConfigs";

export default defineConfig([
  ...baseConfigs,
  ...tsConfigs,
  ...importsConfigs,
  ...stylisticConfigs,
  ...nodeLanguageOptionsConfigs,
]);
