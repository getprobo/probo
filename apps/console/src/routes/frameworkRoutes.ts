import { loadQuery } from "react-relay";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { relayEnvironment } from "/providers/RelayProviders";
import {
  frameworksQuery,
  frameworkNodeQuery,
  frameworkControlNodeQuery,
} from "/hooks/graph/FrameworkGraph";
import { Fragment } from "react";
import { lazy } from "@probo/react-lazy";
import { ControlSkeleton } from "../components/skeletons/ControlSkeleton";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";
import type { FrameworkGraphListQuery } from "/hooks/graph/__generated__/FrameworkGraphListQuery.graphql";
import type { FrameworkGraphNodeQuery } from "/hooks/graph/__generated__/FrameworkGraphNodeQuery.graphql";
import type { FrameworkGraphControlNodeQuery } from "/hooks/graph/__generated__/FrameworkGraphControlNodeQuery.graphql";

export const frameworkRoutes = [
  {
    path: "frameworks",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<FrameworkGraphListQuery>(relayEnvironment, frameworksQuery, { organizationId: organizationId! }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/frameworks/FrameworksPage")
    )),
  },
  {
    path: "frameworks/:frameworkId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ frameworkId }) =>
      loadQuery<FrameworkGraphNodeQuery>(relayEnvironment, frameworkNodeQuery, { frameworkId: frameworkId! }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/frameworks/FrameworkDetailPage")
    )),
    children: [
      {
        path: "",
        Component: Fragment,
      },
      {
        path: "controls/:controlId",
        fallback: ControlSkeleton,
        loader: loaderFromQueryLoader(({ controlId }) =>
          loadQuery<FrameworkGraphControlNodeQuery>(relayEnvironment, frameworkControlNodeQuery, { controlId: controlId! }),
        ),
        Component: withQueryRef(lazy(
          () => import("/pages/organizations/frameworks/FrameworkControlPage")
        )),
      },
    ],
  },
] satisfies AppRoute[];
