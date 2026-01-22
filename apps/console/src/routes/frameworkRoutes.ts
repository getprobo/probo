import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";

import type { FrameworkGraphControlNodeQuery } from "#/__generated__/core/FrameworkGraphControlNodeQuery.graphql";
import type { FrameworkGraphListQuery } from "#/__generated__/core/FrameworkGraphListQuery.graphql";
import type { FrameworkGraphNodeQuery } from "#/__generated__/core/FrameworkGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import {
  frameworkControlNodeQuery,
  frameworkNodeQuery,
  frameworksQuery,
} from "#/hooks/graph/FrameworkGraph";

import { ControlSkeleton } from "../components/skeletons/ControlSkeleton";

export const frameworkRoutes = [
  {
    path: "frameworks",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<FrameworkGraphListQuery>(coreEnvironment, frameworksQuery, {
        organizationId: organizationId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/frameworks/FrameworksPage")),
    ),
  },
  {
    path: "frameworks/:frameworkId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ frameworkId }) =>
      loadQuery<FrameworkGraphNodeQuery>(coreEnvironment, frameworkNodeQuery, {
        frameworkId: frameworkId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/frameworks/FrameworkDetailPage")),
    ),
    children: [
      {
        path: "",
        Component: Fragment,
      },
      {
        path: "controls/:controlId",
        Fallback: ControlSkeleton,
        loader: loaderFromQueryLoader(({ controlId }) =>
          loadQuery<FrameworkGraphControlNodeQuery>(
            coreEnvironment,
            frameworkControlNodeQuery,
            { controlId: controlId },
          ),
        ),
        Component: withQueryRef(
          lazy(
            () =>
              import("#/pages/organizations/frameworks/FrameworkControlPage"),
          ),
        ),
      },
    ],
  },
] satisfies AppRoute[];
