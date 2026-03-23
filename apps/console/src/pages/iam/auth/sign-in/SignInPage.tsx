import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link, useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { SignInPageQuery } from "#/__generated__/iam/SignInPageQuery.graphql";

import { Divider } from "./_components/Divider";
import { OIDCButton } from "./_components/OIDCButton";

export const signInPageQuery = graphql`
  query SignInPageQuery {
    oidcProviders {
      ...OIDCButtonFragment
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<SignInPageQuery>;
};

export default function SignInPage(props: Props) {
  const { __ } = useTranslate();
  const location = useLocation();

  const data = usePreloadedQuery<SignInPageQuery>(signInPageQuery, props.queryRef);

  return (
    <div className="w-full max-w-sm mx-auto pt-8">
      <h1 className="text-2xl font-bold">
        {__("Sign in to your account")}
      </h1>

      <div className="mt-6 space-y-4">
        {data.oidcProviders.map((providerRef, index) => (
          <OIDCButton key={index} providerRef={providerRef} />
        ))}

        <Button
          variant="secondary"
          className="w-full h-10"
          to={{ pathname: "/auth/sso-login", search: location.search }}
        >
          {__("Sign in with SSO")}
        </Button>

        <Divider>{__("Or")}</Divider>

        <Button
          variant="secondary"
          className="w-full h-10"
          to={{ pathname: "/auth/password-login", search: location.search }}
        >
          {__("Sign in with email")}
        </Button>
      </div>

      <p className="mt-8 text-center text-sm text-txt-secondary">
        {__("New to Probo?")}
        {" "}
        <Link
          to={{ pathname: "/auth/register", search: location.search }}
          className="underline hover:text-txt-primary"
        >
          {__("Create account")}
        </Link>
      </p>
    </div>
  );
}
