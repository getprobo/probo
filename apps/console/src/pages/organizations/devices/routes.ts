// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";
import { Fragment } from "react";
import { type LoaderFunctionArgs, redirect } from "react-router";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

const deviceTabs = () => [
  {
    path: "",
    loader: ({
      params: { organizationId, deviceId },
    }: LoaderFunctionArgs) => {
      // eslint-disable-next-line
      throw redirect(
        `/organizations/${organizationId}/devices/${deviceId}/postures`,
      );
    },
    Component: Fragment,
  },
  {
    path: "postures",
    Fallback: LinkCardSkeleton,
    Component: lazy(
      () =>
        import("#/pages/organizations/devices/postures/DevicePosturesPageLoader"),
    ),
  },
];

export const deviceRoutes = [
  {
    path: "devices",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/devices/DevicesPageLoader"),
    ),
  },
  {
    path: "devices/:deviceId",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/devices/DeviceLayoutLoader"),
    ),
    children: deviceTabs(),
  },
] satisfies AppRoute[];
