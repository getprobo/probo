import { baseConfigs } from "./baseConfigs.ts";
import { importsConfigs } from "./importsConfigs.ts";
import { browserLanguageOptionsConfigs, nodeLanguageOptionsConfigs } from "./languageOptionsConfigs.ts";
import { reactConfigs } from "./reactConfigs.ts";
import { stylisticConfigs } from "./stylisticConfigs.ts";
import { tsConfigs } from "./tsConfigs.ts";

export const configs = {
  base: baseConfigs,
  imports: importsConfigs,
  ts: tsConfigs,
  react: reactConfigs,
  stylistic: stylisticConfigs,
  languageOptions: {
    browser: browserLanguageOptionsConfigs,
    node: nodeLanguageOptionsConfigs,
  },
};
