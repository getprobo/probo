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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link, useLocation, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { SignInPageQuery } from "#/__generated__/iam/SignInPageQuery.graphql";
import { usePostAuthRedirectUrl } from "#/hooks/usePostAuthRedirectUrl";
import { isOAuthAuthorizeContinueUrl } from "#/lib/buildAuthorizeContinueURL";

import { Divider } from "./_components/Divider";
import { MagicLinkForm } from "./_components/MagicLinkForm";
import { OAuthClientBrandingSection } from "./_components/OAuthClientBrandingSection";
import { OIDCButton } from "./_components/OIDCButton";

export const signInPageQuery = graphql`
  query SignInPageQuery($clientId: String) {
    oidcProviders {
      ...OIDCButtonFragment
    }
    oauthClientBranding(clientId: $clientId) {
      name
      clientURL
      logo {
        downloadUrl
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<SignInPageQuery>;
};

export default function SignInPage(props: Props) {
  const { __ } = useTranslate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const postAuthRedirectUrl = usePostAuthRedirectUrl();

  const continueParam = searchParams.get("continue");
  const isAuthorizeFlow = isOAuthAuthorizeContinueUrl(continueParam);

  const data = usePreloadedQuery<SignInPageQuery>(signInPageQuery, props.queryRef);

  const clientBranding = data.oauthClientBranding;
  const authorizeHeading = clientBranding?.name
    ? __("Sign in")
    : __("Sign in to continue");

  usePageTitle(
    isAuthorizeFlow
      ? clientBranding?.name
        ? `${__("Sign in to")} ${clientBranding.name}`
        : authorizeHeading
      : __("Sign in to your account"),
  );

  const oidcContinueURL = isAuthorizeFlow ? postAuthRedirectUrl : undefined;

  return (
    <div className="w-full max-w-sm mx-auto pt-8 space-y-6">
      {isAuthorizeFlow && clientBranding && (
        <>
          <OAuthClientBrandingSection
            name={clientBranding.name}
            logoDownloadUrl={clientBranding.logo?.downloadUrl}
            clientURL={clientBranding.clientURL}
          />
          <div className="w-full border-t border-t-border-mid" />
        </>
      )}

      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold">
          {isAuthorizeFlow
            ? authorizeHeading
            : __("Sign in to your account")}
        </h1>
        {isAuthorizeFlow && (
          <p className="text-txt-tertiary">
            {__("Use your email or a connected account to continue")}
          </p>
        )}
      </div>

      <div className="space-y-4">
        {data.oidcProviders.map((providerRef, index) => (
          <OIDCButton
            key={index}
            providerRef={providerRef}
            continueURL={oidcContinueURL}
          />
        ))}

        <MagicLinkForm />

        <Divider>{__("Or")}</Divider>

        <Button
          variant="secondary"
          className="w-full h-10"
          to={{ pathname: "/auth/sso-login", search: location.search }}
        >
          {__("Sign in with SSO")}
        </Button>

        <Button
          variant="secondary"
          className="w-full h-10"
          to={{ pathname: "/auth/password-login", search: location.search }}
        >
          {__("Sign in with password")}
        </Button>
      </div>

      <p className="text-center text-sm text-txt-secondary">
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
