import { baseConfigs } from "./baseConfigs.ts";
import { browserLanguageOptionsConfig, nodeLanguageOptionsConfig } from "./languageOptionsConfigs.ts";
import { tsConfigs } from "./tsConfigs.ts";
import { reactConfigs } from "./reactConfigs.ts";
import { stylisticConfigs } from "./stylisticConfigs.ts";

export const configs = {
  base: baseConfigs,
  ts: tsConfigs,
  languageOptions: {
    browser: browserLanguageOptionsConfig,
    node: nodeLanguageOptionsConfig,
  },
  react: reactConfigs,
  stylistic: stylisticConfigs,
};
