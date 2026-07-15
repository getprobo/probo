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
import { Button, Field, IconChevronLeft, useToast } from "@probo/ui";
import { type FormEventHandler, useEffect, useState } from "react";
import {
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { SSOSignInPageQuery } from "#/__generated__/iam/SSOSignInPageQuery.graphql";
import { usePostAuthRedirectUrl } from "#/hooks/usePostAuthRedirectUrl";

const ssoAvailabilityQuery = graphql`
  query SSOSignInPageQuery($email: EmailAddr!) {
    ssoLoginURL(email: $email) @catch(to: RESULT)
  }
`;

export default function SSOSignInPage() {
  const location = useLocation();
  const { __ } = useTranslate();

  const [queryRef, loadQuery]
    = useQueryLoader<SSOSignInPageQuery>(ssoAvailabilityQuery);
  const [checking, setChecking] = useState(false);

  const handleSSOCheck: FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    setChecking(true);
    const formData = new FormData(e.currentTarget);
    const email = formData.get("email") ? (formData.get("email") as string).toString() : "";

    if (!email) return;

    loadQuery({ email }, { fetchPolicy: "network-only" });
  };

  return (
    <>
      <form className="space-y-6 w-full max-w-md mx-auto pt-4" onSubmit={handleSSOCheck}>
        <Link
          to={{ pathname: "/auth/login", search: location.search }}
          className="flex items-center gap-2 text-txt-secondary hover:text-txt-primary transition-colors mb-4"
        >
          <IconChevronLeft size={20} />
          <span className="text-sm">{__("Back")}</span>
        </Link>

        <h1 className="text-center text-2xl font-bold">
          {__("Login with SSO")}
        </h1>
        <p className="text-center text-txt-tertiary mt-1 mb-6">
          {__("Enter your work email to continue with SSO")}
        </p>

        <Field
          required
          placeholder={__("Work Email")}
          name="email"
          type="email"
          label={__("Work Email")}
          autoFocus
        />

        <Button className="w-xs h-10 mx-auto mt-6" disabled={checking}>
          {checking ? __("Checking...") : __("Continue with SSO")}
        </Button>

        <div className="text-center mt-6 text-sm text-txt-secondary">
          {__("Don't have an account ?")}
          {" "}
          <Link
            to={{ pathname: "/auth/register", search: location.search }}
            className="underline hover:text-txt-primary"
          >
            {__("Register")}
          </Link>
        </div>
      </form>

      {queryRef && (
        <NavigateToSSOLoginURL
          onSSOAvailabilityCheck={setChecking}
          queryRef={queryRef}
          loginSearch={location.search}
        />
      )}
    </>
  );
}

function NavigateToSSOLoginURL(props: {
  queryRef: PreloadedQuery<SSOSignInPageQuery>;
  onSSOAvailabilityCheck: (checking: boolean) => void;
  loginSearch: string;
}) {
  const { queryRef, loginSearch } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const postAuthRedirectUrl = usePostAuthRedirectUrl();

  const { ssoLoginURL } = usePreloadedQuery<SSOSignInPageQuery>(
    ssoAvailabilityQuery,
    queryRef,
  );

  useEffect(() => {
    if (!ssoLoginURL.ok) {
      toast({
        title: __("Error"),
        description:
          ssoLoginURL.errors[0] instanceof Error
            ? ssoLoginURL.errors[0].message
            : __("SSO not available for this email domain"),
        variant: "error",
      });

      void navigate({ pathname: "/auth/login", search: loginSearch });
      return;
    }

    if (!ssoLoginURL.value) {
      toast({
        title: __("Error"),
        description: __("SSO not available for this email domain"),
        variant: "error",
      });
      return;
    }

    const url = new URL(ssoLoginURL.value);
    url.searchParams.set("continue", postAuthRedirectUrl);
    for (const [key, value] of searchParams.entries()) {
      if (key !== "continue") {
        url.searchParams.set(key, value);
      }
    }

    window.location.href = url.toString();
  }, [__, loginSearch, navigate, postAuthRedirectUrl, searchParams, ssoLoginURL, toast]);

  return null;
}
