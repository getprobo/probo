import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useSearchParams } from "react-router";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import z from "zod";
import { useEffect, useRef } from "react";
import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { formatError } from "@probo/helpers";
import type { VerifyMagicLinkPageMutation } from "./__generated__/VerifyMagicLinkPageMutation.graphql";
import { getPathPrefix } from "/utils/pathPrefix";

const verifyMagicLinkMutation = graphql`
  mutation VerifyMagicLinkPageMutation($input: VerifyMagicLinkInput!) {
    verifyMagicLink(input: $input) {
      success
    }
  }
`;

const verifyMagicLinkSchema = z.object({
  token: z.string().min(1, "Please enter a magic token"),
});

export default function VerifyMagicLinkPagePageMutation() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const submittedRef = useRef<boolean>(false);

  usePageTitle(__("Verify Magic Link"));

  const form = useFormWithSchema(verifyMagicLinkSchema, {
    defaultValues: {
      token: searchParams.get("token") ?? "",
    },
  });

  const [verifyMagicLink] = useMutation<VerifyMagicLinkPageMutation>(
    verifyMagicLinkMutation,
  );

  const handleSubmit = form.handleSubmit(async (data) => {
    verifyMagicLink({
      variables: {
        input: {
          token: data.token.trim(),
        },
      },
      onCompleted: (_, errors) => {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to connect"), errors),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Your have successfully signed in"),
          variant: "success",
        });
        window.location.href = getPathPrefix();
      },
      onError: (err) => {
        toast({
          title: __("Error"),
          description: err.message,
          variant: "error",
        });
      },
    });
  });

  useEffect(() => {
    if (!submittedRef.current && searchParams.get("token")) {
      handleSubmit();
      submittedRef.current = true;
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

      <form onSubmit={handleSubmit} className="space-y-4">
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
          className="w-full"
          disabled={form.formState.isSubmitting}
        >
          {form.formState.isSubmitting
            ? __("Confirming...")
            : __("Confirm Email")}
        </Button>
      </form>
    </div>
  );
}
