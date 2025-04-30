import posthog from "posthog-js";
import { PostHogProvider } from "posthog-js/react";
import { lazy } from "@probo/react-lazy";
import { StrictMode, Suspense } from "react";
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

posthog.init(process.env.POSTHOG_KEY!, {
  api_host: process.env.POSTHOG_HOST,
  session_recording: {
    maskAllInputs: true,
  },
  loaded: (posthog) => {
    if (!process.env.POSTHOG_KEY) posthog.debug();
  },
});

const OrganizationSelectionPage = lazy(
  () => import("./pages/OrganizationSelectionPage")
);

function App() {
  return (
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
}

function ErrorBoundaryWithLocation({
  children,
}: {
  children: React.ReactNode;
}) {
  const location = useLocation();
  return <ErrorBoundary key={location.pathname}>{children}</ErrorBoundary>;
}

const container = document.getElementById("root");
const root = createRoot(container!);
root.render(<App />);
