import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { type LoaderFunctionArgs, redirect } from "react-router";

import type { DocumentGraphNodeQuery } from "#/__generated__/core/DocumentGraphNodeQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { documentNodeQuery } from "#/hooks/graph/DocumentGraph";

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
    Component: lazy(() => import("#/pages/organizations/documents/DocumentsPageLoader")),
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
