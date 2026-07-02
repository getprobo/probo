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
