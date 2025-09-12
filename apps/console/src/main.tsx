import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import { App } from "./App";
import { RelayProvider } from "./providers/RelayProviders";
import {
  LangProvider,
  TranslatorProvider,
} from "./providers/TranslatorProvider";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RelayProvider>
        <LangProvider>
          <TranslatorProvider>
            <App />
          </TranslatorProvider>
        </LangProvider>
      </RelayProvider>
    </QueryClientProvider>
  </StrictMode>
);
