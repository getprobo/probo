import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";

const defaultValue = {
  lang: "en" as "en" | "fr",
  translations: {} as Record<string, string>,
  translate: (s: string) => s,
};

type Context = typeof defaultValue;

const TranslatorContext = createContext(defaultValue);

type Props = {
  lang: "en" | "fr";
  loader: (lang: string) => Promise<Record<string, string>>;
};

export function TranslatorProvider({
  lang,
  loader,
  children,
}: PropsWithChildren<Props>) {
  const [translations, setTranslations] = useState(
    {} as Record<string, string>
  );
  const translate = useCallback<Context["translate"]>(
    (s) => {
      return translations[s] ? translations[s] : s;
    },
    [translations]
  );

  useEffect(() => {
    loader(lang).then(setTranslations);
  }, [lang]);

  return (
    <TranslatorContext.Provider value={{ lang, translations, translate }}>
      {children}
    </TranslatorContext.Provider>
  );
}

export function useTranslate() {
  const { translate, lang } = useContext(TranslatorContext);
  return {
    lang,
    __: translate,

    dateFormat: (
      date: Date | string | null | undefined,
      options: Intl.DateTimeFormatOptions = {
        year: "numeric",
        month: "short",
        day: "numeric",
        weekday: "short",
      }
    ) => {
      if (!date) {
        return "";
      }
      if (typeof date === "string") {
        return new Intl.DateTimeFormat(lang, options).format(parseDate(date));
      }
      return new Intl.DateTimeFormat(lang, options).format(date);
    },
  };
}

function parseDate(date: Date | string): Date {
  if (typeof date === "string") {
    if (date.includes("T")) {
      return new Date(date);
    }
    const parts = date.split("-");
    return new Date(
      parseInt(parts[0], 10),
      parts[1] ? parseInt(parts[1], 10) - 1 : 0,
      parts[2] ? parseInt(parts[2], 10) : 1
    );
  }
  return date;
}
