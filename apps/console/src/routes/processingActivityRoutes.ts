import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { processingActivitiesQuery, processingActivityNodeQuery } from "/hooks/graph/ProcessingActivityGraph";
import type { AppRoute } from "/routes";

export const processingActivityRoutes = [
  {
    path: "processing-activities",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, processingActivitiesQuery, {
        organizationId: params.organizationId,
        snapshotId: null
      }),
    Component: lazy(
      () => import("/pages/organizations/processingActivities/ProcessingActivitiesPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/processing-activities",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, processingActivitiesQuery, {
        organizationId: params.organizationId,
        snapshotId: params.snapshotId
      }),
    Component: lazy(
      () => import("/pages/organizations/processingActivities/ProcessingActivitiesPage")
    ),
  },
  {
    path: "processing-activities/:activityId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, processingActivityNodeQuery, {
        processingActivityId: params.activityId
      }),
    Component: lazy(
      () => import("/pages/organizations/processingActivities/ProcessingActivityDetailsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/processing-activities/:activityId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, processingActivityNodeQuery, {
        processingActivityId: params.activityId
      }),
    Component: lazy(
      () => import("/pages/organizations/processingActivities/ProcessingActivityDetailsPage")
    ),
  },
] satisfies AppRoute[];
