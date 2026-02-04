import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useMutation } from "react-relay";
import { Link, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { ResetPasswordPageMutation } from "#/__generated__/iam/ResetPasswordPageMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const resetPasswordMutation = graphql`
  mutation ResetPasswordPageMutation($input: ResetPasswordInput!) {
    resetPassword(input: $input) {
      success
    }
  }
`;

const schema = z
  .object({
    password: z.string().min(8),
    confirmPassword: z.string().min(8),
  })
  .refine(data => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

export default function ResetPasswordPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");

  usePageTitle(__("Reset password"));

  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      password: "",
      confirmPassword: "",
    },
  });

  const [resetPassword] = useMutation<ResetPasswordPageMutation>(
    resetPasswordMutation,
  );

  const onSubmit = handleSubmit((data) => {
    if (!token) {
      toast({
        title: __("Reset failed"),
        description: __("Invalid or missing reset token"),
        variant: "error",
      });
      return;
    }

    resetPassword({
      variables: {
        input: {
          password: data.password,
          token,
        },
      },
      onError: (e: Error) => {
        toast({
          title: __("Reset failed"),
          description: e.message,
          variant: "error",
        });
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Reset failed"),
            description: formatError(
              __("Password reset failed"),
              e as GraphQLError,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Password reset successfully"),
          variant: "success",
        });
        void navigate("/auth/login", { replace: true });
      },
    });
  });

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Reset password")}</h1>
        <p className="text-txt-tertiary">
          {__("Enter your new password to reset your account")}
        </p>
      </div>

      <form onSubmit={e => void onSubmit(e)} className="space-y-4">
        <Field
          label={__("New Password")}
          type="password"
          placeholder="••••••••"
          {...register("password")}
          required
          error={formState.errors.password?.message}
        />

        <Field
          label={__("Confirm Password")}
          type="password"
          placeholder="••••••••"
          {...register("confirmPassword")}
          required
          error={formState.errors.confirmPassword?.message}
        />

        <Button type="submit" className="w-xs h-10 mx-auto mt-6" disabled={formState.isLoading}>
          {formState.isLoading
            ? __("Resetting password...")
            : __("Reset password")}
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
            {__("Log in here")}
          </Link>
        </p>
      </div>
    </div>
  );
}
