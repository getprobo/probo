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
  const { translate, lang } = useContext(TranslatorContext);
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
