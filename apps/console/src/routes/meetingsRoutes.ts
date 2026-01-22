import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { redirect, type LoaderFunctionArgs } from "react-router";
import { lazy } from "@probo/react-lazy";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

import { coreEnvironment } from "/environments";
import { meetingsQuery, meetingNodeQuery } from "/hooks/graph/MeetingGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import type { MeetingGraphListQuery } from "/__generated__/core/MeetingGraphListQuery.graphql";
import type { MeetingGraphNodeQuery } from "/__generated__/core/MeetingGraphNodeQuery.graphql";

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
      lazy(() => import("/pages/organizations/meetings/MeetingsPage")),
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
