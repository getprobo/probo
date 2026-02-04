import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useState } from "react";
import { useMutation } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { ForgotPasswordPageMutation } from "#/__generated__/iam/ForgotPasswordPageMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const sendInstructionsMutation = graphql`
  mutation ForgotPasswordPageMutation($input: ForgotPasswordInput!) {
    forgotPassword(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
});

export default function ForgotPasswordPage() {
  const { toast } = useToast();
  const { __ } = useTranslate();

  usePageTitle(__("Forgot Password"));

  const [instructionsSent, setInstructionsSent] = useState<boolean>();
  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      email: "",
    },
  });

  const [sendInstructions] = useMutation<ForgotPasswordPageMutation>(
    sendInstructionsMutation,
  );

  const onSubmit = handleSubmit(({ email }) => {
    sendInstructions({
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
              __("Failed to send reset instructions"),
              e,
            ),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Password reset instructions sent to your email"),
          variant: "success",
        });
        setInstructionsSent(true);
      },
    });
  });

  return instructionsSent
    ? (
        <div className="space-y-6 w-full max-w-md mx-auto pt-8">
          <div className="space-y-2 text-center">
            <h1 className="text-3xl font-bold">{__("Check your email")}</h1>
            <p className="text-txt-tertiary">
              {__("We've sent password reset instructions to your email address")}
            </p>
          </div>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {__("Didn't receive the email?")}
              {" "}
              <button
                onClick={() => setInstructionsSent(false)}
                className="underline text-txt-primary hover:text-txt-secondary"
              >
                {__("Try again")}
              </button>
            </p>
          </div>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {__("Remember your password?")}
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
            <h1 className="text-3xl font-bold">{__("Forgot Password")}</h1>
            <p className="text-txt-tertiary">
              {__(
                "Enter your email address and we'll send you instructions to reset your password",
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
                ? __("Sending instructions...")
                : __("Send reset instructions")}
            </Button>
          </form>

          <div className="text-center">
            <p className="text-sm text-txt-tertiary">
              {__("Remember your password?")}
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
