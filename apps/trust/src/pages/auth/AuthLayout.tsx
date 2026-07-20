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

import { useSystemTheme } from "@probo/hooks";
import { Card, Logo } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { AuthLayoutQuery } from "./__generated__/AuthLayoutQuery.graphql";

export const authLayoutQuery = graphql`
  query AuthLayoutQuery {
    currentCompliancePortal @required(action: THROW) {
      logo {
        downloadUrl
      }
      darkLogo {
        downloadUrl
      }
    }
  }
`;

export function AuthLayout(props: { queryRef: PreloadedQuery<AuthLayoutQuery> }) {
  const { queryRef } = props;

  const { currentCompliancePortal: compliancePage } = usePreloadedQuery<AuthLayoutQuery>(authLayoutQuery, queryRef);
  const theme = useSystemTheme();

  const logoFileUrl = theme === "dark"
    ? compliancePage.darkLogo?.downloadUrl ?? compliancePage.logo?.downloadUrl
    : compliancePage.logo?.downloadUrl;

  return (
    <div className="min-h-screen text-txt-primary bg-level-0 flex flex-col items-center justify-center">
      <Card className="w-full max-w-lg px-12 py-8 flex flex-col items-center justify-center">
        <div className="w-full flex flex-col items-center justify-center gap-8">
          {logoFileUrl
            ? (
                <img
                  alt=""
                  src={logoFileUrl}
                  className="h-20 rounded-2xl"
                />
              )
            : <Logo withPicto className="w-[110px]" />}
          <div className="w-full border-t border-t-border-mid" />
        </div>

        <Outlet />
      </Card>
    </div>
  );
}
