import type { PropsWithChildren } from "react";
import {
  TranslatorProvider as ProboTranslatorProvider,
  type SupportedLang,
  DEFAULT_LANG,
} from "@probo/i18n";

const locales = import.meta.glob<Record<string, string>>("/locales/*.json", {
  import: "default",
});

const enPath = "/locales/en.json";

const loader = async (lang: SupportedLang): Promise<Record<string, string>> => {
  const langPath = `/locales/${lang}.json`;

  const enTranslations = locales[enPath] ? await locales[enPath]() : {};

  if (lang === DEFAULT_LANG) {
    return enTranslations;
  }

  const langTranslations = locales[langPath] ? await locales[langPath]() : {};
  return { ...enTranslations, ...langTranslations };
};

export function TranslatorProvider({ children }: PropsWithChildren) {
  return (
    <ProboTranslatorProvider loader={loader}>{children}</ProboTranslatorProvider>
  );
}
