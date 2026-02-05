import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useCallback, useEffect, useRef } from "react";
import { useMutation } from "react-relay";
import { Link, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import { getPathPrefix } from "#/utils/pathPrefix";

import type { VerifyMagicLinkPageMutation } from "./__generated__/VerifyMagicLinkPageMutation.graphql";

const verifyMagicLinkMutation = graphql`
  mutation VerifyMagicLinkPageMutation($input: VerifyMagicLinkInput!) {
    verifyMagicLink(input: $input) {
      success
    }
  }
`;

export default function VerifyMagicLinkPagePageMutation() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const submittedRef = useRef<boolean>(false);

  usePageTitle(__("Verify Magic Link"));

  const [verifyMagicLink] = useMutation<VerifyMagicLinkPageMutation>(
    verifyMagicLinkMutation,
  );

  const handleVerifyMagicToken = useCallback((token: string) => {
    if (submittedRef.current) return;

    verifyMagicLink({
      variables: {
        input: { token },
      },
      onCompleted: (_, errors: GraphQLError[] | null) => {
        if (errors) {
          for (const err of errors) {
            if (err.extensions?.code === "ALREADY_AUTHENTICATED") {
              const pathPrefix = getPathPrefix();
              window.location.href = pathPrefix ? getPathPrefix() : "/";
              return;
            }
          }
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
        const pathPrefix = getPathPrefix();
        window.location.href = pathPrefix ? getPathPrefix() : "/";
      },
      onError: (err) => {
        toast({
          title: __("Error"),
          description: err.message,
          variant: "error",
        });
      },
    });
  }, [__, toast, verifyMagicLink]);

  useEffect(() => {
    const token = searchParams.get("token");
    if (!submittedRef.current && token) {
      void handleVerifyMagicToken(token.trim());
      submittedRef.current = true;
    }
  }, [handleVerifyMagicToken, searchParams]);

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Email Confirmation")}</h1>
        <p className="text-txt-tertiary">
          {__("Confirming your email address to complete registration…")}
        </p>
      </div>
      <div className="text-center mt-6 text-sm text-txt-secondary">
        <Link
          to="/connect"
          className="underline hover:text-txt-primary"
        >
          {__("Go back")}
        </Link>
      </div>
    </div>
  );
}
