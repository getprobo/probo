import { ThemeProvider } from "@probo/ui";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createRoot } from "react-dom/client";

import "./index.css";

import { App } from "./App";
import { TranslatorProvider } from "./providers/TranslatorProvider";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

createRoot(document.getElementById("root")!).render(
  <QueryClientProvider client={queryClient}>
    <ThemeProvider>
      <TranslatorProvider>
        <App />
      </TranslatorProvider>
    </ThemeProvider>
  </QueryClientProvider>,
);
