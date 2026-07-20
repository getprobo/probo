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
import { Outlet, useMatch } from "react-router";

import { PoweredBy } from "#/components/PoweredBy/PoweredBy";
import { TopBar } from "#/components/TopBar/TopBar";
import { SignInDialogProvider } from "#/lib/auth/SignInDialogProvider";
import { useResumeAccessRequest } from "#/lib/auth/useResumeAccessRequest";
import { SubscribeDialogProvider } from "#/lib/mailingList/SubscribeDialogProvider";

import type { MainLayoutQuery } from "./__generated__/MainLayoutQuery.graphql";

export const mainLayoutQuery = graphql`
  query MainLayoutQuery {
    viewer {
      __typename
    }
    ...TopBar_query
    ...SubscribeDialogProvider_query
  }
`;

interface MainLayoutProps {
  queryRef: PreloadedQuery<MainLayoutQuery>;
}

export function MainLayout({ queryRef }: MainLayoutProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<MainLayoutQuery>(mainLayoutQuery, queryRef);
  const isDocumentViewer = useMatch("documents/:alias") != null;

  // Resume a deferred "request access" once the user lands back authenticated.
  useResumeAccessRequest(data.viewer != null);

  // Document viewer fills the viewport under the TopBar and scrolls its own
  // stage; every other page uses normal document flow so the footer sits after
  // content (and at the bottom of short pages via flex-1 main).
  return (
    <SignInDialogProvider>
      <SubscribeDialogProvider queryKey={data}>
        <div
          className={
            isDocumentViewer
              ? "flex h-dvh flex-col bg-sand-2"
              : "flex min-h-dvh flex-col bg-sand-2"
          }
        >
          <TopBar queryKey={data} />
          <div
            className={
              isDocumentViewer
                ? "flex min-h-0 flex-1 flex-col overflow-hidden"
                : "flex flex-1 flex-col"
            }
          >
            <Outlet />
          </div>
          {isDocumentViewer ? null : <PoweredBy label={t("footer.poweredBy")} />}
        </div>
      </SubscribeDialogProvider>
    </SignInDialogProvider>
  );
}
