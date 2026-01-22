import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { type LoaderFunctionArgs, redirect } from "react-router";

import type { MeetingGraphListQuery } from "#/__generated__/core/MeetingGraphListQuery.graphql";
import type { MeetingGraphNodeQuery } from "#/__generated__/core/MeetingGraphNodeQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { meetingNodeQuery, meetingsQuery } from "#/hooks/graph/MeetingGraph";

const meetingTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      loader: ({
        params: { organizationId, meetingId },
      }: LoaderFunctionArgs) => {
        const basePath = `/organizations/${organizationId}/meetings/${meetingId}`;
        const redirectPath = `${basePath}/minutes`;
        // eslint-disable-next-line
        throw redirect(redirectPath);
      },
      Component: Fragment,
    },
    {
      path: `${prefix}minutes`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () => import("../pages/organizations/meetings/tabs/MeetingMinutesTab"),
      ),
    },
  ];
};

export const meetingsRoutes = [
  {
    path: "meetings",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<MeetingGraphListQuery>(coreEnvironment, meetingsQuery, {
        organizationId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/meetings/MeetingsPage")),
    ),
  },
  {
    path: "meetings/:meetingId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ meetingId }) =>
      loadQuery<MeetingGraphNodeQuery>(coreEnvironment, meetingNodeQuery, {
        meetingId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/meetings/MeetingDetailPage")),
    ),
    children: [...meetingTabs("")],
  },
] satisfies AppRoute[];
