import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import { App } from "./App.tsx";
import { RelayProvider } from "./providers/RelayProviders.tsx";
import { TranslatorProvider } from "./providers/TranslatorProvider.tsx";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <RelayProvider>
      <TranslatorProvider>
        <App />
      </TranslatorProvider>
    </RelayProvider>
  </StrictMode>
);
