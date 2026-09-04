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
import { graphql } from "relay-runtime";

import type { ForgotPasswordPageMutation } from "#/__generated__/iam/ForgotPasswordPageMutation.graphql";

const sendInstructionsMutation = graphql`
  mutation ForgotPasswordPageMutation($input: ForgotPasswordInput!) {
    forgotPassword(input: $input) {
      success
    }
  }
`;

export default function ForgotPasswordPage() {
  const { toast } = useToast();
  const { t } = useTranslation();

  usePageTitle(t("forgotPasswordPage.pageTitle"));

  const [instructionsSent, setInstructionsSent] = useState<boolean>();

  const [sendInstructions, isSendingInstructions]
    = useMutation<ForgotPasswordPageMutation>(sendInstructionsMutation);

  const onSubmit = (email: string) => {
    if (isSendingInstructions) return;

    sendInstructions({
      variables: {
        input: { email },
      },
      onError: (e: Error) => {
        toast({
          title: t("forgotPasswordPage.errors.requestFailed"),
          description: e.message,
          variant: "error",
        });
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: t("forgotPasswordPage.errors.requestFailed"),
            description: formatError(
              t("forgotPasswordPage.errors.sendInstructions"),
              e,
            ),
            variant: "error",
          });
          return;
        }

        toast({
          title: t("common.success"),
          description: t("forgotPasswordPage.messages.instructionsSent"),
          variant: "success",
        });
        setInstructionsSent(true);
      },
    });
  };

  if (instructionsSent) {
    return (
      <div className="flex w-full flex-col gap-8">
        <div className="flex flex-col gap-1">
          <Heading level={1} size={4} weight="medium" align="center" highContrast>
            {t("forgotPasswordPage.sent.title")}
          </Heading>
          <Text size={2} align="center" className="block">
            {t("forgotPasswordPage.sent.description")}
          </Text>
        </div>

        <Text align="center" size={2} className="block">
          {t("forgotPasswordPage.sent.didNotReceive")}
          {" "}
          <button
            type="button"
            onClick={() => setInstructionsSent(false)}
            className="underline"
          >
            {t("forgotPasswordPage.actions.tryAgain")}
          </button>
        </Text>

        <Text align="center" size={2} className="block">
          {t("forgotPasswordPage.rememberPassword")}
          {" "}
          <Link to="/auth/login">
            {t("forgotPasswordPage.actions.backToLogin")}
          </Link>
        </Text>
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("forgotPasswordPage.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("forgotPasswordPage.description")}
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
            {t("forgotPasswordPage.fields.email")}
          </Field.Label>
          <TextField
            type="email"
            name="email"
            required
            placeholder={t("forgotPasswordPage.fields.emailPlaceholder")}
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
          loading={isSendingInstructions}
        >
          {t("forgotPasswordPage.actions.sendInstructions")}
        </Button>
      </Form>

      <Text align="center" size={2} className="block">
        {t("forgotPasswordPage.rememberPassword")}
        {" "}
        <Link to="/auth/login">
          {t("forgotPasswordPage.actions.backToLogin")}
        </Link>
      </Text>
    </div>
  );
}
