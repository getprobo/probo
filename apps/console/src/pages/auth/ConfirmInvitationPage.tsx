import { Link, useNavigate } from "react-router";
import { Button, Field, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { usePageTitle } from "@probo/hooks";
import { buildEndpoint } from "/providers/RelayProviders";

const schema = z.object({
  password: z.string().min(8),
});

export default function ConfirmInvitationPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { register, handleSubmit, formState } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        password: "",
      },
    }
  );

  const onSubmit = handleSubmit(async (data) => {
    const searchParams = new URLSearchParams(location.search);
    const token = searchParams.get("token");

    if (!token) {
      toast({
        title: __("Confirmation failed"),
        description: __("Invalid or missing invitation token"),
        variant: "error",
      });
      return;
    }

    const response = await fetch(
      buildEndpoint("/api/console/v1/auth/invitation"),
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          token: token,
          password: data.password,
        }),
      }
    );

    // Registration failed
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      toast({
        title: __("Confirmation failed"),
        description: errorData.message || __("Confirmation failed"),
        variant: "error",
      });
      return;
    }

    toast({
      title: __("Success"),
      description: __("Invitation confirmed successfully"),
      variant: "success",
    });
    navigate("/", { replace: true });
  });

  usePageTitle(__("Confirm invitation"));

  return (
    <div className="space-y-6 w-full max-w-md mx-auto">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Confirm invitation")}</h1>
        <p className="text-txt-tertiary">
          {__("Enter your information to confirm your invitation")}
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-4">
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
            ? __("Confirming invitation...")
            : __("Confirm invitation")}
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
