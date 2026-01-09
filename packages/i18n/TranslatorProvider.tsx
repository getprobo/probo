import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";

export const SUPPORTED_LANGUAGES = ["en", "fr"] as const;
export type SupportedLang = (typeof SUPPORTED_LANGUAGES)[number];
export const DEFAULT_LANG: SupportedLang = "en";
const STORAGE_KEY = "probo-lang";

type TranslatorContextValue = {
  lang: SupportedLang;
  translations: Record<string, string>;
  translate: (s: string) => string;
  setLang: (lang: SupportedLang) => void;
};

const defaultValue: TranslatorContextValue = {
  lang: DEFAULT_LANG,
  translations: {},
  translate: (s: string) => s,
  setLang: () => {},
};

const TranslatorContext = createContext<TranslatorContextValue>(defaultValue);

function detectLanguage(): SupportedLang {
  if (typeof window === "undefined") {
    return DEFAULT_LANG;
  }

  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored && SUPPORTED_LANGUAGES.includes(stored as SupportedLang)) {
    return stored as SupportedLang;
  }

  const browserLang = navigator.language.split("-")[0];
  if (SUPPORTED_LANGUAGES.includes(browserLang as SupportedLang)) {
    return browserLang as SupportedLang;
  }

  return DEFAULT_LANG;
}

type Props = {
  loader: (lang: SupportedLang) => Promise<Record<string, string>>;
  defaultLang?: SupportedLang;
};

export function TranslatorProvider({
  loader,
  defaultLang,
  children,
}: PropsWithChildren<Props>) {
  const [lang, setLangState] = useState<SupportedLang>(() => {
    return defaultLang ?? detectLanguage();
  });

  const [translations, setTranslations] = useState<Record<string, string>>({});

  const translate = useCallback(
    (s: string): string => {
      return translations[s] ?? s;
    },
    [translations]
  );

  const setLang = useCallback((newLang: SupportedLang) => {
    if (SUPPORTED_LANGUAGES.includes(newLang)) {
      setLangState(newLang);
      if (typeof window !== "undefined") {
        localStorage.setItem(STORAGE_KEY, newLang);
      }
    }
  }, []);

  useEffect(() => {
    loader(lang).then(setTranslations);
  }, [lang, loader]);

  return (
    <TranslatorContext.Provider value={{ lang, translations, translate, setLang }}>
      {children}
    </TranslatorContext.Provider>
  );
}

const SECONDS = 1000;
const MINUTES = SECONDS * 60;
const HOURS = MINUTES * 60;
const DAYS = HOURS * 24;
const WEEKS = DAYS * 7;
const MONTHS = DAYS * 30;
const YEARS = DAYS * 365;

const relativeFormat = [
  { limit: YEARS, unit: "years" },
  { limit: MONTHS, unit: "months" },
  { limit: WEEKS, unit: "weeks" },
  { limit: DAYS, unit: "days" },
  { limit: HOURS, unit: "hours" },
  { limit: MINUTES, unit: "minutes" },
  { limit: SECONDS, unit: "seconds" },
] as const;

export function useTranslate() {
  const { translate, lang, setLang } = useContext(TranslatorContext);
  const dateFormat = (
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
  };

  const relativeDateFormat = (
    date: Date | string | null | undefined,
    options: Intl.RelativeTimeFormatOptions = {
      style: "long",
    }
  ) => {
    if (!date) {
      return "";
    }
    const distanceInSeconds =
      (date instanceof Date ? date.getTime() : parseDate(date).getTime()) -
      Date.now();

    const formatter = new Intl.RelativeTimeFormat(lang, options);
    for (const { limit, unit } of relativeFormat) {
      if (Math.abs(distanceInSeconds) > limit) {
        return formatter.format(Math.round(distanceInSeconds / limit), unit);
      }
    }
    return "";
  };

  return {
    lang,
    setLang,
    __: translate,
    dateFormat: dateFormat,
    relativeDateFormat,
    dateTimeFormat: (
      date: Date | string | null | undefined,
      options: Intl.DateTimeFormatOptions = {
        hour: "2-digit",
        hour12: false,
        minute: "2-digit",
        day: "numeric",
        month: "short",
        year: "numeric",
      }
    ) => {
      return dateFormat(date, options);
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
