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
import { formatError, type GraphQLError } from "@probo/helpers";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery, useMutation } from "react-relay";
import { matchPath, useLocation, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { PasswordSignInPageMutation } from "#/__generated__/iam/PasswordSignInPageMutation.graphql";
import type { PasswordSignInPageQuery } from "#/__generated__/iam/PasswordSignInPageQuery.graphql";
import { usePostAuthRedirectUrl } from "#/hooks/usePostAuthRedirectUrl";

import { CreateAccountFooter } from "./_components/CreateAccountFooter";

const signInMutation = graphql`
  mutation PasswordSignInPageMutation($input: SignInInput!) {
    signIn(input: $input) {
      session {
        id
      }
    }
  }
`;

const passwordSignInPageQuery = graphql`
  query PasswordSignInPageQuery {
    ...CreateAccountFooterFragment
  }
`;

function PasswordCreateAccountFooter() {
  const { t } = useTranslation();
  const data = useLazyLoadQuery<PasswordSignInPageQuery>(
    passwordSignInPageQuery,
    {},
  );

  return (
    <CreateAccountFooter
      queryKey={data}
      prefix={t("passwordSignInPage.noAccount")}
      label={t("passwordSignInPage.actions.register")}
    />
  );
}

export default function PasswordSignInPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const postAuthRedirectUrl = usePostAuthRedirectUrl();

  const { t } = useTranslation();
  const { toast } = useToast();

  const [signIn, isSigningIn]
    = useMutation<PasswordSignInPageMutation>(signInMutation);

  const handlePasswordLogin = (emailValue: string, passwordValue: string) => {
    const match = matchPath(
      { path: "/organizations/:organizationId", caseSensitive: false, end: false },
      new URL(postAuthRedirectUrl, window.location.origin).pathname,
    );
    const organizationId
      = match?.params.organizationId ?? searchParams.get("organization-id") ?? undefined;

    signIn({
      variables: {
        input: {
          email: emailValue,
          password: passwordValue,
          organizationId,
        },
      },
      onCompleted: (_, error) => {
        if (error) {
          const errors = Array.isArray(error) ? error : [error];
          const emailNotVerified = errors.some(
            e => (e as GraphQLError).extensions?.code === "EMAIL_NOT_VERIFIED",
          );
          if (emailNotVerified) {
            const loginSearch = searchParams.toString();
            void navigate(
              {
                pathname: "/auth/resend-verification-email",
                search: loginSearch === "" ? "" : "?" + loginSearch,
              },
              { state: { email: emailValue } },
            );
            return;
          }

          toast({
            title: t("common.error"),
            description: formatError(
              t("passwordSignInPage.errors.login"),
              error,
            ),
            variant: "error",
          });
          return;
        }

        window.location.href = postAuthRedirectUrl;
      },
      onError: (e) => {
        toast({
          title: t("common.error"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("passwordSignInPage.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("passwordSignInPage.description")}
        </Text>
      </div>

      <Form
        className="flex flex-col gap-5"
        onFormSubmit={(values) => {
          handlePasswordLogin(String(values.email ?? ""), String(values.password ?? ""));
        }}
      >
        <Field.Root name="email" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("passwordSignInPage.fields.email")}
          </Field.Label>
          <TextField type="email" name="email" required autoFocus />
          <Field.Error className="text-1 text-red-11" />
        </Field.Root>

        <Field.Root name="password" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("passwordSignInPage.fields.password")}
          </Field.Label>
          <TextField type="password" name="password" required />
          <Field.Error className="text-1 text-red-11" />
          <Text align="right" size={2} className="block">
            {t("passwordSignInPage.forgotPassword")}
            {" "}
            <Link to="/auth/forgot-password">
              {t("passwordSignInPage.actions.resetPassword")}
            </Link>
          </Text>
        </Field.Root>

        <Button
          type="submit"
          variant="solid"
          color="neutral"
          highContrast
          size={3}
          className="w-full"
          loading={isSigningIn}
        >
          {t("passwordSignInPage.actions.login")}
        </Button>
      </Form>

      <div className="flex flex-col gap-2">
        <Suspense fallback={null}>
          <PasswordCreateAccountFooter />
        </Suspense>

        <Text align="center" size={2} className="block">
          <Link to={{ pathname: "/auth/login", search: location.search }}>
            {t("passwordSignInPage.actions.backToLogin")}
          </Link>
        </Text>
      </div>
    </div>
  );
}
