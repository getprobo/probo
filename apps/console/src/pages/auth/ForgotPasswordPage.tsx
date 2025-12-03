import { Link } from "react-router";
import { Button, Field, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { usePageTitle } from "@probo/hooks";
import { useState } from "react";

const schema = z.object({
  email: z.string().email(),
});

export default function ForgotPasswordPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [isSubmitted, setIsSubmitted] = useState(false);
  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      email: "",
    },
  });

  const onSubmit = handleSubmit(async (data) => {
    const response = await fetch("/connect/forget-password", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      toast({
        title: __("Request failed"),
        description:
          errorData.message || __("Failed to send reset instructions"),
        variant: "error",
      });
      return;
    }

    toast({
      title: __("Success"),
      description: __("Password reset instructions sent to your email"),
      variant: "success",
    });
    setIsSubmitted(true);
  });

  usePageTitle(__("Forgot Password"));

  if (isSubmitted) {
    return (
      <div className="space-y-6 w-full max-w-md mx-auto">
        <div className="space-y-2 text-center">
          <h1 className="text-3xl font-bold">{__("Check your email")}</h1>
          <p className="text-txt-tertiary">
            {__("We've sent password reset instructions to your email address")}
          </p>
        </div>

        <div className="text-center">
          <p className="text-sm text-txt-tertiary">
            {__("Didn't receive the email?")}{" "}
            <button
              onClick={() => setIsSubmitted(false)}
              className="underline text-txt-primary hover:text-txt-secondary"
            >
              {__("Try again")}
            </button>
          </p>
        </div>

        <div className="text-center">
          <p className="text-sm text-txt-tertiary">
            {__("Remember your password?")}{" "}
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

  return (
    <div className="space-y-6 w-full max-w-md mx-auto">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Forgot Password")}</h1>
        <p className="text-txt-tertiary">
          {__(
            "Enter your email address and we'll send you instructions to reset your password"
          )}
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-4">
        <Field
          label={__("Email")}
          type="email"
          placeholder={__("name@example.com")}
          {...register("email")}
          required
          error={formState.errors.email?.message}
        />

        <Button type="submit" className="w-full" disabled={formState.isLoading}>
          {formState.isLoading
            ? __("Sending instructions...")
            : __("Send reset instructions")}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-sm text-txt-tertiary">
          {__("Remember your password?")}{" "}
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
