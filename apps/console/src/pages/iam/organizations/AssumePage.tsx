import { useTranslate } from "@probo/i18n";
import { UnAuthenticatedError } from "@probo/relay";
import { useEffect } from "react";
import { useMutation } from "react-relay";
import { useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { AssumePageMutation } from "#/__generated__/iam/AssumePageMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import AuthLayout from "../auth/AuthLayout";

const assumeMutation = graphql`
  mutation AssumePageMutation(
    $input: AssumeOrganizationSessionInput!
  ) {
    assumeOrganizationSession(input: $input) {
      result {
        __typename
        ... on PasswordRequired {
          reason
        }
        ... on SAMLAuthenticationRequired {
          reason
          redirectUrl
        }
      }
    }
  }
`;

function AssumePageInner() {
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { __ } = useTranslate();

  const [assumeOrganizationSession] = useMutation<AssumePageMutation>(assumeMutation);

  const continueUrlParam = searchParams.get("continue");
  let safeContinueUrl: string;
  if (continueUrlParam) {
    const continueUrl = new URL(continueUrlParam);
    safeContinueUrl = window.location.origin + continueUrl.pathname + continueUrl.search;
  } else {
    safeContinueUrl = window.location.origin + `/organizations/${organizationId}`;
  }

  useEffect(() => {
    assumeOrganizationSession({
      variables: {
        input: { organizationId, continue: safeContinueUrl },
      },
      onError: (error) => {
        if (error instanceof UnAuthenticatedError) {
          const search = new URLSearchParams([
            ["organization-id", organizationId],
            ["continue", safeContinueUrl],
          ]);

          void navigate({ pathname: "/auth/login", search: "?" + search.toString() });
          return;
        }
      },
      onCompleted: ({ assumeOrganizationSession }) => {
        if (!assumeOrganizationSession) {
          throw new Error("complete mutation result is empty");
        }

        const { result } = assumeOrganizationSession;
        const search = new URLSearchParams();
        let samlSSOLoginURL: URL;

        switch (result.__typename) {
          case "PasswordRequired":
            search.set("organization-id", organizationId);
            search.set("continue", safeContinueUrl);

            void navigate({ pathname: "/auth/password-login", search: "?" + search.toString() });
            break;
          case "SAMLAuthenticationRequired":
            samlSSOLoginURL = new URL(result.redirectUrl);
            samlSSOLoginURL.search = "?" + searchParams.toString();

            window.location.href = samlSSOLoginURL.toString();
            break;
          default:
            window.location.href = safeContinueUrl;
        }
      },
    });
  }, [organizationId, navigate, assumeOrganizationSession, safeContinueUrl, searchParams]);

  return (
    <AuthLayout>
      <div className="space-y-6 w-full max-w-md mx-auto pt-8">
        <div className="space-y-2 text-center">
          <h1 className="text-3xl font-bold">{__("Sign in Redirection")}</h1>
          <p className="text-txt-tertiary">
            {__("Redirecting you to your authentication URL…")}
          </p>
        </div>
      </div>
    </AuthLayout>
  );
}

export default function AssumePage() {
  return (
    <IAMRelayProvider>
      <AssumePageInner />
    </IAMRelayProvider>
  );
}
