// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const contextRoutes = [
  {
    path: "context",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/context/ContextLayoutLoader"),
    ),
    children: [
      {
        index: true,
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/context/ContextPageLoader"),
        ),
      },
      {
        path: "meetings",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/meetings/MeetingsPageLoader"),
        ),
      },
      {
        path: "meetings/:meetingId",
        Fallback: PageSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/meetings/MeetingDetailPageLoader"),
        ),
        children: [
          {
            index: true,
            Fallback: LinkCardSkeleton,
            Component: lazy(
              () => import("#/pages/organizations/meetings/tabs/MeetingMinutesTab"),
            ),
          },
        ],
      },
    ],
  },
] satisfies AppRoute[];
