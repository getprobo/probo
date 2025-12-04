import { lazy } from "@probo/react-lazy";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { tasksQuery } from "/hooks/graph/TaskGraph";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "@probo/routes";
import type { TaskGraphQuery } from "/hooks/graph/__generated__/TaskGraphQuery.graphql";
export const taskRoutes = [
  {
    path: "tasks",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<TaskGraphQuery>(relayEnvironment, tasksQuery, {
        organizationId,
      }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/tasks/TasksPage"))),
  },
] satisfies AppRoute[];
