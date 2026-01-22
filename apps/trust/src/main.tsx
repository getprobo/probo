import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import "./index.css";

import { App } from "./App";
import { RelayProvider } from "./providers/RelayProviders";
import { TranslatorProvider } from "./providers/TranslatorProvider";

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
        <TranslatorProvider>
          <App />
        </TranslatorProvider>
      </RelayProvider>
    </QueryClientProvider>
  </StrictMode>,
);
