import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import {
  processingActivitiesQuery,
  processingActivityNodeQuery,
} from "/hooks/graph/ProcessingActivityGraph";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";
import type { ProcessingActivityGraphListQuery } from "/hooks/graph/__generated__/ProcessingActivityGraphListQuery.graphql";
import type { ProcessingActivityGraphNodeQuery } from "/hooks/graph/__generated__/ProcessingActivityGraphNodeQuery.graphql";

export const processingActivityRoutes = [
  {
    path: "processing-activities",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<ProcessingActivityGraphListQuery>(
        coreEnvironment,
        processingActivitiesQuery,
        {
          organizationId: organizationId!,
          snapshotId: null,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("/pages/organizations/processingActivities/ProcessingActivitiesPage"),
      ),
    ),
  },
  {
    path: "snapshots/:snapshotId/processing-activities",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<ProcessingActivityGraphListQuery>(
        coreEnvironment,
        processingActivitiesQuery,
        {
          organizationId: organizationId!,
          snapshotId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("/pages/organizations/processingActivities/ProcessingActivitiesPage"),
      ),
    ),
  },
  {
    path: "processing-activities/:activityId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ activityId }) =>
      loadQuery<ProcessingActivityGraphNodeQuery>(
        coreEnvironment,
        processingActivityNodeQuery,
        {
          processingActivityId: activityId!,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("/pages/organizations/processingActivities/ProcessingActivityDetailsPage"),
      ),
    ),
  },
  {
    path: "snapshots/:snapshotId/processing-activities/:activityId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ activityId }) =>
      loadQuery<ProcessingActivityGraphNodeQuery>(
        coreEnvironment,
        processingActivityNodeQuery,
        {
          processingActivityId: activityId!,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("/pages/organizations/processingActivities/ProcessingActivityDetailsPage"),
      ),
    ),
  },
] satisfies AppRoute[];
