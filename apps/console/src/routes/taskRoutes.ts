import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { TaskGraphQuery } from "#/__generated__/core/TaskGraphQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { tasksQuery } from "#/hooks/graph/TaskGraph";
export const taskRoutes = [
  {
    path: "tasks",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<TaskGraphQuery>(coreEnvironment, tasksQuery, {
        organizationId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/tasks/TasksPage")),
    ),
  },
] satisfies AppRoute[];
