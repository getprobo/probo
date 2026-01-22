import type { PropsWithChildren } from "react";

import { TranslatorProvider as ProboTranslatorProvider } from "../../../../packages/i18n/TranslatorProvider";

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
