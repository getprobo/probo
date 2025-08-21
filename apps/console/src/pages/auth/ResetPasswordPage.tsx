import { Link, useNavigate } from "react-router";
import { Button, Field, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { usePageTitle } from "@probo/hooks";
import { buildEndpoint } from "/providers/RelayProviders";
import { useEffect } from "react";

const schema = z
  .object({
    authToken: z.string(),
    password: z.string().min(8),
    confirmPassword: z.string().min(8),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

export default function ResetPasswordPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { register, handleSubmit, formState, setValue } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        authToken: "",
        password: "",
        confirmPassword: "",
      },
    }
  );

  const onSubmit = handleSubmit(async (data) => {
    const response = await fetch(
      buildEndpoint("/api/console/v1/auth/reset-password"),
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          authToken: data.authToken,
          password: data.password,
        }),
      }
    );

    // Reset failed
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      toast({
        title: __("Reset failed"),
        description: errorData.message || __("Password reset failed"),
        variant: "error",
      });
      return;
    }

    toast({
      title: __("Success"),
      description: __("Password reset successfully"),
      variant: "success",
    });
    navigate("/auth/login", { replace: true });
  });

  usePageTitle(__("Reset password"));

  useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const urlToken = searchParams.get("authToken");

    if (urlToken) {
      setValue("authToken", urlToken);
    }
  }, [location.search, setValue]);

  return (
    <div className="space-y-6 w-full max-w-md mx-auto">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Reset password")}</h1>
        <p className="text-txt-tertiary">
          {__("Enter your new password to reset your account")}
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-4">
        <Field
          label={__("Token")}
          type="text"
          hidden
          placeholder={__("Enter your token")}
          {...register("authToken")}
          required
          error={formState.errors.authToken?.message}
        />

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

        <Button type="submit" className="w-full" disabled={formState.isLoading}>
          {formState.isLoading
            ? __("Resetting password...")
            : __("Reset password")}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-sm text-txt-tertiary">
          {__("Remember your password?")}{" "}
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
