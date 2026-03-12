import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useCallback, useEffect, useRef } from "react";
import { useMutation } from "react-relay";
import { Link, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { ActivateAccountPageMutation$data, ActivateAccountPageMutation } from "#/__generated__/iam/ActivateAccountPageMutation.graphql";
import { useSafeContinueUrl } from "#/hooks/useSafeContinueUrl";

const activateAccountMutation = graphql`
  mutation ActivateAccountPageMutation(
    $input: ActivateAccountInput!
  ) {
    activateAccount(input: $input) {
      createPasswordToken
      ssoLoginUrl
    }
  }
`;

export default function ActivateAccountPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const submittedRef = useRef<boolean>(false);
  const safeContinueUrl = useSafeContinueUrl();

  usePageTitle(__("Activate Account"));

  const [activateAccount] = useMutation<ActivateAccountPageMutation>(activateAccountMutation);

  const handleActivateAccount = useCallback((token: string) => {
    if (submittedRef.current) return;

    activateAccount({
      variables: {
        input: { token },
      },
      onCompleted: (response: ActivateAccountPageMutation$data, errors: GraphQLError[] | null) => {
        if (errors) {
          for (const err of errors) {
            if (err.extensions?.code === "ACCOUNT_ALREADY_ACTIVATED") {
              void navigate({
                pathname: safeContinueUrl.pathname,
                search: safeContinueUrl.search,
              }, { replace: true });
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

        const { activateAccount } = response;

        if (!activateAccount) {
          throw new Error("mutation data missing");
        }

        if (activateAccount.ssoLoginUrl) {
          const url = new URL(activateAccount.ssoLoginUrl);
          url.searchParams.set("continue", safeContinueUrl.toString());

          window.location.href = url.toString();
          return;
        }

        if (activateAccount.createPasswordToken) {
          const search = new URLSearchParams([
            ["token", activateAccount.createPasswordToken],
            ["continue", safeContinueUrl.toString()],
          ]);
          void navigate(
            {
              pathname: "/auth/create-password",
              search: "?" + search.toString(),
            },
            { replace: true },
          );
          return;
        }

        const search = new URLSearchParams([["continue", safeContinueUrl.toString()]]);
        void navigate({
          pathname: "/auth/password-login",
          search: "?" + search.toString(),
        }, { replace: true });
      },
      onError: (e) => {
        toast({
          title: __("Activation failed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  }, [__, toast, activateAccount, navigate, safeContinueUrl, searchParams]);

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
