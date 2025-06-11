import { lazy } from "@probo/react-lazy";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { tasksQuery } from "/hooks/graph/TaskGraph";
import type { AppRoute } from "/routes.tsx";

export const taskRoutes = [
  {
    path: "tasks",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, tasksQuery, {
        organizationId,
      }),
    Component: lazy(() => import("/pages/organizations/tasks/TasksPage")),
  },
] satisfies AppRoute[];
