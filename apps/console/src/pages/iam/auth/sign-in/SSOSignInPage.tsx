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

import { Button, Field, IconChevronLeft, useToast } from "@probo/ui";
import { type FormEventHandler, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();

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
          <span className="text-sm">{t("ssoSignInPage.actions.back")}</span>
        </Link>

        <h1 className="text-center text-2xl font-bold">
          {t("ssoSignInPage.title")}
        </h1>
        <p className="text-center text-txt-tertiary mt-1 mb-6">
          {t("ssoSignInPage.description")}
        </p>

        <Field
          required
          placeholder={t("ssoSignInPage.fields.workEmail")}
          name="email"
          type="email"
          label={t("ssoSignInPage.fields.workEmail")}
          autoFocus
        />

        <Button className="w-xs h-10 mx-auto mt-6" disabled={checking}>
          {checking ? t("ssoSignInPage.actions.checking") : t("ssoSignInPage.actions.continue")}
        </Button>

        <div className="text-center mt-6 text-sm text-txt-secondary">
          {t("ssoSignInPage.noAccount")}
          {" "}
          <Link
            to={{ pathname: "/auth/register", search: location.search }}
            className="underline hover:text-txt-primary"
          >
            {t("ssoSignInPage.actions.register")}
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

  const { t } = useTranslation();
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
        title: t("common.error"),
        description:
          ssoLoginURL.errors[0] instanceof Error
            ? ssoLoginURL.errors[0].message
            : t("ssoSignInPage.errors.unavailable"),
        variant: "error",
      });

      void navigate({ pathname: "/auth/login", search: loginSearch });
      return;
    }

    if (!ssoLoginURL.value) {
      toast({
        title: t("common.error"), description: t("ssoSignInPage.errors.unavailable"),
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
  }, [t, loginSearch, navigate, postAuthRedirectUrl, searchParams, ssoLoginURL, toast]);

  return null;
}
