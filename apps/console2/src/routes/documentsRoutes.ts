import { Fragment, lazy } from "react";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { documentsQuery } from "/hooks/graph/DocumentGraph";
import { documentNodeQuery } from "/hooks/graph/DocumentGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";

export const documentsRoutes = [
  {
    path: "documents",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, documentsQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/documents/DocumentsPage")
    ),
  },
  {
    path: "documents/:documentId",
    fallback: PageSkeleton,
    queryLoader: ({ documentId }) =>
      loadQuery(relayEnvironment, documentNodeQuery, { documentId }),
    Component: lazy(
      () => import("../pages/organizations/documents/DocumentDetailPage")
    ),
    children: [
      {
        path: "",
        queryLoader: ({ organizationId, documentId }) => {
          throw redirect(
            `/organizations/${organizationId}/documents/${documentId}/description`
          );
        },
        Component: Fragment,
      },
      {
        path: "description",
        Component: lazy(
          () =>
            import(
              "../pages/organizations/documents/tabs/DocumentDescriptionTab"
            )
        ),
      },
      {
        path: "controls",
        Component: lazy(
          () =>
            import("../pages/organizations/documents/tabs/DocumentControlsTab")
        ),
      },
    ],
  },
] satisfies AppRoute[];
