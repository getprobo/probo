import { useSystemTheme } from "@probo/hooks";
import { Card, Logo } from "@probo/ui";
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
            : <Logo withPicto className="w-[440px]" />}
          <div className="w-full border-t border-t-border-mid" />
        </div>

        <Outlet />
      </Card>
    </div>
  );
}
