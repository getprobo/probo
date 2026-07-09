// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { UpdateDetailPageSkeleton } from "./UpdateDetailPageSkeleton";
import { UpdatesPageSkeleton } from "./UpdatesPageSkeleton";

export const updateRoutes = [
  {
    path: "updates",
    Fallback: UpdatesPageSkeleton,
    Component: lazy(() => import("./UpdatesPageLoader")),
  },
  {
    path: "updates/:updateId",
    Fallback: UpdateDetailPageSkeleton,
    Component: lazy(() => import("./UpdateDetailPageLoader")),
  },
] satisfies AppRoute[];
