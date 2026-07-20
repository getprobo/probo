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

import { getCompliancePortalUrl } from "@probo/helpers";
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
          <TabLink to={getCompliancePortalUrl("overview")}>{__("Overview")}</TabLink>
          <TabLink to={getCompliancePortalUrl("documents")}>{__("Documents")}</TabLink>
          <TabLink to={getCompliancePortalUrl("subprocessors")}>
            {__("Subprocessors")}
          </TabLink>
          <TabLink to={getCompliancePortalUrl("updates")}>{__("Updates")}</TabLink>
        </Tabs>
        <TabSkeleton />
      </main>
    </div>
  );
}
