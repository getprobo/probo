import { baseConfigs } from "./baseConfigs.ts";
import { browserLanguageOptionsConfig, nodeLanguageOptionsConfig } from "./languageOptionsConfigs.ts";
import { tsConfigs } from "./tsConfigs.ts";
import { reactConfigs } from "./reactConfigs.ts";
import { stylisticConfigs } from "./stylisticConfigs.ts";
import { importsConfigs } from "./importsConfigs.ts";

export const configs = {
  base: baseConfigs,
  imports: importsConfigs,
  ts: tsConfigs,
  languageOptions: {
    browser: browserLanguageOptionsConfig,
    node: nodeLanguageOptionsConfig,
  },
  react: reactConfigs,
  stylistic: stylisticConfigs,
};
