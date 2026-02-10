import { UnAuthenticatedError } from "@probo/relay";
import { useEffect } from "react";
import { useMutation } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { useAssumeMutation } from "#/__generated__/iam/useAssumeMutation.graphql";

import { useOrganizationId } from "../useOrganizationId";

interface UseAssumeParameters {
  afterAssumePath?: string;
  onSuccess: () => void;
}

const assumeMutation = graphql`
  mutation useAssumeMutation(
    $input: AssumeOrganizationSessionInput!
  ) {
    assumeOrganizationSession(input: $input) {
      result {
        __typename
        ... on OrganizationSessionCreated {
          membership {
            id
            lastSession {
              id
              expiresAt
            }
          }
        }
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

export function useAssume(params: UseAssumeParameters) {
  const { afterAssumePath, onSuccess } = params;

  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const [assumeOrganizationSession] = useMutation<useAssumeMutation>(assumeMutation);

  useEffect(() => {
    assumeOrganizationSession({
      variables: {
        input: { organizationId },
      },
      onError: (error) => {
        if (error instanceof UnAuthenticatedError) {
          const search = new URLSearchParams([
            ["organization-id", organizationId],
            ["redirect-path", afterAssumePath ?? window.location.pathname + window.location.search],
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

        switch (result.__typename) {
          case "PasswordRequired":
            search.set("organization-id", organizationId);
            search.set("redirect-path", afterAssumePath ?? window.location.pathname + window.location.search);

            void navigate({ pathname: "/auth/passord-login", search: "?" + search.toString() });
            break;
          case "SAMLAuthenticationRequired":
            window.location.href = result.redirectUrl;
            break;
          default:
            onSuccess();
        }
      },
    });
  }, [afterAssumePath, organizationId, onSuccess, navigate, assumeOrganizationSession]);

  return;
}
