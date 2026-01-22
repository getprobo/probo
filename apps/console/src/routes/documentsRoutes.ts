import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { redirect, type LoaderFunctionArgs } from "react-router";
import { lazy } from "@probo/react-lazy";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

import { coreEnvironment } from "/environments";
import { documentsQuery, documentNodeQuery } from "/hooks/graph/DocumentGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import type { DocumentGraphListQuery } from "/__generated__/core/DocumentGraphListQuery.graphql";
import type { DocumentGraphNodeQuery } from "/__generated__/core/DocumentGraphNodeQuery.graphql";

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
          import("../pages/organizations/documents/tabs/DocumentDescriptionTab"),
      ),
    },
    {
      path: `${prefix}controls`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("../pages/organizations/documents/tabs/DocumentControlsTab"),
      ),
    },
    {
      path: `${prefix}signatures`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("../pages/organizations/documents/tabs/DocumentSignaturesTab"),
      ),
    },
  ];
};

export const documentsRoutes = [
  {
    path: "documents",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<DocumentGraphListQuery>(coreEnvironment, documentsQuery, {
        organizationId: organizationId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/documents/DocumentsPage")),
    ),
  },
  {
    path: "documents/:documentId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ documentId }) =>
      loadQuery<DocumentGraphNodeQuery>(coreEnvironment, documentNodeQuery, {
        documentId: documentId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/documents/DocumentDetailPage")),
    ),
    children: [...documentTabs(""), ...documentTabs("versions/:versionId/")],
  },
] satisfies AppRoute[];
