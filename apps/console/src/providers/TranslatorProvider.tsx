import { TranslatorProvider as ProboTranslatorProvider } from "@probo/i18n";
import type { PropsWithChildren } from "react";

// TODO : implement a way to retrieve translations strings
const loader = () => {
  return Promise.resolve({} as Record<string, string>);
};

/**
 * Provider for the translator
 */
export function TranslatorProvider({ children }: PropsWithChildren) {
  return (
    <ProboTranslatorProvider lang="en" loader={loader}>
      {children}
    </ProboTranslatorProvider>
  );
}
