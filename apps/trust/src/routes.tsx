import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  routeFromAppRoute,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { createBrowserRouter, redirect, useRouteError } from "react-router";

import { MainLayout } from "#/layouts/MainLayout";
import { DocumentsPage } from "#/pages/DocumentsPage";
import { OverviewPage } from "#/pages/OverviewPage";
import { SubprocessorsPage } from "#/pages/SubprocessorsPage";
import {
  currentTrustDocumentsQuery,
  currentTrustGraphQuery,
  currentTrustVendorsQuery,
} from "#/queries/TrustGraph";

import { PageError } from "./components/PageError";
import { MainSkeleton } from "./components/Skeletons/MainSkeleton";
import { TabSkeleton } from "./components/Skeletons/TabSkeleton";
import { consoleEnvironment } from "./providers/RelayProviders";

/**
 * Top level error boundary
 */
function ErrorBoundary() {
  const error = useRouteError();

  return <PageError error={error instanceof Error ? error : new Error("unkown error")} />;
}

const routes = [
  {
    Component: lazy(() => import("#/pages/auth/AuthLayoutLoader")),
    children: [
      {
        path: "/connect",
        Component: lazy(() => import("#/pages/auth/ConnectPageLoader")),
      },
      {
        path: "/verify-magic-link",
        Component: lazy(() => import("#/pages/auth/VerifyMagicLinkPage")),
      },
    ],
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
