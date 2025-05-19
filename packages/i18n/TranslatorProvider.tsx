import {
    createContext,
    PropsWithChildren,
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
        {} as Record<string, string>,
    );
    const translate = useCallback<Context["translate"]>(
        (s) => {
            return translations[s] ? translations[s] : s;
        },
        [translations],
    );

    useEffect(() => {
        loader(lang).then(setTranslations);
    }, [lang]);

    return (
        <TranslatorContext.Provider
            value={{ lang, translations, translate }}
        >
            {children}
        </TranslatorContext.Provider>
    );
}

export function useTranslate() {
    const { translate, lang } =
        useContext(TranslatorContext);
    return {
        lang,
        __: translate,
    };
}
