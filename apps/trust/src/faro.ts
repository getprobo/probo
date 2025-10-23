import {
  initializeFaro,
  ReactIntegration,
  getWebInstrumentations,
  createReactRouterV6DataOptions,
} from "@grafana/faro-react";
import { matchRoutes } from "react-router";

export function initFaro() {
  const url = import.meta.env.VITE_FARO_URL;

  if (!url) {
    console.warn("Faro URL not configured. Skipping Faro initialization.");
    return;
  }

  const appName = import.meta.env.VITE_FARO_APP_NAME || "@probo/trust";

  initializeFaro({
    url,
    app: {
      name: appName,
      version: import.meta.env.VITE_APP_VERSION || "0.0.0",
      environment: import.meta.env.MODE,
    },
    instrumentations: [
      ...getWebInstrumentations({
        captureConsole: true,
        captureConsoleDisabledLevels: [],
      }),
      new ReactIntegration({
        router: createReactRouterV6DataOptions({
          matchRoutes,
        }),
      }),
    ],
  });
}
