import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useCallback, useEffect, useRef } from "react";
import { useMutation } from "react-relay";
import { Link, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { ActivateAccountPageMutation } from "#/__generated__/iam/ActivateAccountPageMutation.graphql";

const activateAccountMutation = graphql`
  mutation ActivateAccountPageMutation(
    $input: ActivateAccountInput!
  ) {
    activateAccount(input: $input) {
      profile {
        id
      }
    }
  }
`;

export default function ActivateAccountPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const submittedRef = useRef<boolean>(false);

  usePageTitle(__("Activate Account"));

  const [activateAccount] = useMutation<ActivateAccountPageMutation>(activateAccountMutation);

  const handleActivateAccount = useCallback((token: string) => {
    if (submittedRef.current) return;

    activateAccount({
      variables: {
        input: { token },
      },
      onCompleted: (_, errors: GraphQLError[] | null) => {
        if (errors) {
          for (const err of errors) {
            if (err.extensions?.code === "ALREADY_AUTHENTICATED") {
              window.location.href = "/";
              return;
            }
          }
          toast({
            title: __("Activation failed"),
            description: formatError(__("Activation failed"), errors),
            variant: "error",
          });

          return;
        }

        toast({
          title: __("Success"),
          description: __(
            "Account activated successfully.",
          ),
          variant: "success",
        });
        void navigate("/", { replace: true });
      },
      onError: (e) => {
        toast({
          title: __("Activation failed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  }, [__, toast, activateAccount, navigate]);

  useEffect(() => {
    const token = searchParams.get("token");
    if (!submittedRef.current && token) {
      void handleActivateAccount(token.trim());
      submittedRef.current = true;
    }
  }, [handleActivateAccount, searchParams]);

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Account Activation")}</h1>
        <p className="text-txt-tertiary">
          {__("Activating your account…")}
        </p>
      </div>
      <div className="text-center mt-6 text-sm text-txt-secondary">
        <Link
          to="/auth/login"
          className="underline hover:text-txt-primary"
        >
          {__("Go back")}
        </Link>
      </div>
    </div>
  );
}
