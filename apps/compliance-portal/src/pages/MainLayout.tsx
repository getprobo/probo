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

import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import { PoweredBy } from "#/components/PoweredBy/PoweredBy";
import { TopBar } from "#/components/TopBar/TopBar";

import type { MainLayoutQuery } from "./__generated__/MainLayoutQuery.graphql";

export const mainLayoutQuery = graphql`
  query MainLayoutQuery {
    ...TopBar_query
  }
`;

interface MainLayoutProps {
  queryRef: PreloadedQuery<MainLayoutQuery>;
}

export function MainLayout({ queryRef }: MainLayoutProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<MainLayoutQuery>(mainLayoutQuery, queryRef);

  return (
    <div className="flex min-h-screen flex-col bg-sand-2">
      <TopBar queryKey={data} />
      <div className="flex-1">
        <Outlet />
      </div>
      <PoweredBy label={t("footer.poweredBy")} />
    </div>
  );
}
