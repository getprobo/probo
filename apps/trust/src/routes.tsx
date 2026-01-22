import { createBrowserRouter, redirect, useRouteError } from "react-router";
import { Fragment, lazy } from "react";
import { consoleEnvironment } from "./providers/RelayProviders.tsx";
import { loadQuery } from "react-relay";
import { PageError } from "./components/PageError.tsx";
import { MainLayout } from "/layouts/MainLayout";
import {
  currentTrustGraphQuery,
  currentTrustDocumentsQuery,
  currentTrustVendorsQuery,
} from "/queries/TrustGraph";
import { OverviewPage } from "/pages/OverviewPage";
import { DocumentsPage } from "/pages/DocumentsPage";
import { SubprocessorsPage } from "/pages/SubprocessorsPage";
import { TabSkeleton } from "./components/Skeletons/TabSkeleton.tsx";
import { MainSkeleton } from "./components/Skeletons/MainSkeleton.tsx";
import {
  loaderFromQueryLoader,
  routeFromAppRoute,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

/**
 * Top level error boundary
 */
function ErrorBoundary({ error: propsError }: { error?: string }) {
  const error = useRouteError() ?? propsError;

  return <PageError error={error instanceof Error ? error.message : ""} />;
}

const routes = [
  {
    path: "/connect",
    Component: lazy(() => import("./pages/auth/ConnectPageLoader.tsx")),
  },
  {
    path: "/verify-magic-link",
    Component: lazy(() => import("./pages/auth/VerifyMagicLinkPage.tsx")),
  },
  {
    path: "/",
    loader: () => {
      // eslint-disable-next-line
      throw redirect("/overview");
    },
    Component: Fragment,
    ErrorBoundary: ErrorBoundary,
  },
  // Custom domain routes (subdomain-based)
  {
    path: "/overview",
    loader: loaderFromQueryLoader(() =>
      loadQuery(consoleEnvironment, currentTrustGraphQuery, {}),
    ),
    Component: withQueryRef(MainLayout),
    Fallback: MainSkeleton,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        Fallback: TabSkeleton,
        Component: OverviewPage,
      },
    ],
  },
  {
    path: "/documents",
    loader: loaderFromQueryLoader(() =>
      loadQuery(consoleEnvironment, currentTrustGraphQuery, {}),
    ),
    Component: withQueryRef(MainLayout),
    Fallback: MainSkeleton,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        loader: loaderFromQueryLoader(() =>
          loadQuery(consoleEnvironment, currentTrustDocumentsQuery, {}),
        ),
        Fallback: TabSkeleton,
        Component: withQueryRef(DocumentsPage),
      },
    ],
  },
  {
    path: "/subprocessors",
    loader: loaderFromQueryLoader(() =>
      loadQuery(consoleEnvironment, currentTrustGraphQuery, {}),
    ),
    Component: withQueryRef(MainLayout),
    Fallback: MainSkeleton,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        loader: loaderFromQueryLoader(() =>
          loadQuery(consoleEnvironment, currentTrustVendorsQuery, {}),
        ),
        Fallback: TabSkeleton,
        Component: withQueryRef(SubprocessorsPage),
      },
    ],
  },
  // Fallback URL to the NotFound Page
  {
    path: "*",
    Component: PageError,
  },
] satisfies AppRoute[];

// Detect basename from current URL path
// If URL starts with /trust/{slug}, extract that as the basename
// Otherwise, use "/" for custom domains
function getBasename(): string {
  const path = window.location.pathname;
  const trustMatch = path.match(/^\/trust\/[^/]+/);
  return trustMatch ? trustMatch[0] : "/";
}

export const router = createBrowserRouter(routes.map(routeFromAppRoute), {
  basename: getBasename(),
});
