import { useState, useEffect } from "react";
import { useLocation, useNavigate, Link } from "react-router";
import { Button, Field, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import type { PayloadError } from "relay-runtime";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { usePageTitle } from "@probo/hooks";
import type { ConfirmEmailPageMutation } from "./__generated__/ConfirmEmailPageMutation.graphql";

const ConfirmEmailMutation = graphql`
  mutation ConfirmEmailPageMutation($input: ConfirmEmailInput!) {
    confirmEmail(input: $input) {
      success
    }
  }
`;

const confirmEmailSchema = z.object({
  token: z.string().min(1, "Please enter a confirmation token"),
});

export default function ConfirmEmailPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [isConfirmed, setIsConfirmed] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [commitMutation] =
    useMutation<ConfirmEmailPageMutation>(ConfirmEmailMutation);

  const form = useFormWithSchema(confirmEmailSchema, {
    defaultValues: {
      token: "",
    },
  });

  usePageTitle(__("Confirm Email"));

  useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const urlToken = searchParams.get("token");

    if (urlToken) {
      form.setValue("token", urlToken);
    }
  }, [location.search, form]);

  const handleSubmit = form.handleSubmit(async (data) => {
    setIsLoading(true);

    try {
      commitMutation({
        variables: {
          input: {
            token: data.token.trim(),
          },
        },
        onCompleted: (_response, errors: PayloadError[] | null) => {
          if (errors) {
            toast({
              title: __("Error"),
              description: errors[0]?.message || __("Failed to confirm email"),
              variant: "error",
            });
            setIsLoading(false);
            return;
          }

          setIsConfirmed(true);
          toast({
            title: __("Success"),
            description: __("Your email has been confirmed successfully"),
            variant: "success",
          });
          setIsLoading(false);
        },
        onError: (err) => {
          toast({
            title: __("Error"),
            description: err.message || __("Failed to confirm email. Please try again."),
            variant: "error",
          });
          setIsLoading(false);
        },
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: error instanceof Error ? error.message : __("Failed to confirm email. Please try again."),
        variant: "error",
      });
      setIsLoading(false);
    }
  });

  return (
    <div className="space-y-6 w-full max-w-md mx-auto">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Email Confirmation")}</h1>
        <p className="text-txt-tertiary">
          {__("Confirm your email address to complete registration")}
        </p>
      </div>

      {isConfirmed ? (
        <div className="space-y-4 text-center">
          <p className="text-green-600 dark:text-green-400">
            {__("Your email has been confirmed successfully!")}
          </p>
          <Button onClick={() => navigate("/auth/login")} className="w-full">
            {__("Proceed to Login")}
          </Button>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4">
          <Field
            label={__("Confirmation Token")}
            type="text"
            placeholder={__("Enter your confirmation token")}
            {...form.register("token")}
            error={form.formState.errors.token?.message}
            disabled={isLoading}
            help={__("The token has been automatically filled from the URL if available")}
          />

          <Button
            type="submit"
            className="w-full"
            disabled={isLoading}
          >
            {isLoading ? __("Confirming...") : __("Confirm Email")}
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
