import { Link, useNavigate, useSearchParams } from "react-router";
import { Button, Field, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { usePageTitle } from "@probo/hooks";
import { buildEndpoint } from "/providers/RelayProviders";
import { useEffect } from "react";

const schema = z.object({
  fullName: z.string().min(2),
  password: z.string().min(8),
});

export default function SignupFromInvitationPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();

  const { register, handleSubmit, formState, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        fullName: "",
        password: "",
      },
    }
  );

  useEffect(() => {
    const fullNameFromParams = searchParams.get("fullName") || "";
    if (fullNameFromParams) {
      reset({
        fullName: fullNameFromParams,
        password: "",
      });
    }
  }, [searchParams, reset]);

  const onSubmit = handleSubmit(async (data) => {
    const token = searchParams.get("token");

    if (!token) {
      toast({
        title: __("Signup failed"),
        description: __("Invalid or missing invitation token"),
        variant: "error",
      });
      return;
    }

    const response = await fetch(
      buildEndpoint("/auth/signup-from-invitation"),
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          token: token,
          password: data.password,
          fullName: data.fullName,
        }),
      }
    );

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      toast({
        title: __("Signup failed"),
        description: errorData.message || __("Signup failed"),
        variant: "error",
      });
      return;
    }

    toast({
      title: __("Success"),
      description: __("Account created successfully. Please accept your invitation to join the organization."),
      variant: "success",
    });
    navigate("/", { replace: true });
  });

  usePageTitle(__("Create your account"));

  return (
    <div className="space-y-6 w-full max-w-md mx-auto">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Create your account")}</h1>
        <p className="text-txt-tertiary">
          {__("Set your password to join the organization")}
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
            : __("Create account")}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-sm text-txt-tertiary">
          {__("Already have an account?")}{" "}
          <Link
            to="/authentication/login"
            className="underline text-txt-primary hover:text-txt-secondary"
          >
            {__("Log in here")}
          </Link>
        </p>
      </div>
    </div>
  );
}
