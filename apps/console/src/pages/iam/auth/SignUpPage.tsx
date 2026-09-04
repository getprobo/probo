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
import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation, usePreloadedQuery, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";

import type { SignUpPageMutation } from "#/__generated__/iam/SignUpPageMutation.graphql";
import type { SignUpPageQuery } from "#/__generated__/iam/SignUpPageQuery.graphql";

const signUpPageQuery = graphql`
  query SignUpPageQuery {
    signUpEnabled
  }
`;

const signUpMutation = graphql`
  mutation SignUpPageMutation($input: SignUpInput!) {
    signUp(input: $input) {
      identity {
        id
      }
    }
  }
`;

function SignUpPageContent(props: { queryRef: NonNullable<ReturnType<typeof useQueryLoader<SignUpPageQuery>>[0]> }) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [signedUpEmail, setSignedUpEmail] = useState("");

  usePageTitle(t("signUpPage.pageTitle"));

  const data = usePreloadedQuery<SignUpPageQuery>(signUpPageQuery, props.queryRef);

  const [signUp, isSigningUp] = useMutation<SignUpPageMutation>(signUpMutation);

  const onSubmit = (email: string, password: string, fullName: string) => {
    signUp({
      variables: {
        input: {
          email,
          password,
          fullName,
        },
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: t("signUpPage.errors.failed"),
            description: formatError(t("signUpPage.errors.failed"), e),
            variant: "error",
          });
          return;
        }

        setSignedUpEmail(email);
        toast({
          title: t("common.success"), description: t("signUpPage.messages.created"),
          variant: "success",
        });
      },
      onError: (e) => {
        toast({
          title: t("signUpPage.errors.failed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  if (signedUpEmail !== "") {
    const resendSearch = new URLSearchParams({ email: signedUpEmail }).toString();

    return (
      <div className="flex w-full flex-col gap-8">
        <div className="flex flex-col gap-1">
          <Heading level={1} size={4} weight="medium" align="center" highContrast>
            {t("signUpPage.checkEmail.title")}
          </Heading>
          <Text size={2} align="center" className="block">
            {t("signUpPage.checkEmail.description", { email: signedUpEmail })}
          </Text>
        </div>

        <Text align="center" size={2} className="block">
          {t("signUpPage.checkEmail.resend")}
          {" "}
          <Link to={{ pathname: "/auth/resend-verification-email", search: `?${resendSearch}` }}>
            {t("signUpPage.actions.resend")}
          </Link>
        </Text>

        <Text align="center" size={2} className="block">
          <Link to="/auth/login">
            {t("signUpPage.actions.backToLogin")}
          </Link>
        </Text>
      </div>
    );
  }

  if (!data.signUpEnabled) {
    return (
      <div className="flex w-full flex-col gap-8">
        <div className="flex flex-col gap-1">
          <Heading level={1} size={4} weight="medium" align="center" highContrast>
            {t("signUpPage.unavailable.title")}
          </Heading>
          <Text size={2} align="center" className="block">
            {t("signUpPage.unavailable.description")}
          </Text>
        </div>

        <ButtonLink
          to="/auth/login"
          variant="soft"
          color="neutral"
          size={3}
          className="w-full"
        >
          {t("signUpPage.actions.backToLogin")}
        </ButtonLink>
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("signUpPage.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("signUpPage.description")}
        </Text>
      </div>

      <Form
        className="flex flex-col gap-5"
        onFormSubmit={(values) => {
          onSubmit(
            String(values.email ?? ""),
            String(values.password ?? ""),
            String(values.fullName ?? ""),
          );
        }}
      >
        <Field.Root name="fullName" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("signUpPage.fields.fullName")}
          </Field.Label>
          <TextField
            type="text"
            name="fullName"
            required
            minLength={2}
            placeholder={t("signUpPage.fields.fullNamePlaceholder")}
          />
          <Field.Error className="text-1 text-red-11" />
        </Field.Root>

        <Field.Root name="email" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("signUpPage.fields.email")}
          </Field.Label>
          <TextField
            type="email"
            name="email"
            required
            placeholder={t("signUpPage.fields.emailPlaceholder")}
          />
          <Field.Error className="text-1 text-red-11" />
        </Field.Root>

        <Field.Root name="password" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("signUpPage.fields.password")}
          </Field.Label>
          <TextField
            type="password"
            name="password"
            required
            minLength={8}
            placeholder="••••••••"
          />
          <Field.Error className="text-1 text-red-11" />
        </Field.Root>

        <Button
          type="submit"
          variant="solid"
          color="neutral"
          highContrast
          size={3}
          className="w-full"
          loading={isSigningUp}
        >
          {t("signUpPage.actions.signUpWithEmail")}
        </Button>
      </Form>

      <Text align="center" size={2} className="block">
        {t("signUpPage.alreadyHaveAccount")}
        {" "}
        <Link to="/auth/login">
          {t("signUpPage.actions.logIn")}
        </Link>
      </Text>
    </div>
  );
}

export default function SignUpPage() {
  const [queryRef, loadQuery] = useQueryLoader<SignUpPageQuery>(signUpPageQuery);

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) return null;

  return <SignUpPageContent queryRef={queryRef} />;
}
