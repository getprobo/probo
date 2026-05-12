// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { getTrustCenterUrl } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Skeleton, TabLink, Tabs } from "@probo/ui";

import { TabSkeleton } from "./TabSkeleton";

export function MainSkeleton() {
  const { __ } = useTranslate();
  return (
    <div className="grid grid-cols-1 max-w-[1280px] mx-4 pt-6 gap-4 lg:mx-auto lg:gap-10 lg:pt-20 lg:grid-cols-[400px_1fr] ">
      <Skeleton className="w-full h-300" />
      <main>
        <Tabs className="mb-8">
          <TabLink to={getTrustCenterUrl("overview")}>{__("Overview")}</TabLink>
          <TabLink to={getTrustCenterUrl("documents")}>{__("Documents")}</TabLink>
          <TabLink to={getTrustCenterUrl("subprocessors")}>
            {__("Subprocessors")}
          </TabLink>
          <TabLink to={getTrustCenterUrl("updates")}>{__("Updates")}</TabLink>
        </Tabs>
        <TabSkeleton />
      </main>
    </div>
  );
}
