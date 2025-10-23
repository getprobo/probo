import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { FaroErrorBoundary } from "@grafana/faro-react";
import "./index.css";
import { App } from "./App";
import { RelayProvider } from "./providers/RelayProviders";
import { TranslatorProvider } from "./providers/TranslatorProvider";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { initFaro } from "./faro";

initFaro();

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <FaroErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <RelayProvider>
          <TranslatorProvider>
            <App />
          </TranslatorProvider>
        </RelayProvider>
      </QueryClientProvider>
    </FaroErrorBoundary>
  </StrictMode>
);
