// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useState } from "react";
import { useMutation } from "react-relay";
import { Link, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { ResendVerificationEmailPageMutation } from "#/__generated__/iam/ResendVerificationEmailPageMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const resendVerificationEmailMutation = graphql`
  mutation ResendVerificationEmailPageMutation(
    $input: ResendVerificationEmailInput!
  ) {
    resendVerificationEmail(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
});

export default function ResendVerificationEmailPage() {
  const { toast } = useToast();
  const { __ } = useTranslate();
  const [searchParams] = useSearchParams();

  usePageTitle(__("Verify Email"));

  const [emailSent, setEmailSent] = useState<boolean>(false);
  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      email: searchParams.get("email") ?? "",
    },
  });

  const [resendVerificationEmail]
    = useMutation<ResendVerificationEmailPageMutation>(
      resendVerificationEmailMutation,
    );

  const onSubmit = handleSubmit(({ email }) => {
    resendVerificationEmail({
      variables: {
        input: { email },
      },
      onError: (e: Error) => {
        toast({
          title: __("Request failed"),
          description: e.message,
          variant: "error",
        });
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Request failed"),
            description: formatError(
              __("Failed to send verification email"),
              e,
            ),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Verification email sent"),
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
            <h1 className="text-3xl font-bold">{__("Check your email")}</h1>
            <p className="text-txt-tertiary">
              {__(
                "We've sent a verification link to your email address. Please check your inbox and click the link to verify your account.",
              )}
            </p>
          </div>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {__("Didn't receive the email?")}
              {" "}
              <button
                onClick={() => setEmailSent(false)}
                className="underline text-txt-primary hover:text-txt-secondary"
              >
                {__("Try again")}
              </button>
            </p>
          </div>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {__("Already verified?")}
              {" "}
              <Link
                to="/auth/login"
                className="underline text-txt-primary hover:text-txt-secondary"
              >
                {__("Back to login")}
              </Link>
            </p>
          </div>
        </div>
      )
    : (
        <div className="space-y-6 w-full max-w-md mx-auto pt-8">
          <div className="space-y-2 text-center">
            <h1 className="text-3xl font-bold">
              {__("Verify your email")}
            </h1>
            <p className="text-txt-tertiary">
              {__(
                "Your email address has not been verified yet. Enter your email below and we'll send you a verification link.",
              )}
            </p>
          </div>

          <form onSubmit={e => void onSubmit(e)} className="space-y-4">
            <Field
              label={__("Email")}
              type="email"
              placeholder={__("name@example.com")}
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
                ? __("Sending...")
                : __("Send verification email")}
            </Button>
          </form>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {__("Already verified?")}
              {" "}
              <Link
                to="/auth/login"
                className="underline text-txt-primary hover:text-txt-secondary"
              >
                {__("Back to login")}
              </Link>
            </p>
          </div>
        </div>
      );
}
