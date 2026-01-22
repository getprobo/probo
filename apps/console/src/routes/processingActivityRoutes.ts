import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { ProcessingActivityGraphListQuery } from "#/__generated__/core/ProcessingActivityGraphListQuery.graphql";
import type { ProcessingActivityGraphNodeQuery } from "#/__generated__/core/ProcessingActivityGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import {
  processingActivitiesQuery,
  processingActivityNodeQuery,
} from "#/hooks/graph/ProcessingActivityGraph";

export const processingActivityRoutes = [
  {
    path: "processing-activities",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<ProcessingActivityGraphListQuery>(
        coreEnvironment,
        processingActivitiesQuery,
        {
          organizationId: organizationId,
          snapshotId: null,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/processingActivities/ProcessingActivitiesPage"),
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
          organizationId: organizationId,
          snapshotId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/processingActivities/ProcessingActivitiesPage"),
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
          processingActivityId: activityId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/processingActivities/ProcessingActivityDetailsPage"),
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
          processingActivityId: activityId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/processingActivities/ProcessingActivityDetailsPage"),
      ),
    ),
  },
] satisfies AppRoute[];
