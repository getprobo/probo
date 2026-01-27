import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
} from "@probo/routes";
import { Fragment } from "react";
import { type LoaderFunctionArgs, redirect } from "react-router";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

const documentTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      loader: ({
        params: { organizationId, documentId, versionId },
      }: LoaderFunctionArgs) => {
        const basePath = `/organizations/${organizationId}/documents/${documentId}`;
        const redirectPath = versionId
          ? `${basePath}/versions/${versionId}/description`
          : `${basePath}/description`;
        // eslint-disable-next-line
        throw redirect(redirectPath);
      },
      Component: Fragment,
    },
    {
      path: `${prefix}description`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("#/pages/organizations/documents/description/DocumentDescriptionPageLoader"),
      ),
    },
    {
      path: `${prefix}controls`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("#/pages/organizations/documents/controls/DocumentControlsPageLoader"),
      ),
    },
    {
      path: `${prefix}signatures`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("#/pages/organizations/documents/signatures/DocumentSignaturesPageLoader"),
      ),
    },
  ];
};

export const documentsRoutes = [
  {
    path: "documents",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/documents/DocumentsPageLoader")),
  },
  {
    path: "documents/:documentId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/documents/DocumentLayoutLoader")),
    children: [...documentTabs(""), ...documentTabs("versions/:versionId/")],
  },
] satisfies AppRoute[];
