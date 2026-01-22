import { lazy } from "@probo/react-lazy";
import { loadQuery } from "react-relay";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { coreEnvironment } from "/environments";
import { tasksQuery } from "/hooks/graph/TaskGraph";
import type { TaskGraphQuery } from "/__generated__/core/TaskGraphQuery.graphql";
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
      lazy(() => import("/pages/organizations/tasks/TasksPage")),
    ),
  },
] satisfies AppRoute[];
