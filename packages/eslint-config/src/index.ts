import { baseConfigs } from "./baseConfigs";
import { browserLanguageOptionsConfig, nodeLanguageOptionsConfig } from "./languageOptionsConfigs";
import { tsConfigs } from "./tsConfigs";
import { reactConfigs } from "./reactConfigs";
import { stylisticConfigs } from "./stylisticConfigs";

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
