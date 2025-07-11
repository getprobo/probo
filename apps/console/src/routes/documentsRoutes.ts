import { Fragment } from "react";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { documentsQuery } from "/hooks/graph/DocumentGraph";
import { documentNodeQuery } from "/hooks/graph/DocumentGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";

const documentTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      queryLoader: ({ organizationId, documentId }) => {
        throw redirect(
          `/organizations/${organizationId}/documents/${documentId}/description`,
        );
      },
      Component: Fragment,
    },
    {
      path: `${prefix}description`,
      fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import(
            "../pages/organizations/documents/tabs/DocumentDescriptionTab"
          ),
      ),
    },
    {
      path: `${prefix}controls`,
      fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("../pages/organizations/documents/tabs/DocumentControlsTab"),
      ),
    },
    {
      path: `${prefix}signatures`,
      fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("../pages/organizations/documents/tabs/DocumentSignaturesTab"),
      ),
    },
  ] satisfies AppRoute[];
};

export const documentsRoutes = [
  {
    path: "documents",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, documentsQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/documents/DocumentsPage"),
    ),
  },
  {
    path: "documents/:documentId",
    fallback: PageSkeleton,
    queryLoader: ({ documentId }) =>
      loadQuery(relayEnvironment, documentNodeQuery, { documentId }),
    Component: lazy(
      () => import("../pages/organizations/documents/DocumentDetailPage"),
    ),
    children: [...documentTabs(""), ...documentTabs("versions/:versionId/")],
  },
] satisfies AppRoute[];
