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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { useNavigate, useSearchParams } from "react-router";
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
  const submittedRef = useRef(false);
  const safeContinueUrl = useSafeContinueUrl();
  const token = searchParams.get("token")?.trim() ?? "";

  usePageTitle(t("activateAccountPage.pageTitle"));

  const [activateAccount, isActivating] = useMutation<ActivateAccountPageMutation>(activateAccountMutation);

  const handleActivateAccount = useCallback(() => {
    if (submittedRef.current) {
      return;
    }

    if (token === "") {
      toast({
        title: t("activateAccountPage.errors.activationFailed"),
        description: t("activateAccountPage.errors.missingToken"),
        variant: "error",
      });
      return;
    }

    submittedRef.current = true;

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

          submittedRef.current = false;
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
        submittedRef.current = false;
        toast({
          title: t("activateAccountPage.errors.activationFailed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  }, [t, toast, activateAccount, navigate, safeContinueUrl, token]);

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("activateAccountPage.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("activateAccountPage.description")}
        </Text>
      </div>
      <Button
        type="button"
        variant="solid"
        color="neutral"
        highContrast
        size={3}
        className="w-full"
        loading={isActivating}
        onClick={handleActivateAccount}
      >
        {t("activateAccountPage.actions.continue")}
      </Button>
      <Text align="center" size={2} className="block">
        <Link to="/auth/login">
          {t("activateAccountPage.actions.goBack")}
        </Link>
      </Text>
    </div>
  );
}
