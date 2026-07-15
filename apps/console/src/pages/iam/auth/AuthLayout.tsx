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

import { Card, Logo } from "@probo/ui";
import type { PropsWithChildren } from "react";
import { Outlet, useSearchParams } from "react-router";

import { isOAuthAuthorizeContinueUrl } from "#/lib/buildAuthorizeContinueURL";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

export default function AuthLayout(props: PropsWithChildren) {
  const { children } = props;
  const [searchParams] = useSearchParams();
  const isAuthorizeFlow = isOAuthAuthorizeContinueUrl(searchParams.get("continue"));

  return (
    <div className="min-h-screen text-txt-primary bg-level-0 flex flex-col items-center justify-center">
      <Card className="w-full max-w-lg px-12 py-8 flex flex-col items-center justify-center">
        <div className="w-full flex flex-col items-center justify-center gap-8">
          {!isAuthorizeFlow && (
            <>
              <Logo withPicto className="w-[110px]" />
              <div className="w-full border-t border-t-border-mid" />
            </>
          )}
        </div>
        <IAMRelayProvider>
          {children ?? <Outlet />}
        </IAMRelayProvider>
      </Card>
    </div>
  );
}
