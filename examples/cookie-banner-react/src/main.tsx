import { createRoot } from "react-dom/client";
import { installTCFStub } from "@probo/cookie-banner-tcf";
import { App } from "./App";

installTCFStub();

createRoot(document.getElementById("root")!).render(
  <App />
);
