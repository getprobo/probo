import {
  type PropsWithChildren,
  useState,
  createContext,
  useContext,
  useMemo,
} from "react";
import { TranslatorProvider as ProboTranslatorProvider } from "@probo/i18n";

const loader = (lang: string) => {
  return fetch(`/locales/${lang}.json`).then((res) => res.json());
};

type Lang = "en" | "fr";

type LangContextType = {
  lang: Lang;
  setLang: (lang: Lang) => void;
};

const LangContext = createContext<LangContextType>({
  lang: "en",
  setLang: () => {},
});

export function useLang() {
  return useContext(LangContext);
}

export function LangProvider({ children }: PropsWithChildren) {
  const [lang, setLang] = useState<Lang>("en");

  const value = useMemo(() => ({ lang, setLang }), [lang]);

  return <LangContext.Provider value={value}>{children}</LangContext.Provider>;
}

/**
 * Provider for the translator
 */
export function TranslatorProvider({ children }: PropsWithChildren) {
  const { lang } = useLang();
  return (
    <ProboTranslatorProvider lang={lang} loader={loader}>
      {children}
    </ProboTranslatorProvider>
  );
}
