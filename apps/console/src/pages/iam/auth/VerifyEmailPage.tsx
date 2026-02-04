import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useState } from "react";
import { useMutation } from "react-relay";
import { Link, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { VerifyEmailPageMutation } from "#/__generated__/iam/VerifyEmailPageMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const verifyEmailMutation = graphql`
  mutation VerifyEmailPageMutation($input: VerifyEmailInput!) {
    verifyEmail(input: $input) {
      success
    }
  }
`;

const confirmEmailSchema = z.object({
  token: z.string().min(1, "Please enter a confirmation token"),
});

export default function VerifyEmailPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();

  usePageTitle(__("Confirm Email"));

  const [isConfirmed, setIsConfirmed] = useState(false);

  const form = useFormWithSchema(confirmEmailSchema, {
    defaultValues: {
      token: searchParams.get("token") ?? "",
    },
  });

  const [verifyEmail]
    = useMutation<VerifyEmailPageMutation>(verifyEmailMutation);

  const handleSubmit = form.handleSubmit((data) => {
    verifyEmail({
      variables: {
        input: {
          token: data.token.trim(),
        },
      },
      onCompleted: (_, errors) => {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to confirm email"), errors),
            variant: "error",
          });
          return;
        }

        setIsConfirmed(true);
        toast({
          title: __("Success"),
          description: __("Your email has been confirmed successfully"),
          variant: "success",
        });
      },
      onError: (err) => {
        toast({
          title: __("Error"),
          description: err.message || __("Failed to confirm email"),
          variant: "error",
        });
      },
    });
  });

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Email Confirmation")}</h1>
        <p className="text-txt-tertiary">
          {__("Confirm your email address to complete registration")}
        </p>
      </div>

      {isConfirmed
        ? (
            <div className="space-y-4 text-center">
              <p className="text-green-600 dark:text-green-400">
                {__("Your email has been confirmed successfully!")}
              </p>
              <Button to="/auth/login" className="w-full">
                {__("Proceed to Login")}
              </Button>
            </div>
          )
        : (
            <form onSubmit={e => void handleSubmit(e)} className="space-y-4">
              <Field
                label={__("Confirmation Token")}
                type="text"
                placeholder={__("Enter your confirmation token")}
                {...form.register("token")}
                error={form.formState.errors.token?.message}
                disabled={form.formState.isSubmitting}
                help={__(
                  "The token has been automatically filled from the URL if available",
                )}
              />

              <Button
                type="submit"
                className="w-xs h-10 mx-auto mt-6"
                disabled={form.formState.isSubmitting}
              >
                {form.formState.isSubmitting
                  ? __("Confirming...")
                  : __("Confirm Email")}
              </Button>
            </form>
          )}

      <div className="text-center">
        {!isConfirmed && (
          <p className="text-sm text-txt-tertiary">
            <Link
              to="/auth/login"
              className="underline text-txt-primary hover:text-txt-secondary"
            >
              {__("Back to Login")}
            </Link>
          </p>
        )}
      </div>
    </div>
  );
}
