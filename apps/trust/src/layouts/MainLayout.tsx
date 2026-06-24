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

import { useFavicon, useSystemTheme } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Logo, TabLink, Tabs } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import { OrganizationSidebar } from "#/components/OrganizationSidebar";
import { useRequestAccessCallback } from "#/hooks/useRequestAccessCallback";
import { TrustCenterProvider } from "#/providers/TrustCenterProvider";
import type { TrustGraphCurrentQuery } from "#/queries/__generated__/TrustGraphCurrentQuery.graphql";
import { currentTrustGraphQuery } from "#/queries/TrustGraph";

type Props = {
  queryRef: PreloadedQuery<TrustGraphCurrentQuery>;
};

export function MainLayout(props: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery<TrustGraphCurrentQuery>(currentTrustGraphQuery, props.queryRef);
  const trustCenter = data.currentTrustCenter;
  const isAuthenticated = data.viewer != null;

  const theme = useSystemTheme();

  useFavicon(
    theme === "dark"
      ? (trustCenter?.darkLogo?.downloadUrl ?? trustCenter?.logo?.downloadUrl)
      : trustCenter?.logo?.downloadUrl,
  );
  useRequestAccessCallback();

  return (
    <TrustCenterProvider trustCenter={trustCenter}>
      <div className="grid grid-cols-1 max-w-[1280px] mx-4 pt-6 gap-4 lg:mx-auto lg:gap-10 lg:pt-20 lg:grid-cols-[400px_1fr] lg:items-start ">
        <OrganizationSidebar trustCenter={trustCenter} isAuthenticated={isAuthenticated} />
        <main>
          <Tabs className="mb-8">
            <TabLink to="/overview">{__("Overview")}</TabLink>
            <TabLink to="/documents">{__("Documents")}</TabLink>
            {trustCenter.subprocessorInfo.totalCount > 0
              && <TabLink to="/subprocessors">{__("Subprocessors")}</TabLink>}
            <TabLink to="/updates">{__("Updates")}</TabLink>
          </Tabs>
          <Outlet context={{ trustCenter }} />
        </main>
      </div>

      <a
        href="https://www.probo.com/"
        className="flex gap-2 text-sm font-medium text-txt-tertiary items-center w-max mx-auto my-10"
      >
        {__("Powered by")}
        {" "}
        <Logo withPicto className="h-6" />
      </a>
    </TrustCenterProvider>
  );
}
