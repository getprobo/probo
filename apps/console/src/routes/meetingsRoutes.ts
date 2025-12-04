import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { meetingsQuery } from "/hooks/graph/MeetingGraph";
import { meetingNodeQuery } from "/hooks/graph/MeetingGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect, type LoaderFunctionArgs } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import type { MeetingGraphListQuery } from "/hooks/graph/__generated__/MeetingGraphListQuery.graphql";
import type { MeetingGraphNodeQuery } from "/hooks/graph/__generated__/MeetingGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef } from "/routes";

const meetingTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      loader: ({ params: { organizationId, meetingId } }: LoaderFunctionArgs) => {
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
  ];
};

export const meetingsRoutes = [
  {
    path: "meetings",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<MeetingGraphListQuery>(relayEnvironment, meetingsQuery, { organizationId }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/meetings/MeetingsPage"),
    )),
  },
  {
    path: "meetings/:meetingId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ meetingId }) =>
      loadQuery<MeetingGraphNodeQuery>(relayEnvironment, meetingNodeQuery, { meetingId }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/meetings/MeetingDetailPage"),
    )),
    children: [...meetingTabs("")],
  },
];

