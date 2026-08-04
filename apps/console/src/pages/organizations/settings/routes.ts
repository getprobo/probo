// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { redirect } from "react-router";

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import { MalaysiaPDPASettingsPageSkeleton } from "./malaysia-pdpa/MalaysiaPDPASettingsPageSkeleton";

export const settingsRoutes = [
  {
    path: "settings",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/iam/organizations/settings/SettingsLayout"),
    ),
    children: [
      {
        index: true,
        loader: () => {
          // eslint-disable-next-line @typescript-eslint/only-throw-error
          throw redirect("general");
        },
      },
      {
        path: "general",
        Component: lazy(
          () =>
            import("#/pages/iam/organizations/settings/GeneralSettingsPageLoader"),
        ),
      },
      {
        path: "saml-sso",
        Component: lazy(
          () =>
            import("#/pages/iam/organizations/settings/SAMLSettingsPageLoader"),
        ),
      },
      {
        path: "scim",
        Component: lazy(
          () =>
            import("#/pages/iam/organizations/settings/SCIMSettingsPageLoader"),
        ),
      },
      {
        path: "webhooks",
        Component: lazy(
          () =>
            import("#/pages/iam/organizations/settings/WebhooksSettingsPageLoader"),
        ),
      },
      {
        path: "audit-log",
        Component: lazy(
          () =>
            import("#/pages/iam/organizations/settings/AuditLogSettingsPageLoader"),
        ),
      },
      {
        path: "malaysia-pdpa",
        Fallback: MalaysiaPDPASettingsPageSkeleton,
        Component: lazy(
          () =>
            import("./malaysia-pdpa/MalaysiaPDPASettingsPageLoader"),
        ),
      },
    ],
  },
] satisfies AppRoute[];
