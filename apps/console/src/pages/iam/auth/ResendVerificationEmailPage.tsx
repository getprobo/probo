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
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { ResendVerificationEmailPageMutation } from "#/__generated__/iam/ResendVerificationEmailPageMutation.graphql";

function emailFromLocationState(state: unknown): string {
  if (state == null || typeof state !== "object" || !("email" in state)) {
    return "";
  }

  const { email } = state;
  return typeof email === "string" ? email : "";
}

const resendVerificationEmailMutation = graphql`
  mutation ResendVerificationEmailPageMutation($input: ResendVerificationEmailInput!) {
    resendVerificationEmail(input: $input) {
      success
    }
  }
`;

export default function ResendVerificationEmailPage() {
  const { toast } = useToast();
  const { t } = useTranslation();
  const location = useLocation();
  const defaultEmail = emailFromLocationState(location.state);

  usePageTitle(t("resendVerificationEmailPage.pageTitle"));

  const [emailSent, setEmailSent] = useState<boolean>();

  const [resendVerificationEmail, isResending]
    = useMutation<ResendVerificationEmailPageMutation>(resendVerificationEmailMutation);

  const onSubmit = (email: string) => {
    if (isResending) return;

    resendVerificationEmail({
      variables: {
        input: { email },
      },
      onError: (e: Error) => {
        toast({
          title: t("resendVerificationEmailPage.errors.requestFailed"),
          description: e.message,
          variant: "error",
        });
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: t("resendVerificationEmailPage.errors.requestFailed"),
            description: formatError(
              t("resendVerificationEmailPage.errors.sendVerification"),
              e,
            ),
            variant: "error",
          });
          return;
        }

        toast({
          title: t("common.success"),
          description: t("resendVerificationEmailPage.messages.verificationSent"),
          variant: "success",
        });
        setEmailSent(true);
      },
    });
  };

  if (emailSent) {
    return (
      <div className="flex w-full flex-col gap-8">
        <div className="flex flex-col gap-1">
          <Heading level={1} size={4} weight="medium" align="center" highContrast>
            {t("resendVerificationEmailPage.sent.title")}
          </Heading>
          <Text size={2} align="center" className="block">
            {t("resendVerificationEmailPage.sent.description")}
          </Text>
        </div>

        <Text align="center" size={2} className="block">
          {t("resendVerificationEmailPage.sent.didNotReceive")}
          {" "}
          <button
            type="button"
            onClick={() => setEmailSent(false)}
            className="underline"
          >
            {t("resendVerificationEmailPage.actions.tryAgain")}
          </button>
        </Text>

        <Text align="center" size={2} className="block">
          {t("resendVerificationEmailPage.alreadyVerified")}
          {" "}
          <Link to="/auth/login">
            {t("resendVerificationEmailPage.actions.backToLogin")}
          </Link>
        </Text>
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("resendVerificationEmailPage.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("resendVerificationEmailPage.description")}
        </Text>
      </div>

      <Form
        className="flex flex-col gap-5"
        onFormSubmit={(values) => {
          onSubmit(String(values.email ?? ""));
        }}
      >
        <Field.Root name="email" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("resendVerificationEmailPage.fields.email")}
          </Field.Label>
          <TextField
            type="email"
            name="email"
            required
            defaultValue={defaultEmail}
            placeholder={t("resendVerificationEmailPage.fields.emailPlaceholder")}
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
          loading={isResending}
        >
          {t("resendVerificationEmailPage.actions.sendVerification")}
        </Button>
      </Form>

      <Text align="center" size={2} className="block">
        {t("resendVerificationEmailPage.alreadyVerified")}
        {" "}
        <Link to="/auth/login">
          {t("resendVerificationEmailPage.actions.backToLogin")}
        </Link>
      </Text>
    </div>
  );
}
