import { baseConfigs } from "#baseConfigs";
import { importsConfigs } from "#importsConfigs";
import { browserLanguageOptionsConfigs, nodeLanguageOptionsConfigs } from "#languageOptionsConfigs";
import { reactConfigs } from "#reactConfigs";
import { stylisticConfigs } from "#stylisticConfigs";
import { tsConfigs } from "#tsConfigs";

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
