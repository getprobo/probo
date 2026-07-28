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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Button, Field, useToast } from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { Link, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { ResendVerificationEmailPageMutation } from "#/__generated__/iam/ResendVerificationEmailPageMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const resendVerificationEmailMutation = graphql`
  mutation ResendVerificationEmailPageMutation($input: ResendVerificationEmailInput!) {
    resendVerificationEmail(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  email: z.email(),
});

export default function ResendVerificationEmailPage() {
  const { toast } = useToast();
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();

  usePageTitle(t("resendVerificationEmailPage.pageTitle"));

  const [emailSent, setEmailSent] = useState<boolean>();
  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      email: searchParams.get("email") ?? "",
    },
  });

  const [resendVerificationEmail] = useMutation<ResendVerificationEmailPageMutation>(
    resendVerificationEmailMutation,
  );

  const onSubmit = handleSubmit(({ email }) => {
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
  });

  return emailSent
    ? (
        <div className="space-y-6 w-full max-w-md mx-auto pt-8">
          <div className="space-y-2 text-center">
            <h1 className="text-3xl font-bold">{t("resendVerificationEmailPage.sent.title")}</h1>
            <p className="text-txt-tertiary">
              {t("resendVerificationEmailPage.sent.description")}
            </p>
          </div>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {t("resendVerificationEmailPage.sent.didNotReceive")}
              {" "}
              <button
                onClick={() => setEmailSent(false)}
                className="underline text-txt-primary hover:text-txt-secondary"
              >
                {t("resendVerificationEmailPage.actions.tryAgain")}
              </button>
            </p>
          </div>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {t("resendVerificationEmailPage.alreadyVerified")}
              {" "}
              <Link
                to="/auth/login"
                className="underline text-txt-primary hover:text-txt-secondary"
              >
                {t("resendVerificationEmailPage.actions.backToLogin")}
              </Link>
            </p>
          </div>
        </div>
      )
    : (
        <div className="space-y-6 w-full max-w-md mx-auto pt-8">
          <div className="space-y-2 text-center">
            <h1 className="text-3xl font-bold">{t("resendVerificationEmailPage.title")}</h1>
            <p className="text-txt-tertiary">
              {t("resendVerificationEmailPage.description")}
            </p>
          </div>

          <form onSubmit={e => void onSubmit(e)} className="space-y-4">
            <Field
              label={t("resendVerificationEmailPage.fields.email")}
              type="email"
              placeholder={t("resendVerificationEmailPage.fields.emailPlaceholder")}
              {...register("email")}
              required
              error={formState.errors.email?.message}
            />

            <Button
              type="submit"
              className="w-xs h-10 mx-auto mt-6"
              disabled={formState.isSubmitting}
            >
              {formState.isSubmitting
                ? t("resendVerificationEmailPage.actions.sendingVerification")
                : t("resendVerificationEmailPage.actions.sendVerification")}
            </Button>
          </form>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {t("resendVerificationEmailPage.alreadyVerified")}
              {" "}
              <Link
                to="/auth/login"
                className="underline text-txt-primary hover:text-txt-secondary"
              >
                {t("resendVerificationEmailPage.actions.backToLogin")}
              </Link>
            </p>
          </div>
        </div>
      );
}
