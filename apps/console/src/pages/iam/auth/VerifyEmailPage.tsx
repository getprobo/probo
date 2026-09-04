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
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { VerifyEmailPageMutation } from "#/__generated__/iam/VerifyEmailPageMutation.graphql";

const verifyEmailMutation = graphql`
  mutation VerifyEmailPageMutation($input: VerifyEmailInput!) {
    verifyEmail(input: $input) {
      success
    }
  }
`;

export default function VerifyEmailPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();

  usePageTitle(t("verifyEmailPage.pageTitle"));

  const [isConfirmed, setIsConfirmed] = useState(false);

  const [verifyEmail, isVerifying]
    = useMutation<VerifyEmailPageMutation>(verifyEmailMutation);

  const handleSubmit = (token: string) => {
    verifyEmail({
      variables: {
        input: {
          token: token.trim(),
        },
      },
      onCompleted: (_, errors) => {
        if (errors) {
          toast({
            title: t("common.error"), description: formatError(t("verifyEmailPage.errors.confirm"), errors),
            variant: "error",
          });
          return;
        }

        setIsConfirmed(true);
        toast({
          title: t("common.success"), description: t("verifyEmailPage.messages.confirmed"),
          variant: "success",
        });
      },
      onError: (err) => {
        toast({
          title: t("common.error"), description: err.message || t("verifyEmailPage.errors.confirm"),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("verifyEmailPage.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("verifyEmailPage.description")}
        </Text>
      </div>

      {isConfirmed
        ? (
            <div className="flex flex-col gap-5">
              <Text align="center" color="green" highContrast className="block">
                {t("verifyEmailPage.messages.confirmedWithExclamation")}
              </Text>
              <ButtonLink
                to="/auth/login"
                variant="solid"
                color="neutral"
                highContrast
                size={3}
                className="w-full"
              >
                {t("verifyEmailPage.actions.proceedToLogin")}
              </ButtonLink>
            </div>
          )
        : (
            <Form
              className="flex flex-col gap-5"
              onFormSubmit={(values) => {
                handleSubmit(String(values.token ?? ""));
              }}
            >
              <Field.Root name="token" className="flex flex-col gap-1.5">
                <Field.Label className="text-1 font-medium text-sand-12">
                  {t("verifyEmailPage.fields.token")}
                </Field.Label>
                <TextField
                  type="text"
                  name="token"
                  required
                  defaultValue={searchParams.get("token") ?? ""}
                  placeholder={t("verifyEmailPage.fields.tokenPlaceholder")}
                  disabled={isVerifying}
                />
                <Field.Description className="text-1 text-sand-11">
                  {t("verifyEmailPage.fields.tokenHelp")}
                </Field.Description>
                <Field.Error className="text-1 text-red-11" />
              </Field.Root>

              <Button
                type="submit"
                variant="solid"
                color="neutral"
                highContrast
                size={3}
                className="w-full"
                loading={isVerifying}
              >
                {t("verifyEmailPage.actions.confirm")}
              </Button>
            </Form>
          )}

      {!isConfirmed && (
        <Text align="center" size={2} className="block">
          <Link to="/auth/login">
            {t("verifyEmailPage.actions.backToLogin")}
          </Link>
        </Text>
      )}
    </div>
  );
}
