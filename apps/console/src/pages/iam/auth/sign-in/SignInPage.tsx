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
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation, useSearchParams } from "react-router";
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
      logoUrl
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<SignInPageQuery>;
};

export default function SignInPage(props: Props) {
  const { t } = useTranslation();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const postAuthRedirectUrl = usePostAuthRedirectUrl();

  const continueParam = searchParams.get("continue");
  const isAuthorizeFlow = isOAuthAuthorizeContinueUrl(continueParam);

  const data = usePreloadedQuery<SignInPageQuery>(signInPageQuery, props.queryRef);

  const clientBranding = data.oauthClientBranding;
  const authorizeHeading = clientBranding?.name
    ? t("auth.actions.signIn")
    : t("signInPage.authorize.title");

  usePageTitle(
    isAuthorizeFlow
      ? clientBranding?.name
        ? t("signInPage.authorize.titleWithClient", { name: clientBranding.name })
        : authorizeHeading
      : t("signInPage.title"),
  );

  const oidcContinueURL = isAuthorizeFlow ? postAuthRedirectUrl : undefined;

  return (
    <div className="flex w-full flex-col gap-8">
      {isAuthorizeFlow && clientBranding && (
        <OAuthClientBrandingSection
          name={clientBranding.name}
          logoDownloadUrl={clientBranding.logoUrl}
          clientURL={clientBranding.clientURL}
        />
      )}

      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {isAuthorizeFlow
            ? authorizeHeading
            : t("signInPage.title")}
        </Heading>
        {isAuthorizeFlow && (
          <Text size={2} align="center" className="block">
            {t("signInPage.authorize.description")}
          </Text>
        )}
      </div>

      <div className="flex flex-col gap-5">
        <MagicLinkForm />

        {data.oidcProviders.length > 0 && (
          <>
            <Divider>{t("signInPage.or")}</Divider>
            {data.oidcProviders.map((providerRef, index) => (
              <OIDCButton
                key={index}
                providerRef={providerRef}
                continueURL={oidcContinueURL}
              />
            ))}
          </>
        )}

        <Text align="center" size={2} className="block">
          <Link
            to={{ pathname: "/auth/password-login", search: location.search }}
          >
            {t("signInPage.actions.usePassword")}
          </Link>
        </Text>
      </div>

      <Text align="center" size={2} className="block">
        {t("signInPage.newToProbo")}
        {" "}
        <Link
          to={{ pathname: "/auth/register", search: location.search }}
        >
          {t("signInPage.actions.createAccount")}
        </Link>
      </Text>
    </div>
  );
}
