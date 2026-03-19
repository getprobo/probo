import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { type LoaderFunctionArgs, redirect } from "react-router";

import type { ContextPageLoaderQuery } from "#/__generated__/core/ContextPageLoaderQuery.graphql";
import type { MeetingDetailPageQuery } from "#/__generated__/core/MeetingDetailPageQuery.graphql";
import type { MeetingsPageQuery } from "#/__generated__/core/MeetingsPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { contextPageQuery } from "#/pages/organizations/context/ContextPageLoader";
import { meetingDetailPageQuery } from "#/pages/organizations/meetings/MeetingDetailPage";
import { meetingsPageQuery } from "#/pages/organizations/meetings/MeetingsPage";

const meetingTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      loader: ({
        params: { organizationId, meetingId },
      }: LoaderFunctionArgs) => {
        const basePath = `/organizations/${organizationId}/context/meetings/${meetingId}`;
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

export const contextRoutes = [
  {
    path: "context",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/context/ContextLayout"),
    ),
    children: [
      {
        path: "",
        loader: ({
          params: { organizationId },
        }: LoaderFunctionArgs) => {
          // eslint-disable-next-line
          throw redirect(`/organizations/${organizationId}/context/overview`);
        },
        Component: Fragment,
      },
      {
        path: "overview",
        Fallback: PageSkeleton,
        loader: loaderFromQueryLoader(({ organizationId }) =>
          loadQuery<ContextPageLoaderQuery>(
            coreEnvironment,
            contextPageQuery,
            { organizationId },
          ),
        ),
        Component: withQueryRef(
          lazy(
            () => import("#/pages/organizations/context/ContextPageLoader"),
          ),
        ),
      },
      {
        path: "meetings",
        Fallback: PageSkeleton,
        loader: loaderFromQueryLoader(({ organizationId }) =>
          loadQuery<MeetingsPageQuery>(
            coreEnvironment,
            meetingsPageQuery,
            { organizationId },
          ),
        ),
        Component: withQueryRef(
          lazy(
            () => import("#/pages/organizations/meetings/MeetingsPage"),
          ),
        ),
      },
      {
        path: "meetings/:meetingId",
        Fallback: PageSkeleton,
        loader: loaderFromQueryLoader(({ meetingId }) =>
          loadQuery<MeetingDetailPageQuery>(
            coreEnvironment,
            meetingDetailPageQuery,
            { meetingId },
          ),
        ),
        Component: withQueryRef(
          lazy(
            () => import("../pages/organizations/meetings/MeetingDetailPage"),
          ),
        ),
        children: [...meetingTabs("")],
      },
    ],
  },
] satisfies AppRoute[];
