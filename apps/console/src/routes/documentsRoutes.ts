import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { documentsQuery } from "/hooks/graph/DocumentGraph";
import { documentNodeQuery } from "/hooks/graph/DocumentGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect, type LoaderFunctionArgs } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";
import type { DocumentGraphListQuery } from "/hooks/graph/__generated__/DocumentGraphListQuery.graphql";
import type { DocumentGraphNodeQuery } from "/hooks/graph/__generated__/DocumentGraphNodeQuery.graphql";

const documentTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      loader: ({ params: { organizationId, documentId, versionId } }: LoaderFunctionArgs) => {
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
  ];
};

export const documentsRoutes = [
  {
    path: "documents",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<DocumentGraphListQuery>(relayEnvironment, documentsQuery, { organizationId: organizationId! }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/documents/DocumentsPage"),
    )),
  },
  {
    path: "documents/:documentId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ documentId }) =>
      loadQuery<DocumentGraphNodeQuery>(relayEnvironment, documentNodeQuery, { documentId: documentId! }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/documents/DocumentDetailPage"),
    )),
    children: [...documentTabs(""), ...documentTabs("versions/:versionId/")],
  },
] satisfies AppRoute[];
