import posthog from "posthog-js";
import { PostHogProvider } from "posthog-js/react";
import { lazy } from "@probo/react-lazy";
import { StrictMode, Suspense, type PropsWithChildren } from "react";
import { createRoot } from "react-dom/client";
import { HelmetProvider } from "react-helmet-async";
import { RelayEnvironmentProvider } from "react-relay";
import { BrowserRouter, Route, Routes, useLocation } from "react-router";
import "./App.css";
import "./styles/policy-content.css";
import ErrorBoundary from "./components/ErrorBoundary";
import AuthLayout from "./layouts/AuthLayout";
import { RelayEnvironment } from "./RelayEnvironment";
import { AuthenticationRoutes } from "./pages/authentication/Routes";
import { OrganizationsRoutes } from "./pages/organizations/Routes";
import SigningRequestsPage from "./pages/SigningRequestsPage";

posthog.init(import.meta.env.POSTHOG_KEY!, {
  api_host: import.meta.env.POSTHOG_HOST,
  session_recording: {
    maskAllInputs: true,
  },
  loaded: (posthog) => {
    if (!import.meta.env.POSTHOG_KEY) posthog.debug();
  },
});

const OrganizationSelectionPage = lazy(
  () => import("./pages/OrganizationSelectionPage")
);

function ErrorBoundaryWithLocation({ children }: PropsWithChildren) {
  const location = useLocation();
  return <ErrorBoundary key={location.pathname}>{children}</ErrorBoundary>;
}

console.log(import.meta.env.API_SERVER_HOST);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <PostHogProvider client={posthog}>
      <RelayEnvironmentProvider environment={RelayEnvironment}>
        <HelmetProvider>
          <BrowserRouter>
            <Routes>
              <Route path="/*">
                <Route
                  index
                  element={
                    <Suspense>
                      <ErrorBoundaryWithLocation>
                        <OrganizationSelectionPage />
                      </ErrorBoundaryWithLocation>
                    </Suspense>
                  }
                />

                <Route
                  path="organizations/*"
                  element={<OrganizationsRoutes />}
                />

                <Route
                  path="policies/signing-requests"
                  element={<SigningRequestsPage />}
                />

                <Route
                  path="*"
                  element={
                    <AuthLayout>
                      <AuthenticationRoutes />
                    </AuthLayout>
                  }
                />
              </Route>
            </Routes>
          </BrowserRouter>
        </HelmetProvider>
      </RelayEnvironmentProvider>
    </PostHogProvider>
  </StrictMode>
);
