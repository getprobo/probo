import { Link, useNavigate } from "react-router";
import { Button, Field, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { usePageTitle } from "@probo/hooks";
import { buildEndpoint } from "/providers/RelayProviders";

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
  fullName: z.string().min(2),
});

export default function RegisterPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      email: "",
      password: "",
      fullName: "",
    },
  });

  const onSubmit = handleSubmit(async (data) => {
    const response = await fetch(
      buildEndpoint("/api/console/v1/auth/register"),
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(data),
      },
    );

    // Registration failed
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      toast({
        title: __("Registration failed"),
        description: errorData.message || __("Registration failed"),
        variant: "error",
      });
      return;
    }

    toast({
      title: __("Success"),
      description: __("Account created successfully"),
      variant: "success",
    });
    navigate("/", { replace: true });
  });

  usePageTitle(__("Sign up"));

  return (
    <div className="space-y-6 w-full max-w-md mx-auto">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Sign up")}</h1>
        <p className="text-txt-tertiary">
          {__("Enter your information to create an account")}
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-4">
        <Field
          label={__("Full Name")}
          type="text"
          placeholder={__("John Doe")}
          {...register("fullName")}
          required
          error={formState.errors.fullName?.message}
        />

        <Field
          label={__("Email")}
          type="email"
          placeholder={__("name@example.com")}
          {...register("email")}
          required
          error={formState.errors.email?.message}
        />

        <Field
          label={__("Password")}
          type="password"
          placeholder="••••••••"
          {...register("password")}
          required
          error={formState.errors.password?.message}
        />

        <Button type="submit" className="w-full" disabled={formState.isLoading}>
          {formState.isLoading
            ? __("Creating account...")
            : __("Sign up with email")}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-sm text-txt-tertiary">
          {__("Already have an account?")}{" "}
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
