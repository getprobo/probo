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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { ProboLogo } from "@probo/ui/src/v2/ProboLogo/ProboLogo";
import type { PropsWithChildren } from "react";
import { useTranslation } from "react-i18next";
import { matchPath, Outlet, useLocation, useSearchParams } from "react-router";

import { isOAuthAuthorizeContinueUrl } from "#/lib/buildAuthorizeContinueURL";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { authLayout } from "./_components/variants";

export default function AuthLayout(props: PropsWithChildren) {
  const { children } = props;
  const { t } = useTranslation();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const isAuthorizeFlow = isOAuthAuthorizeContinueUrl(searchParams.get("continue"));
  const showBack
    = matchPath("/auth/password-login", location.pathname) != null
      || matchPath("/auth/sso-login", location.pathname) != null;
  const slots = authLayout();

  return (
    <div className="flex min-h-dvh flex-col items-center justify-center bg-sand-2 p-4">
      <div className={slots.column()}>
        {(!isAuthorizeFlow || showBack) && (
          <div className={slots.header()}>
            {!isAuthorizeFlow && (
              <ProboLogo className="h-6 w-auto text-sand-12" />
            )}
            {showBack && (
              <Link
                to={{ pathname: "/auth/login", search: location.search }}
                iconStart={<CaretLeftIcon />}
                className={slots.back()}
              >
                {t("common.actions.back")}
              </Link>
            )}
          </div>
        )}
        <div className={slots.stack()}>
          <div className={slots.wash()} />
          <Card variant="soft" size={3} padding="none" className={slots.frame()}>
            <div className={slots.body()}>
              <IAMRelayProvider>
                {children ?? <Outlet />}
              </IAMRelayProvider>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}
