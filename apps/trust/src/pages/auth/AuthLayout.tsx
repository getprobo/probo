import { useSystemTheme } from "@probo/hooks";
import { Logo } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { AuthLayoutQuery } from "./__generated__/AuthLayoutQuery.graphql";

export const authLayoutQuery = graphql`
  query AuthLayoutQuery {
    currentTrustCenter @required(action: THROW) {
      logoFileUrl
      darkLogoFileUrl
    }
  }
`;

export function AuthLayout(props: { queryRef: PreloadedQuery<AuthLayoutQuery> }) {
  const { queryRef } = props;

  const { currentTrustCenter: compliancePage } = usePreloadedQuery<AuthLayoutQuery>(authLayoutQuery, queryRef);
  const theme = useSystemTheme();

  const logoFileUrl = theme === "dark"
    ? compliancePage.darkLogoFileUrl ?? compliancePage.logoFileUrl
    : compliancePage.logoFileUrl;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 min-h-screen text-txt-primary">
      <div className="bg-level-0 flex flex-col items-center justify-center">
        <div className="w-full max-w-md px-6">
          <Outlet />
        </div>
      </div>
      <div className="hidden lg:flex bg-dialog font-bold flex-col items-center justify-center p-8 text-txt-primary lg:p-10">
        <div className="flex flex-col items-center justify-center gap-4">
          {logoFileUrl
            ? (
                <img
                  alt=""
                  src={logoFileUrl}
                  className="size-[440px] rounded-2xl"
                />
              )
            : <Logo withPicto className="w-[440px]" />}
        </div>
      </div>
    </div>
  );
}
