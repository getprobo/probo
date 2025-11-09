import { Fragment } from "react";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { meetingsQuery } from "/hooks/graph/MeetingGraph";
import { meetingNodeQuery } from "/hooks/graph/MeetingGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";

const meetingTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      queryLoader: ({ organizationId, meetingId }) => {
        const basePath = `/organizations/${organizationId}/meetings/${meetingId}`;
        const redirectPath = `${basePath}/minutes`;
        throw redirect(redirectPath);
      },
      Component: Fragment,
    },
    {
      path: `${prefix}minutes`,
      fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import(
            "../pages/organizations/meetings/tabs/MeetingMinutesTab"
          ),
      ),
    },
  ] satisfies AppRoute[];
};

export const meetingsRoutes = [
  {
    path: "meetings",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, meetingsQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/meetings/MeetingsPage"),
    ),
  },
  {
    path: "meetings/:meetingId",
    fallback: PageSkeleton,
    queryLoader: ({ meetingId }) =>
      loadQuery(relayEnvironment, meetingNodeQuery, { meetingId }),
    Component: lazy(
      () => import("../pages/organizations/meetings/MeetingDetailPage"),
    ),
    children: [...meetingTabs("")],
  },
] satisfies AppRoute[];

