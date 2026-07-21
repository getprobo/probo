// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { useCallback, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const submittedRef = useRef<boolean>(false);
  const safeContinueUrl = useSafeContinueUrl();

  usePageTitle(t("activateAccountPage.pageTitle"));

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
            title: t("activateAccountPage.errors.activationFailed"),
            description: formatError(t("activateAccountPage.errors.activationFailed"), errors),
            variant: "error",
          });

          return;
        }

        toast({
          title: t("common.success"),
          description: t("activateAccountPage.messages.activated"),
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
          pathname: "/auth/login",
          search: "?" + search.toString(),
        }, { replace: true });
      },
      onError: (e) => {
        toast({
          title: t("activateAccountPage.errors.activationFailed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  }, [t, toast, activateAccount, navigate, safeContinueUrl]);

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
        <h1 className="text-3xl font-bold">{t("activateAccountPage.title")}</h1>
        <p className="text-txt-tertiary">
          {t("activateAccountPage.activating")}
        </p>
      </div>
      <div className="text-center mt-6 text-sm text-txt-secondary">
        <Link
          to="/auth/login"
          className="underline hover:text-txt-primary"
        >
          {t("activateAccountPage.actions.goBack")}
        </Link>
      </div>
    </div>
  );
}
