import { Fragment } from "react";
import type { AppRoute } from "/routes.tsx";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";

const documentTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      queryLoader: ({ organizationId, documentId, versionId }) => {
        const basePath = `/organizations/${organizationId}/documents/${documentId}`;
        const redirectPath = versionId
          ? `${basePath}/versions/${versionId}/description`
          : `${basePath}/description`;
        throw redirect(redirectPath);
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
    Component: lazy(
      () => import("/pages/organizations/documents/DocumentsPage"),
    ),
  },
  {
    path: "documents/:documentId",
    fallback: PageSkeleton,
    Component: lazy(
      () => import("../pages/organizations/documents/DocumentDetailPage"),
    ),
    children: [...documentTabs(""), ...documentTabs("versions/:versionId/")],
  },
] satisfies AppRoute[];
