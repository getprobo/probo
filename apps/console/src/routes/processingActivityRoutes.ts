import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { processingActivitiesQuery, processingActivityNodeQuery } from "/hooks/graph/ProcessingActivityGraph";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";
import type { ProcessingActivityGraphListQuery } from "/hooks/graph/__generated__/ProcessingActivityGraphListQuery.graphql";
import type { ProcessingActivityGraphNodeQuery } from "/hooks/graph/__generated__/ProcessingActivityGraphNodeQuery.graphql";

export const processingActivityRoutes = [
  {
    path: "processing-activities",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<ProcessingActivityGraphListQuery>(relayEnvironment, processingActivitiesQuery, {
        organizationId: organizationId!,
        snapshotId: null
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/processingActivities/ProcessingActivitiesPage")
    )),
  },
  {
    path: "snapshots/:snapshotId/processing-activities",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<ProcessingActivityGraphListQuery>(relayEnvironment, processingActivitiesQuery, {
        organizationId: organizationId!,
        snapshotId,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/processingActivities/ProcessingActivitiesPage")
    )),
  },
  {
    path: "processing-activities/:activityId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ activityId }) =>
      loadQuery<ProcessingActivityGraphNodeQuery>(relayEnvironment, processingActivityNodeQuery, {
        processingActivityId: activityId!,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/processingActivities/ProcessingActivityDetailsPage")
    )),
  },
  {
    path: "snapshots/:snapshotId/processing-activities/:activityId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ activityId }) =>
      loadQuery<ProcessingActivityGraphNodeQuery>(relayEnvironment, processingActivityNodeQuery, {
        processingActivityId: activityId!,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/processingActivities/ProcessingActivityDetailsPage")
    )),
  },
] satisfies AppRoute[];
