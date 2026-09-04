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

import { Field } from "@base-ui/react/field";
import { Form } from "@base-ui/react/form";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  useLazyLoadQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { useLocation, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { SSOSignInPageCreateAccountQuery } from "#/__generated__/iam/SSOSignInPageCreateAccountQuery.graphql";
import type { SSOSignInPageQuery } from "#/__generated__/iam/SSOSignInPageQuery.graphql";
import { usePostAuthRedirectUrl } from "#/hooks/usePostAuthRedirectUrl";
import { clientIdFromContinueUrl } from "#/lib/buildAuthorizeContinueURL";

import { CreateAccountFooter } from "./_components/CreateAccountFooter";

const ssoAvailabilityQuery = graphql`
  query SSOSignInPageQuery($email: EmailAddr!) {
    ssoLoginURL(email: $email) @catch(to: RESULT)
  }
`;

const ssoCreateAccountQuery = graphql`
  query SSOSignInPageCreateAccountQuery($clientId: String) {
    ...CreateAccountFooterFragment @arguments(clientId: $clientId)
  }
`;

function SSOCreateAccountFooter() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const clientId = clientIdFromContinueUrl(searchParams.get("continue"));
  const data = useLazyLoadQuery<SSOSignInPageCreateAccountQuery>(
    ssoCreateAccountQuery,
    { clientId },
  );

  return (
    <CreateAccountFooter
      queryKey={data}
      prefix={t("ssoSignInPage.noAccount")}
      label={t("ssoSignInPage.actions.register")}
    />
  );
}

export default function SSOSignInPage() {
  const location = useLocation();
  const { t } = useTranslation();

  const [queryRef, loadQuery]
    = useQueryLoader<SSOSignInPageQuery>(ssoAvailabilityQuery);
  const [checking, setChecking] = useState(false);

  return (
    <>
      <div className="flex w-full flex-col gap-8">
        <div className="flex flex-col gap-1">
          <Heading level={1} size={4} weight="medium" align="center" highContrast>
            {t("ssoSignInPage.title")}
          </Heading>
          <Text size={2} align="center" className="block">
            {t("ssoSignInPage.description")}
          </Text>
        </div>

        <Form
          className="flex flex-col gap-5"
          onFormSubmit={(values) => {
            const email = String(values.email ?? "");
            if (!email) return;
            setChecking(true);
            loadQuery({ email }, { fetchPolicy: "network-only" });
          }}
        >
          <Field.Root name="email" className="flex flex-col gap-1.5">
            <Field.Label className="text-1 font-medium text-sand-12">
              {t("ssoSignInPage.fields.workEmail")}
            </Field.Label>
            <TextField type="email" name="email" required autoFocus />
            <Field.Error className="text-1 text-red-11" />
          </Field.Root>

          <Button
            type="submit"
            variant="solid"
            color="neutral"
            highContrast
            size={3}
            className="w-full"
            loading={checking}
          >
            {t("ssoSignInPage.actions.continue")}
          </Button>
        </Form>

        <div className="flex flex-col gap-2">
          <Suspense fallback={null}>
            <SSOCreateAccountFooter />
          </Suspense>

          <Text align="center" size={2} className="block">
            <Link to={{ pathname: "/auth/login", search: location.search }}>
              {t("ssoSignInPage.actions.backToLogin")}
            </Link>
          </Text>
        </div>
      </div>

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
